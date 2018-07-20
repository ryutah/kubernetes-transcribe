package conversion

import (
	"fmt"
	"reflect"
)

type typePair struct {
	source reflect.Type
	dest   reflect.Type
}

// DebugLogger allows you to get debugging messages if necessary.
type DebugLogger interface {
	Logf(format string, args ...interface{})
}

// Converter knows how to convert one type to another.
type Converter struct {
	// map from the conversion pair to a function which can.
	funcs map[typePair]reflect.Value

	// If non-nil, will be called to print helpful debugging info. Quite verbose.
	Debug DebugLogger

	// NameFunc is called to retrive the name of a type; this name is used for the
	// purpose of deciding whether two types match or not (i.e., will we attempt to
	// do a conversion). The default returns the go type name.
	NameFunc func(t reflect.Type) string
}

// NewConverter creates a new Converter object
func NewConverter() *Converter {
	return &Converter{
		funcs:    map[typePair]reflect.Value{},
		NameFunc: func(t reflect.Type) string { return t.Name() },
	}
}

// Scope is passed to conversion funcs to allow them to continue an ongoing conversion.
// If multiple converters exist in the system, Scope will allow you to use the correct one
// from a conversion function--that is, the one your conversion function was called by.
type Scope interface {
	// Call Convert to convert sub-object. Note that if you call it with your own exact
	// parameters, you'll run out of stack space before anything useful happens.
	Convert(src, dest interface{}, flags FieldMatchingFlags) error

	// SrcTggs and DestTags contain the struct tags that src and dest had, respectively.
	// If the enclosing object was not a struct, then these will contain no tags, of cource.
	SrcTag() reflect.StructTag
	DestTag() reflect.StructTag

	// Flags returns the flags with which the conversion was started.
	Flags() FieldMatchingFlags

	// Meta returns any information originally passed to Convert.
	Meta() *Meta
}

// Meta is supplied by Scheme, when it calls Convert.
type Meta struct {
	SrcVersion  string
	DestVersion string

	// TODO: If needed, add a user dat field here.
}

// scope contains information about an ongoing conversion.
type scope struct {
	converter    *Converter
	meta         *Meta
	flags        FieldMatchingFlags
	srcTagStack  []reflect.StructTag
	destTagStack []reflect.StructTag
}

// push adds a level to the src/dest tag stacks.
func (s *scope) push() {
	s.srcTagStack = append(s.srcTagStack, "")
	s.destTagStack = append(s.destTagStack, "")
}

// pop removes a level to tag src/dest tag stacks.
func (s *scope) pop() {
	n := len(s.srcTagStack)
	s.srcTagStack = s.srcTagStack[:n-1]
	s.destTagStack = s.destTagStack[:n-1]
}

func (s *scope) setSrcTag(tag reflect.StructTag) {
	s.srcTagStack[len(s.srcTagStack)-1] = tag
}

func (s *scope) setDestTag(tag reflect.StructTag) {
	s.destTagStack[len(s.destTagStack)-1] = tag
}

// Converter continues a aconversion.
func (s *scope) Convert(src, dest interface{}, flags FieldMatchingFlags) error {
	return s.converter.Convert(src, dest, flags, s.meta)
}

// SrcTag returns the tag of the struct containing the current source item, if any.
func (s *scope) SrcTag() reflect.StructTag {
	return s.srcTagStack[len(s.srcTagStack)-1]
}

// DestTag returns the tag of the struct containing the current dest item, if any.
func (s *scope) DestTag() reflect.StructTag {
	return s.destTagStack[len(s.destTagStack)-1]
}

// Flags returns the flags with which the current conversion was started.
func (s *scope) Flags() FieldMatchingFlags {
	return s.flags
}

// Meta returns the meta object that was originally passed to Convert.
func (s *scope) Meta() *Meta {
	return s.meta
}

// Register registers a conversion func with the Converter. conversionFuncs must take
// three parameters: a pointer to the input type, a pointer to the output type, and
// a conversion.Scope (which should be used if recursive conversion calls are desired).
// It must return an error.
//
// Example:
//  c.Register(func(in *Pod, out *viveta1.Pod, s Scope) error { ... return nil })
func (c *Converter) Register(conversionFuncs interface{}) error {
	fv := reflect.ValueOf(conversionFuncs)
	ft := fv.Type()
	if ft.Kind() != reflect.Func {
		return fmt.Errorf("expected func, got: %v", ft)
	}
	if ft.NumIn() != 3 {
		return fmt.Errorf("expected three 'in' params, got: %v", ft)
	}
	if ft.NumOut() != 1 {
		return fmt.Errorf("expected one 'out' param, got: %v", ft)
	}
	if ft.In(0).Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer arg for 'in' param 0, got: %v", ft)
	}
	if ft.In(1).Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer arg for 'in' param 1, got: %v", ft)
	}
	scopeType := Scope(nil)
	if e, a := reflect.TypeOf(&scopeType).Elem(), ft.In(2); e != a {
		return fmt.Errorf("expected '%v' arg for 'in' param 2, got '%v' (%v)", e, a, ft)
	}
	var forErrorType error
	errorType := reflect.TypeOf(&forErrorType).Elem()
	if ft.Out(0) != errorType {
		return fmt.Errorf("expected error return, got: %v", ft)
	}
	c.funcs[typePair{ft.In(0).Elem(), ft.In(1).Elem()}] = fv
	return nil
}

// FieldMatchingFlags contains a list of ways in which struct fields could be copied.
// These constants may be | combined.
type FieldMatchingFlags int

const (
	// Loop thorough destination fields, search for matching source
	// field to copy it from. Source fields with no corresponding
	// destination field will be ignored. If SourceToDest is
	// specified, this flag is ignored. If neither is specified,
	// or no flags are passed, this flag is the default.
	DestFromSource FieldMatchingFlags = 0
	// Loop thorough source fields, search for matching dest field
	// to copy it into. Destination fields with no corresponding
	// source field will be ignored.
	SourceToDest FieldMatchingFlags = 1 << iota
	// Don't treat it as an error if the corresponding source or
	// dest field can't be found.
	IgnoreMissingFields
	// Don't require type name to match.
	AllowDifferentFieldTypeNames
)

// IsSet return true if the given flag or combination of flags is set.
func (f FieldMatchingFlags) IsSet(flag FieldMatchingFlags) bool {
	return f&flag == flag
}

func (c *Converter) Convert(src, dest interface{}, flags FieldMatchingFlags, meta *Meta) error {
	dv, sv := reflect.ValueOf(dest), reflect.ValueOf(src)
	if dv.Kind() != reflect.Ptr {
		return fmt.Errorf("Need pointer, but got %#v", dest)
	}
	if sv.Kind() != reflect.Ptr {
		return fmt.Errorf("Need pointer, but got %#v", src)
	}

	dv = dv.Elem()
	sv = sv.Elem()
	if !dv.CanAddr() {
		return fmt.Errorf("Can't write to dest")
	}
	s := &scope{
		converter: c,
		flags:     flags,
		meta:      meta,
	}
	s.push()

	return c.convert(sv, dv, s)
}

func (c *Converter) convert(sv, dv reflect.Value, scope *scope) error {
	dt, st := dv.Type(), sv.Type()
	if fv, ok := c.funcs[typePair{st, dt}]; ok {
		if c.Debug != nil {
			c.Debug.Logf("Calling custom conversion of '%v' to '%v'", st, dt)
		}
		args := []reflect.Value{sv.Addr(), dv.Addr(), reflect.ValueOf(scope)}
		ret := fv.Call(args)[0].Interface()
		if ret == nil {
			return nil
		}
		return ret.(error)
	}

	if !scope.flags.IsSet(AllowDifferentFieldTypeNames) && c.NameFunc(dt) != c.NameFunc(st) {
		return fmt.Errorf("Can't convert %v to %v because type names don't match", st, dt)
	}

	if st.AssignableTo(dt) {
		dv.Set(sv)
		return nil
	}
	if st.ConvertibleTo(dt) {
		dv.Set(sv.Convert(dt))
		return nil
	}

	if c.Debug != nil {
		c.Debug.Logf("Trying to convert '%v' to '%v'", st, dt)
	}

	scope.push()
	defer scope.pop()

	switch dv.Kind() {
	case reflect.Struct:
		listType := dt
		if scope.flags.IsSet(SourceToDest) {
			listType = st
		}
		// log.Printf("DV: %v, SV: %v", dt.Name(), st.Name())
		for i := 0; i < listType.NumField(); i++ {
			f := listType.Field(i)
			df := dv.FieldByName(f.Name)
			sf := sv.FieldByName(f.Name)
			if sf.IsValid() {
				// No need to check error, since we know it's valid
				field, _ := st.FieldByName(f.Name)
				scope.setSrcTag(field.Tag)
			}
			if df.IsValid() {
				field, _ := dt.FieldByName(f.Name)
				scope.setDestTag(field.Tag)
			}
			// TODO: set top level of scope.src/destTagStack with these field tags here.
			if !df.IsValid() || !sf.IsValid() {
				switch {
				case scope.flags.IsSet(IgnoreMissingFields):
					// No error.
				case scope.flags.IsSet(SourceToDest):
					return fmt.Errorf("%v not present in dest (%v to %v)", f.Name, st, dt)
				default:
					return fmt.Errorf("%v not present in src (%v to %v)", f.Name, st, dt)
				}
				continue
			}
			// log.Printf("Next sf: %v", f.Name)
			if err := c.convert(sf, df, scope); err != nil {
				return err
			}
		}
	case reflect.Slice:
		if sv.IsNil() {
			// Don't make a zero-length slice.
			dv.Set(reflect.Zero(dt))
			return nil
		}
		dv.Set(reflect.MakeSlice(dt, sv.Len(), sv.Cap()))
		for i := 0; i < sv.Len(); i++ {
			if err := c.convert(sv.Index(i), dv.Index(i), scope); err != nil {
				return err
			}
		}
	case reflect.Ptr:
		if sv.IsNil() {
			// Don't copy a nil ptr!
			dv.Set(reflect.Zero(dt))
			return nil
		}
		dv.Set(reflect.New(dt.Elem()))
		return c.convert(sv.Elem(), dv.Elem(), scope)
	case reflect.Map:
		if sv.IsNil() {
			// Don't copy a nil ptr!
			dv.Set(reflect.Zero(dt))
			return nil
		}
		dv.Set(reflect.MakeMap(dt))
		for _, sk := range sv.MapKeys() {
			dk := reflect.New(dt.Key()).Elem()
			if err := c.convert(sk, dk, scope); err != nil {
				return err
			}
			dkv := reflect.New(dt.Elem()).Elem()
			if err := c.convert(sv.MapIndex(sk), dkv, scope); err != nil {
				return err
			}
			dv.SetMapIndex(dk, dkv)
		}
	default:
		return fmt.Errorf("Couldn't copy '%v' into '%v'", st, dt)
	}
	return nil
}
