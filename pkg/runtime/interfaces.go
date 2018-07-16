package runtime

// Decoder defines methods for deserializing API objects into a given type
type Decoder interface {
	Decode(data []byte) (Object, error)
	DecodeInto(data []byte, obj Object) error
}

// Encoder defines methods for serializing API objects into bytes
type Encoder interface {
	Encode(obj Object) (data []byte, err error)
}

// Codec defines methods for serializing and deserializing API objects.
type Codec interface {
	Decoder
	Encoder
}

// ResourceVersioner provides methods for setting and retrieving
// the resource version from an API object.
type ResourceVersioner interface {
	SetResourceVersion(obj Object, version uint64) error
	ResourceVersion(obj Object) (uint64, error)
}

// Object is interface must be support all api types. It's deliberately tiny so that this is not an onerous burden.
// Implement it with a pointer receiver; this will allow us to use the go compiler to check the
// one thing about our objects that it's capable of checking for us.
type Object interface {
	// This function is used only to enforce membership. It's never called.
	// TODO: Consider mass rename in the future to make it do something useful.
	IsAnAPIObject()
}
