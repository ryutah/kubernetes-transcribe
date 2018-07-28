package health

// Status represents the result of a single health-check operation.
type Status int

// Status values must be one of these constants.
const (
	Healthy Status = iota
	Unhealthy
	Unknown
)

func (s Status) String() string {
	if s == Healthy {
		return "healthy"
	} else if s == Unhealthy {
		return "unhealthy"
	}
	return "unknown"
}
