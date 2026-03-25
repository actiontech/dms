package timeutil

import "time"

// ToUTC returns a pointer to the same instant expressed in UTC.
// Returns nil if t is nil.
//
// The go-sql-driver/mysql with loc=Local symmetrically converts on both
// write (In(loc)) and read (parsed as loc), so the time.Time already
// carries the correct instant; calling UTC() only normalises the zone.
func ToUTC(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	utc := t.UTC()
	return &utc
}

// NowUTC returns the current time in UTC.
func NowUTC() time.Time {
	return time.Now().UTC()
}
