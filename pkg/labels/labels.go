package labels

import (
	"sort"
	"strings"
)

// Labels allows you to present labels independently from their storage.
type Labels interface {
	// Get returns the value for the provided label.
	Get(label string) (value string)
}

// Set is a map of label:value. It implements Labels.
type Set map[string]string

func (ls Set) String() string {
	selector := make([]string, 0, len(ls))
	for key, value := range ls {
		selector = append(selector, key+"="+value)
	}
	sort.StringSlice(selector).Sort()
	return strings.Join(selector, ",")
}

// Get returns the value in the map for the provided label.
func (ls Set) Get(label string) string {
	return ls[label]
}

// AsSelector converts labels into a selectors.
func (ls Set) AsSelector() Selector {
	return nil
}
