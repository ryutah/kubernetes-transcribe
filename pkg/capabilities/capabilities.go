package capabilities

import (
	"sync"
)

// Capabilities defines the set of capabilities available within whe system.
// For now these are global.  Eventually they may be per-user
type Capabilities struct {
	AllowPrivileged bool
}

var once sync.Once
var capabilities *Capabilities

// Initialize the capabilities set.  This can only be done once per binary, subsequent calls are ignored.
func Initialize(c Capabilities) {
	// Only do this once
	once.Do(func() {
		capabilities = &c
	})
}

// SetCapabilitiesForTests.  Convenience method for testing.  This should only be called from tests.
func SetForTests(c Capabilities) {
	capabilities = &c
}

// Returns a read-only copy of the system capabilities.
func Get() Capabilities {
	if capabilities == nil {
		if capabilities == nil {
			Initialize(Capabilities{
				AllowPrivileged: false,
			})
		}
	}
	return *capabilities
}
