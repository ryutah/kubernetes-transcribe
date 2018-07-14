package util

import (
	"encoding/json"
	"time"
)

type Time struct {
	time.Time
}

func Date(year int, month time.Month, day, hour, min, sec, nsec int, loc *time.Location) Time {
	return Time{time.Date(year, month, day, hour, min, sec, nsec, loc)}
}

func Now() Time {
	return Time{time.Now()}
}

func Unix(sec, nsec int64) Time {
	return Time{time.Unix(sec, nsec)}
}

func (t Time) Rfc3339Copy() Time {
	copied, _ := time.Parse(time.RFC3339, t.Format(time.RFC3339))
	return Time{copied}
}

// UnmarshalJSON implements json.Marshaler interface.
func (t *Time) UnmarshalJSON(b []byte) error {
	if len(b) == 4 && string(b) == "null" {
		t.Time = time.Time{}
		return nil
	}

	var str string
	json.Unmarshal(b, &str)

	pt, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return err
	}
	t.Time = pt
	return nil
}

// MarshalJSON implements json.Marshaler interface.
func (t Time) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}

	return json.Marshal(t.Format(time.RFC3339))
}

// SetYAML implements the yaml.Setter interface.
func (t *Time) SetYAML(tag string, value interface{}) bool {
	if value == nil {
		t.Time = time.Time{}
		return true
	}

	str, ok := value.(string)
	if !ok {
		return false
	}

	pt, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return false
	}

	t.Time = pt
	return true
}

// GetYAML implements the yaml.Setter interface.
func (t Time) GetYAML() (tag string, value interface{}) {
	if t.IsZero() {
		value = "null"
		return
	}

	value = t.Format(time.RFC3339)
	return
}
