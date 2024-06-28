// langur/object/datetime_ops.go

package object

import (
	"langur/native"
	"time"
)

func (l *DateTime) Add(dt2 Object) Object {
	switch r := dt2.(type) {
	case *Duration:
		years, ok := native.ToInt(r.Years)
		if !ok {
			return NewError(ERR_MATH, "Add", "failure to convert years of duration to native Go int")
		}
		months, ok := native.ToInt(r.Months)
		if !ok {
			return NewError(ERR_MATH, "Add", "failure to convert months of duration to native Go int")
		}
		days, ok := native.ToInt(r.Days)
		if !ok {
			return NewError(ERR_MATH, "Add", "failure to convert days of duration to native Go int")
		}
		return &DateTime{Time: l.Time.AddDate(years, months, days).Add(
			time.Duration(r.Hours*nsPerHour + r.Minutes*nsPerMin + r.Seconds*nsPerSec + r.Nanoseconds))}

	case *Number:
		nsec, err := r.ToInt64()
		if err != nil {
			return NewError(ERR_MATH, "Add", "failure to convert number to native Go int64")
		}
		return &DateTime{Time: l.Time.Add(time.Duration(nsec)).Round(0)}

	default:
		return nil
	}
}

func (l *DateTime) Subtract(dt2 Object) Object {
	switch r := dt2.(type) {
	case *DateTime:
		return l.Difference(r)

	case *Duration:
		years, ok := native.ToIntNegated(r.Years)
		if !ok {
			return NewError(ERR_MATH, "Subtract", "failure to convert years of duration to native Go int")
		}
		months, ok := native.ToIntNegated(r.Months)
		if !ok {
			return NewError(ERR_MATH, "Subtract", "failure to convert months of duration to native Go int")
		}
		days, ok := native.ToIntNegated(r.Days)
		if !ok {
			return NewError(ERR_MATH, "Subtract", "failure to convert days of duration to native Go int")
		}
		return &DateTime{Time: l.Time.AddDate(years, months, days).Add(
			time.Duration(-(r.Hours*nsPerHour + r.Minutes*nsPerMin + r.Seconds*nsPerSec + r.Nanoseconds)))}

	case *Number:
		nsec, err := r.ToInt64()
		if err != nil {
			return NewError(ERR_MATH, "Subtract", "failure to convert number to native Go int64")
		}
		return &DateTime{Time: l.Time.Add(time.Duration(-nsec)).Round(0)}

	default:
		return nil
	}
}

func (left *DateTime) Difference(right *DateTime) *Duration {
	// thanks to help from ...
	// https://stackoverflow.com/questions/36530251/time-since-with-months-and-years/36531443, 1/22/24

	daysInMonth := func(year int, month time.Month) int {
		return 32 - time.Date(year, month, 32, 0, 0, 0, 0, time.UTC).Day()
	}

	a, b := left.Time, right.Time
	if a.Location() != b.Location() {
		b = b.In(a.Location())
	}
	if a.After(b) {
		a, b = b, a
	}

	var years, months, days, hours, minutes, seconds, nanoseconds int

	y1, M1, d1 := a.Date()
	y2, M2, d2 := b.Date()

	h1, m1, s1 := a.Clock()
	h2, m2, s2 := b.Clock()

	ns1, ns2 := a.Nanosecond(), b.Nanosecond()

	years = y2 - y1
	months = int(M2 - M1)
	days = d2 - d1

	hours = h2 - h1
	minutes = m2 - m1
	seconds = s2 - s1
	nanoseconds = ns2 - ns1

	if nanoseconds < 0 {
		nanoseconds += 1e9
		seconds--
	}
	if seconds < 0 {
		seconds += 60
		minutes--
	}
	if minutes < 0 {
		minutes += 60
		hours--
	}
	if hours < 0 {
		hours += 24
		days--
	}
	if days < 0 {
		days += daysInMonth(y1, M1)
		months--
	}
	if months < 0 {
		months += 12
		years--
	}

	return NewDuration(int64(years), int64(months), int64(days), int64(hours), int64(minutes), int64(seconds), int64(nanoseconds))
}
