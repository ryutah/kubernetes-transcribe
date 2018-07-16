package labels

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ryutah/kubernetes-transcribe/pkg/util"
)

// Selector represents a label selector.
type Selector interface {
	// Matches returns true if this selector matches the given set of labels.
	Matches(Labels) bool

	// Empty returns true if this selector does not restrict the selection space.
	Empty() bool

	// RequiredExactMatch allows a caller to introspect whether a given selector
	// requires a single specific label to be set, and if so returns the value it
	// requires.
	// TODO: expand this to be more general
	RequiredExactMatch(label string) (value string, found bool)

	// String returns a human readable string that represents this selector.
	String() string
}

// Everything returns a selector that matches all labels.
func Everything() Selector {
	return andTerm{}
}

type hasTerm struct {
	label string
	value string
}

func (h *hasTerm) Matches(ls Labels) bool {
	return ls.Get(h.label) == h.value
}

func (h *hasTerm) Empty() bool {
	return false
}

func (h *hasTerm) RequiredExactMatch(label string) (value string, found bool) {
	if h.label == label {
		return h.value, true
	}
	return "", false
}

func (h *hasTerm) String() string {
	return fmt.Sprintf("%v=%v", h.label, h.value)
}

type notHasTerm struct {
	label string
	value string
}

func (n *notHasTerm) Matches(ls Labels) bool {
	return ls.Get(n.label) != n.value
}

func (n *notHasTerm) Empty() bool {
	return false
}

func (n *notHasTerm) RequiredExactMatch(label string) (value string, found bool) {
	return "", false
}

func (n *notHasTerm) String() string {
	return fmt.Sprintf("%v!=%v", n.label, n.value)
}

type andTerm []Selector

func (a andTerm) Matches(ls Labels) bool {
	for _, q := range a {
		if !q.Matches(ls) {
			return false
		}
	}
	return true
}

func (a andTerm) Empty() bool {
	if a == nil {
		return true
	}
	if len(a) == 0 {
		return true
	}
	for _, t := range a {
		if !t.Empty() {
			return false
		}
	}
	return true
}

func (a andTerm) RequiredExactMatch(label string) (value string, found bool) {
	if a == nil || len(a) == 0 {
		return "", false
	}
	for _, t := range a {
		if value, found := t.RequiredExactMatch(label); found {
			return value, found
		}
	}
	return "", false
}

func (a andTerm) String() string {
	var terms []string
	for _, q := range a {
		terms = append(terms, q.String())
	}
	return strings.Join(terms, ",")
}

// Operator represents a key's relationship to a set of values in a Requirement.
// TODO: Should also represent key's existance.
type Operator int

const (
	IN Operator = iota + 1
	NOT_IN
)

type Requirement struct {
	key       string
	operator  Operator
	strValues util.StringSet
}

func (r *Requirement) Matches(ls Labels) bool {
	switch r.operator {
	case IN:
		return r.strValues.Has(ls.Get(r.key))
	case NOT_IN:
		return !r.strValues.Has(ls.Get(r.key))
	default:
		return false
	}
}

// LabelSelector only not named 'Selector' due to name conflict until Selector is deprecated.
type LabelSelector struct {
	Requirement []Requirement
}

func (l *LabelSelector) Matches(ls Labels) bool {
	for _, req := range l.Requirement {
		if !req.Matches(ls) {
			return false
		}
	}
	return true
}

func try(selectorPiece, op string) (lhs, rhs string, ok bool) {
	pieces := strings.Split(selectorPiece, op)
	if len(pieces) == 2 {
		return pieces[0], pieces[1], true
	}
	return "", "", false
}

// SelectorFromSet returns a Selector which will match exactly the given Set.
// A nil Set is considered equivalent to Everything().
func SelectorFromSet(ls Set) Selector {
	if ls == nil {
		return Everything()
	}
	items := make([]Selector, 0, len(ls))
	for label, value := range ls {
		items = append(items, &hasTerm{label: label, value: value})
	}
	if len(items) == 1 {
		return items[0]
	}
	return andTerm(items)
}

// ParseSelector takes a string representing a selector and returns an object suitable for matching, or an error.
func ParseSelector(selector string) (Selector, error) {
	parts := strings.Split(selector, ",")
	sort.Strings(parts)
	var items []Selector
	for _, part := range parts {
		if part == "" {
			continue
		}
		if lhs, rhs, ok := try(part, "!="); ok {
			items = append(items, &notHasTerm{label: lhs, value: rhs})
		} else if lhs, rhs, ok := try(part, "=="); ok {
			items = append(items, &hasTerm{label: lhs, value: rhs})
		} else if lhs, rhs, ok := try(part, "="); ok {
			items = append(items, &hasTerm{label: lhs, value: rhs})
		} else {
			return nil, fmt.Errorf("invalid selector: '%s'; can't understand '%s'", selector, part)
		}
	}
	if len(items) == 1 {
		return items[0], nil
	}
	return andTerm(items), nil
}
