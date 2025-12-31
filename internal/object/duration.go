// langur/object/duration.go

package object

import (
	"fmt"
	"langur/common"
	"langur/str"
	"strings"
)

// duration objects keeping a separate count of years, months, etc.

type Duration struct {
	Years       int64
	Months      int64
	Days        int64
	Hours       int64
	Minutes     int64
	Seconds     int64
	Nanoseconds int64
}

var ZeroDuration = &Duration{}

func NewDurationFromString(s string) (*Duration, error) {
	return durationStringToObject(s)
}

// NOTE: assuming only positive numbers
func NewDuration(years, months, days, hours, minutes, seconds, nanoseconds int64) *Duration {
	return &Duration{Years: years, Months: months, Days: days,
		Hours: hours, Minutes: minutes, Seconds: seconds, Nanoseconds: nanoseconds}
}

func NewDurationFromNanoseconds(nanoseconds int64) *Duration {
	var years, months, days, hours, minutes, seconds int64

	// cannot be negative
	if nanoseconds < 0 {
		nanoseconds = -nanoseconds
	}

	years, nanoseconds = nanoseconds/nsPerYear, nanoseconds%nsPerYear
	months, nanoseconds = nanoseconds/nsPerMon, nanoseconds%nsPerMon
	days, nanoseconds = nanoseconds/nsPerDay, nanoseconds%nsPerDay
	hours, nanoseconds = nanoseconds/nsPerHour, nanoseconds%nsPerHour
	minutes, nanoseconds = nanoseconds/nsPerMin, nanoseconds%nsPerMin
	seconds, nanoseconds = nanoseconds/nsPerSec, nanoseconds%nsPerSec

	return &Duration{Years: years, Months: months, Days: days,
		Hours: hours, Minutes: minutes, Seconds: seconds, Nanoseconds: nanoseconds}
}

func NewDurationOfNanoseconds(nanoseconds int64) *Duration {
	// just nanoseconds

	// cannot be negative
	if nanoseconds < 0 {
		nanoseconds = -nanoseconds
	}
	
	return &Duration{Nanoseconds: nanoseconds}
}

func (d *Duration) Copy() Object {
	return &Duration{
		Years:       d.Years,
		Months:      d.Months,
		Days:        d.Days,
		Hours:       d.Hours,
		Minutes:     d.Minutes,
		Seconds:     d.Seconds,
		Nanoseconds: d.Nanoseconds,
	}
}

func (l *Duration) Equal(d2 Object) bool {
	right, ok := d2.(*Duration)
	if !ok {
		return false
	}
	return l.Years == right.Years && l.Months == right.Months && l.Days == right.Days &&
		l.Hours == right.Hours && l.Minutes == right.Minutes && l.Seconds == right.Seconds && l.Nanoseconds == right.Nanoseconds
}

func (l *Duration) GreaterThan(d2 Object) (gt, comparable bool) {
	r, ok := d2.(*Duration)
	if !ok {
		return false, false
	}
	comparable = true

	// return l.ToNanoseconds() > r.ToNanoseconds(), true

	// FIXME(?): a somewhat naive implementation
	if l.Years > r.Years {
		gt = true
	} else if l.Years == r.Years {
		if l.Months > r.Months {
			gt = true
		} else if l.Months == r.Months {
			if l.Days > r.Days {
				gt = true
			} else if l.Days == r.Days {
				if l.Hours > r.Hours {
					gt = true
				} else if l.Hours == r.Hours {
					if l.Minutes > r.Minutes {
						gt = true
					} else if l.Minutes == r.Minutes {
						if l.Seconds > r.Seconds {
							gt = true
						} else if l.Seconds == r.Seconds {
							if l.Nanoseconds > r.Nanoseconds {
								gt = true
							}
						}
					}
				}
			}
		}
	}
	return
}

func (d *Duration) Type() ObjectType {
	return DURATION_OBJ
}
func (d *Duration) TypeString() string {
	return common.DurationTypeName
}

func (d *Duration) IsTruthy() bool {
	return !d.IsZero()
}

func (d *Duration) IsZero() bool {
	// return d.Equal(ZeroDuration)

	return !(d.Years != 0 || d.Months != 0 || d.Days != 0 ||
		d.Hours != 0 || d.Minutes != 0 || d.Seconds != 0 || d.Nanoseconds != 0)
}

func (d *Duration) ComposedString() string {
	return common.DurationTokenLiteral + "/" + d.String() + "/"
}

func (d *Duration) ReplString() string {
	return common.DurationTypeName + " " + d.String()
}

func (d *Duration) String() string {
	// ISO 8601
	var sb strings.Builder
	sb.WriteRune('P')

	if d.Years != 0 {
		sb.WriteString(str.Int64ToStr(d.Years, 10))
		sb.WriteByte('Y')
	}
	if d.Months != 0 {
		sb.WriteString(str.Int64ToStr(d.Months, 10))
		sb.WriteByte('M')
	}
	if d.Days != 0 {
		sb.WriteString(str.Int64ToStr(d.Days, 10))
		sb.WriteByte('D')
	}

	if d.Hours != 0 || d.Minutes != 0 || d.Seconds != 0 || d.Nanoseconds != 0 {
		sb.WriteString("T")
		if d.Hours != 0 {
			sb.WriteString(str.Int64ToStr(d.Hours, 10))
			sb.WriteByte('H')
		}
		if d.Minutes != 0 {
			sb.WriteString(str.Int64ToStr(d.Minutes, 10))
			sb.WriteByte('M')
		}
		if d.Seconds != 0 || d.Nanoseconds != 0 {
			sb.WriteString(str.Int64ToStr(d.Seconds, 10))

			if d.Nanoseconds != 0 {
				sb.WriteByte('.')
				nss := str.Int64ToStr(d.Nanoseconds, 10)
				nss = str.PadLeft(nss, 9, '0')
				nss = str.RemoveTrailing(nss, '0')
				sb.WriteString(nss)
			}
			sb.WriteByte('S')
		}
	}

	if sb.Len() == 1 {
		sb.WriteString("T0S")
	}

	return sb.String()
}

func durationStringToObject(s string) (*Duration, error) {
	if s == "" {
		// return a duration of 0
		return &Duration{}, nil
	}

	m := dtRegexDuration.FindStringSubmatch(s)
	if m == nil {
		return nil, fmt.Errorf("Invalid duration string")
	}
	names := dtRegexDuration.SubexpNames()

	years, okyears := subMatchToInt64("years", m, names)
	months, okmonths := subMatchToInt64("months", m, names)
	weeks, okweeks := subMatchToInt64("weeks", m, names)
	days, okdays := subMatchToInt64("days", m, names)

	hours, okhours := subMatchToInt64("hours", m, names)
	minutes, okminutes := subMatchToInt64("minutes", m, names)
	seconds, okseconds := subMatchToInt64("seconds", m, names)

	if !okyears && !okmonths && !okweeks && !okdays && !okhours && !okminutes && !okseconds {
		return nil, fmt.Errorf("Invalid duration string; expected at least one of years/months/days/hours/minutes/seconds or weeks")
	}

	secondsfraction := subMatchByName("secondsfraction", m, names)
	var ns int64 = 0
	if secondsfraction != "" {
		secondsfraction = str.PadRight(secondsfraction, 9, '0')
		ns, _ = str.StrToInt64(secondsfraction, 10)
	}

	return &Duration{Years: years, Months: months, Days: weeks*7 + days,
		Hours: hours, Minutes: minutes, Seconds: seconds, Nanoseconds: ns}, nil
}

func (d *Duration) ToNanoseconds() int64 {
	return nsPerYear*d.Years + nsPerMon*d.Months + nsPerDay*d.Days +
		nsPerHour*d.Hours + nsPerMin*d.Minutes + nsPerSec*d.Seconds + d.Nanoseconds
}

// for building/interpreting hashes
var durStringYears = NewString("years")
var durStringMonths = NewString("months")
var durStringDays = NewString("days")
var durStringHours = NewString("hours")
var durStringMinutes = NewString("minutes")
var durStringSeconds = NewString("seconds")
var durStringNanoseconds = NewString("nanoseconds")

func (d *Duration) ToHash() *Hash {
	hash := &Hash{}

	hash.WritePair(durStringYears, NumberFromInt64(d.Years))
	hash.WritePair(durStringMonths, NumberFromInt64(d.Months))
	hash.WritePair(durStringDays, NumberFromInt64(d.Days))
	hash.WritePair(durStringHours, NumberFromInt64(d.Hours))
	hash.WritePair(durStringMinutes, NumberFromInt64(d.Minutes))
	hash.WritePair(durStringSeconds, NumberFromInt64(d.Seconds))
	hash.WritePair(durStringNanoseconds, NumberFromInt64(d.Nanoseconds))

	return hash
}

func noNegatives(src string) error {
	return NewError(ERR_ARGUMENTS, src, "Duration values cannot be negative.")
}

func (d *Hash) ToDuration() (dr *Duration, err error) {
	var years, months, days, hours, minutes, seconds, nanoseconds int64

	// using tryHashInteger64Value() instead of getHashInteger64Value() to ignore missing keys

	years, err = tryHashInteger64Value(d, durStringYears, 0)
	if err != nil {
		return
	}
	if years < 0 {
		err = noNegatives("years")
		return
	}
	months, err = tryHashInteger64Value(d, durStringMonths, 0)
	if err != nil {
		return
	}
	if months < 0 {
		err = noNegatives("months")
		return
	}
	days, err = tryHashInteger64Value(d, durStringDays, 0)
	if err != nil {
		return
	}
	if days < 0 {
		err = noNegatives("days")
		return
	}
	hours, err = tryHashInteger64Value(d, durStringHours, 0)
	if err != nil {
		return
	}
	if hours < 0 {
		err = noNegatives("hours")
		return
	}
	minutes, err = tryHashInteger64Value(d, durStringMinutes, 0)
	if err != nil {
		return
	}
	if minutes < 0 {
		err = noNegatives("minutes")
		return
	}
	seconds, err = tryHashInteger64Value(d, durStringSeconds, 0)
	if err != nil {
		return
	}
	if seconds < 0 {
		err = noNegatives("seconds")
		return
	}
	nanoseconds, err = tryHashInteger64Value(d, durStringNanoseconds, 0)
	if err != nil {
		return
	}
	if nanoseconds < 0 {
		err = noNegatives("nanoseconds")
		return
	}

	return NewDuration(years, months, days, hours, minutes, seconds, nanoseconds), err
}
