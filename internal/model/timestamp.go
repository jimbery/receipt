package model

import "time"

// Timestamp is a UTC instant with the original zone offset retained for diagnostics.
// Comparison always uses UTC; offset captures timezone skew from upstream sources.
type Timestamp struct {
	UTC    time.Time
	Offset int // seconds east of UTC, as reported by the source
}

func NewTimestamp(utc time.Time, offsetSeconds int) Timestamp {
	return Timestamp{UTC: utc.UTC(), Offset: offsetSeconds}
}

func TimestampFromTime(t time.Time) Timestamp {
	_, offset := t.Zone()
	return Timestamp{UTC: t.UTC(), Offset: offset}
}
