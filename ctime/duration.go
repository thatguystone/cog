package ctime

import "time"

const (
	// Nanosecond is a HumanDuration-compatible duration
	Nanosecond HumanDuration = HumanDuration(time.Nanosecond)

	// Microsecond is a HumanDuration-compatible duration
	Microsecond = HumanDuration(time.Microsecond)

	// Millisecond is a HumanDuration-compatible duration
	Millisecond = HumanDuration(time.Millisecond)

	// Second is a HumanDuration-compatible duration
	Second = HumanDuration(time.Second)

	// Minute is a HumanDuration-compatible duration
	Minute = HumanDuration(time.Minute)

	// Hour is a HumanDuration-compatible duration
	Hour = HumanDuration(time.Hour)
)
