package tcg

import "time"

// Now returns the current time. It is defined for easier mocking in bindings or tests.
func Now() time.Time {
	return time.Now()
}
