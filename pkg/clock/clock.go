package clock

import "time"

var (
	// Default is a proxy to the [time] package.
	Default = defaultClock{}
)

// Clock defines methods of a clock.
type Clock interface {
	Now() time.Time
}

type defaultClock struct{}

func (dc defaultClock) Now() time.Time {
	return time.Now()
}

// Fake fakes the clock.
// Useful in tests.
type Fake struct {
	Base time.Time
}

// NewFakeDefault returns [Fake] initialized at 2000-01-01T00:00:00Z.
func NewFakeDefault() *Fake {
	return &Fake{
		Base: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

// Now fakes the current time.
// It returns whatever base value to clock currently contains.
// It increases the base value by one second for each call
// to make it possible to check a sequence of calls.
func (fc *Fake) Now() time.Time {
	now := fc.Base
	fc.Base = fc.Base.Add(1 * time.Second)
	return now
}
