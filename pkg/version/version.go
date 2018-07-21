package version

// Info contains versioning information.
// TODO: Add []string of api versions supperted? It's still unclear
// how we'll want to distribute that information.
type Info struct {
	Major        string `json:"major" yaml:"major"`
	Minor        string `json:"minor" yaml:"minor"`
	GitVersion   string `json:"gitVersion" yaml:"gitVersion"`
	GitCommit    string `json:"gitCommit" yaml:"gitCommit"`
	GitTreeState string `json:"gitTreeState" yaml:"gitTreeState"`
}

// Get returns the overall codebase version. It's for detecting
// what code a binary was build from.
func Get() Info {
	// These veriables typically come from -ldflags settings and in
	// their absence fallback to the settings in pkg/version/base.go
	return Info{
		Major:        gitMajor,
		Minor:        gitMinor,
		GitVersion:   gitVersion,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
	}
}

// String returns info as human-friendly version string.
func (info Info) String() string {
	return info.GitVersion
}
