// langur/object/datetime.go

package object

import (
	"fmt"
	"langur/common"
	"langur/regexp"
	"langur/str"
	"time"
)

type DateTime struct {
	Time time.Time
}

func (dt *DateTime) Copy() Object {
	return &DateTime{
		Time: dt.Time,
	}
}

func (l *DateTime) Equal(dt2 Object) bool {
	r, ok := dt2.(*DateTime)
	if !ok {
		return false
	}
	return l.Time.Equal(r.Time)
}

func (l *DateTime) GreaterThan(dt2 Object) (bool, bool) {
	right, ok := dt2.(*DateTime)
	if !ok {
		return false, false
	}
	return l.Time.After(right.Time), true
}

func (dt *DateTime) Type() ObjectType {
	return DATETIME_OBJ
}
func (dt *DateTime) TypeString() string {
	return common.DateTimeTypeName
}

func (dt *DateTime) IsTruthy() bool {
	// proleptic Gregorian as false

	// any date prior to 1582-10-15 as proleptic Gregorian
	// time zone neutral (not dependent on time in UTC)
	year, month, day := dt.Time.Date()
	if year < 1582 {
		return false

	} else if year == 1582 {
		m := int(month)
		if m < 10 {
			return false

		} else if m == 10 {
			if day < 15 {
				return false
			}
		}
	}

	// not proleptic
	return true
}

func (dt *DateTime) ComposedString() string {
	return common.DateTimeTokenLiteral + "/" + dt.String() + "/"
}

func (dt *DateTime) ReplString() string {
	return common.DateTimeTypeName + " " + dt.String()
}

func (dt *DateTime) String() string {
	return dt.Time.Format(Format_RFC3339Nano)
}

func (dt *DateTime) FormatString(format string) string {
	return dt.Time.Format(format)
}

// for building/interpreting hashes
var dtStringYear = NewString("year")
var dtStringMonth = NewString("month")
var dtStringMonthName = NewString("Month")
var dtStringDay = NewString("day")
var dtStringWeekDay = NewString("weekday")
var dtStringWeekDayName = NewString("Weekday")
var dtStringHour24 = NewString("hour24")
var dtStringHour = NewString("hour")
var dtStringAMPM = NewString("ampm")
var dtStringMinute = NewString("minute")
var dtStringSecond = NewString("second")
var dtStringNanosecond = NewString("nanosecond")
var dtStringTimeZoneName = NewString("ZONE")
var dtStringTimeZoneOffsetVisual = NewString("Zone")
var dtStringTimeZoneOffsetSeconds = NewString("zone")

func (dt *DateTime) ToHash() *Hash {
	hash := &Hash{}

	hour24 := dt.Time.Hour()
	hour := hour24
	ampm := "AM"
	if hour == 0 {
		hour = 12
	} else if hour == 12 {
		ampm = "PM"
	} else if hour > 12 {
		hour -= 12
		ampm = "PM"
	}

	tzname, tzoffset := dt.Time.Zone()

	hash.WritePair(dtStringYear, NumberFromInt(dt.Time.Year()))
	hash.WritePair(dtStringMonth, NumberFromInt(int(dt.Time.Month())))
	hash.WritePair(dtStringMonthName, NewString(dt.Time.Month().String()))
	hash.WritePair(dtStringDay, NumberFromInt(dt.Time.Day()))
	hash.WritePair(dtStringWeekDay, NumberFromInt(int(dt.Time.Weekday())+1)) // +1 as langur uses 1-based indexing
	hash.WritePair(dtStringWeekDayName, NewString(dt.Time.Weekday().String()))
	hash.WritePair(dtStringHour24, NumberFromInt(hour24))
	hash.WritePair(dtStringHour, NumberFromInt(hour))
	hash.WritePair(dtStringMinute, NumberFromInt(dt.Time.Minute()))
	hash.WritePair(dtStringSecond, NumberFromInt(dt.Time.Second()))
	hash.WritePair(dtStringNanosecond, NumberFromInt(dt.Time.Nanosecond()))
	hash.WritePair(dtStringAMPM, NewString(ampm))
	hash.WritePair(dtStringTimeZoneOffsetVisual, NewString(offsetSecondsToString(tzoffset)))
	hash.WritePair(dtStringTimeZoneOffsetSeconds, NumberFromInt(tzoffset))
	hash.WritePair(dtStringTimeZoneName, NewString(tzname))

	return hash
}

func (d *Hash) ToDateTime() (dt *DateTime, err error) {
	dt = &DateTime{}

	var year, month, day, hour, min, sec, nsec, tzseconds int
	var tzvisual string
	var loc *time.Location

	year, err = getHashIntegerValue(d, dtStringYear)
	if err != nil {
		return
	}
	month, err = getHashIntegerValue(d, dtStringMonth)
	if err != nil {
		return
	}
	if month < 1 || month > 12 {
		err = fmt.Errorf("Month value (%d) out of range (1-12)", month)
		return
	}
	day, err = getHashIntegerValue(d, dtStringDay)
	if err != nil {
		return
	}
	hour, err = getHashIntegerValue(d, dtStringHour24)
	if err != nil {
		hour, _ = getHashIntegerValue(d, dtStringHour)
		ampm, _ := getHashStringValue(d, dtStringAMPM)
		if hour != 12 && (ampm == "PM" || ampm == "pm") {
			hour += 12
		}
	}
	min, err = getHashIntegerValue(d, dtStringMinute)
	if err != nil {
		return
	}
	sec, err = getHashIntegerValue(d, dtStringSecond)
	if err != nil {
		return
	}
	nsec, err = getHashIntegerValue(d, dtStringNanosecond)
	if err != nil {
		return
	}

	tzseconds, err = getHashIntegerValue(d, dtStringTimeZoneOffsetSeconds)
	if err == nil {
		loc = time.FixedZone("", tzseconds)
	} else {
		tzvisual, err = getHashStringValue(d, dtStringTimeZoneOffsetVisual)
		if err == nil {
			loc, err = stringToLocation(tzvisual, false)
			if err != nil {
				return
			}
		} else {
			loc = time.Local
		}
	}

	dt.Time = time.Date(year, time.Month(month), day, hour, min, sec, nsec, loc)
	return
}

func NewDateTimeFromNanoseconds(nsec int64) (dt *DateTime) {
	dt = &DateTime{Time: time.Unix(0, nsec).Round(0)}
	return
}

func NewDateTimeFromSeconds(sec int64) (dt *DateTime) {
	dt = &DateTime{Time: time.Unix(sec, 0).Round(0)}
	return
}

func NewDateTime(from time.Time, includesNano bool) (dt *DateTime) {
	if includesNano {
		dt = &DateTime{Time: from}
		return
	}
	// set nanoseconds to 0
	dt = &DateTime{Time: from.Truncate(time.Second)}
	return
}

func NewDateTimeFromStringFormat(s, format string) (dt *DateTime, err error) {
	dt = &DateTime{}
	dt.Time, err = time.Parse(format, s)
	return
}

func StringForNowDateTime(s string, literal bool) bool {
	// case 1: a location string
	_, err := stringToLocation(s, literal)
	if err == nil {
		return true
	}
	// case 2: a time zone string
	if literal {
		return !dtRegexTZLiteral.MatchString(s)
	}
	return !dtRegexTZ.MatchString(s)
}

func NewDateTimeFromLiteralString(s string, nowIncludesFractionalSeconds bool) (dt *DateTime, err error) {
	return NewDateTimeFromString(s, nowIncludesFractionalSeconds, true)
}

func NewDateTimeFromString(s string, nowIncludesFractionalSeconds, literal bool) (dt *DateTime, err error) {
	var loc *time.Location

	dt = &DateTime{}

	loc, err = stringToLocation(s, literal)
	if err == nil {
		// one case for "now" but not the only one
		dt.Time = time.Now().In(loc)

		if !nowIncludesFractionalSeconds {
			// set nanoseconds to 0
			dt.Time = dt.Time.Truncate(time.Second)
		}
		return

	} else if !IsValidDateTimeString(s, literal) {
		err = fmt.Errorf("Invalid Date-Time String")
		return
	}
	err = nil

	m := dtRegex.FindStringSubmatch(s)
	var names []string
	if m == nil {
		// possible second case for "now"
		m = dtRegexTZ.FindStringSubmatch(s)
		if m == nil {
			err = fmt.Errorf("Invalid Date-Time String")
			return
		}

		// current date and time with a specific time zone offset
		names = dtRegexTZ.SubexpNames()
		tzhours := subMatchByName("tzhours", m, names)
		tzminutes := subMatchByName("tzminutes", m, names)

		var offset int
		offset, err = stringsToOffset(tzhours, tzminutes)
		if err != nil {
			return dt, err
		}
		dt.Time = time.Now().In(time.FixedZone("", offset))

		if !nowIncludesFractionalSeconds {
			// set nanoseconds to 0
			dt.Time = dt.Time.Truncate(time.Second)
		}

		return
	}

	names = dtRegex.SubexpNames()

	year, _ := subMatchToInt("year", m, names)
	month, _ := subMatchToInt("month", m, names)
	day, _ := subMatchToInt("day", m, names)

	hour, _ := subMatchToInt("hour", m, names)
	minute, _ := subMatchToInt("minute", m, names)
	second, _ := subMatchToInt("second", m, names)

	secondsfraction := subMatchByName("secondsfraction", m, names)
	ns := 0
	if secondsfraction != "" {
		secondsfraction = str.PadRight(secondsfraction, 9, '0')
		ns, _ = str.StrToInt(secondsfraction, 10)
	}

	tzhours := subMatchByName("tzhours", m, names)
	tzminutes := subMatchByName("tzminutes", m, names)
	tzutc := subMatchByName("tzutc", m, names)

	if tzutc == "Z" {
		loc = time.UTC
	} else if tzhours == "" {
		loc = time.Local
	} else {
		offset, err := stringsToOffset(tzhours, tzminutes)
		if err != nil {
			return dt, err
		}
		loc = time.FixedZone("", offset)
	}

	dt.Time = time.Date(year, time.Month(month), day, hour, minute, second, ns, loc)

	return
}

/*
langur date-time literal syntax a subset of ISO 8601
extended format only (no mashing together without dashes and colons)

empty literal gives the current date and time

combined date/time uses either the letter T or a space between


langur literal accepted date syntax
YYYY-MM-DD      full year/month/day
YYYY from 0000 to 9999


langur literal accepted time syntax
hh          00:00 mm:ss implied
hh:mm       00 ss implied
hh:mm:ss
hh:mm:ss.fffffffff

seconds fraction may contain 1 to 9 digits

langur literal accepted time zone syntax
(none)      local time zone
Z           UTC
+hh         00 mm implied
-hh         00 mm implied
+hh:mm
-hh:mm

negative offset with ASCII hyphen-minus only, not the Unicode minus sign

no -00 offset (+00 okay)
For RFC 3339, a -00 offset is an "Unknown Local Offset Convention." For ISO 8601, it is invalid.

*/

// regex for date-time literal string verification
// no short form; colons and dashes required
var dtRegexDateLiteralString = "(?P<year>[0-9]{4})-(?P<month>1[0-2]|0[1-9])-(?P<day>3[01]|[12][0-9]|0[1-9])"

var dtRegexTimeLiteralString = `(?x)
	(?P<hour>2[0-3]|[01][0-9])
    	(?: :(?P<minute>[0-5][0-9])
        	(?: :(?P<second>[0-5][0-9]) (?: \. (?P<secondsfraction>[0-9]{1,9}))? )?
		)?`

var dtRegexTimeZoneLiteralString = `(?x)
    (?:(?P<tzutc>Z)
    	|
        (?P<tzhours>[+-] (?: 2[0-3]|[01][0-9]))
        (?: :(?P<tzminutes>[0-5][0-9]) )?
    )?`

var dtRegexLiteral = regexp.MustCompile(
	"^" + dtRegexDateLiteralString + "(?:(?: |T)" + dtRegexTimeLiteralString + dtRegexTimeZoneLiteralString + ")?$")

var dtRegexTZLiteral = regexp.MustCompile("^" + dtRegexTimeZoneLiteralString + "$")

// regex not just for date-time literals (less restrictive)
// colons and dashes optional
// allowing conversion of a string to a date-time from the short form
var dtRegexDateString = "(?P<year>[0-9]{4})-?(?P<month>1[0-2]|0[1-9])-?(?P<day>3[01]|[12][0-9]|0[1-9])"

var dtRegexTimeString = `(?x)
	(?P<hour>2[0-3]|[01][0-9])
    	(?: :?(?P<minute>[0-5][0-9])
        	(?: :?(?P<second>[0-5][0-9]) (?: \. (?P<secondsfraction>[0-9]{1,9}))? )?
		)?`

var dtRegexTimeZoneString = `(?x)
    (?:(?P<tzutc>Z)
    	|
        (?P<tzhours>[+-] (?: 2[0-3]|[01][0-9]))
        (?: :?(?P<tzminutes>[0-5][0-9]) )?
    )?`

var dtRegex = regexp.MustCompile(
	"^" + dtRegexDateString + "(?:(?: |T)" + dtRegexTimeString + dtRegexTimeZoneString + ")?$")

var dtRegexTZ = regexp.MustCompile("^" + dtRegexTimeZoneString + "$")

func IsValidDateTimeString(s string, literal bool) bool {
	_, err := stringToLocation(s, literal)
	if err == nil {
		return true
	}

	var m, names []string

	if literal {
		m = dtRegexLiteral.FindStringSubmatch(s)
	} else {
		m = dtRegex.FindStringSubmatch(s)
	}

	if m == nil {
		if literal {
			m = dtRegexTZLiteral.FindStringSubmatch(s)
		} else {
			m = dtRegexTZ.FindStringSubmatch(s)
		}
		if m == nil {
			return false
		}
		names = dtRegexTZ.SubexpNames()

	} else {
		names = dtRegex.SubexpNames()
	}

	// no -00 offset (+00 okay)
	// For RFC 3339, a -00 offset is an "Unknown Local Offset Convention." For ISO 8601, it is invalid.
	return !(subMatchByName("tzhours", m, names) == "-00" && subMatchByName("tzminutes", m, names) == "")
}

func subMatchByName(name string, subs, names []string) string {
	// NOTE: used internally and assuming slices match
	for i := range subs {
		if names[i] == name {
			return subs[i]
		}
	}
	bug("subMatchByName", fmt.Sprintf("Unknown submatch name %q", name))
	return "ERROR"
}

func subMatchToInt(name string, subs, names []string) (int, bool) {
	// NOTE: used internally and assuming slices match
	s := subMatchByName(name, subs, names)
	if s == "" {
		return 0, false
	}
	i, _ := str.StrToInt(s, 10)
	return i, true
}

func subMatchToInt64(name string, subs, names []string) (int64, bool) {
	// NOTE: used internally and assuming slices match
	s := subMatchByName(name, subs, names)
	if s == "" {
		return 0, false
	}
	i, _ := str.StrToInt64(s, 10)
	return i, true
}

func (dt *DateTime) UnixNano() int64 {
	return dt.Time.UnixNano()
}

func (dt *DateTime) ToLocation(locStr string) (*DateTime, error) {
	loc, err := stringToLocation(locStr, false)
	if err != nil {
		return dt, err
	}
	return &DateTime{Time: dt.Time.In(loc)}, nil
}

// It appears the Go time package expects the following numbers and formats in a format string.
// year: 2006 or 06
// month: 01 or 1
// month name: Jan or January
// month day: 02 or _2 or 2
//		02 is zero-padded, _2 is space-padded
// weekday name: Mon or Monday
// hour: 03 or 3 or 15
// minute: 04 or 4
// second: 05 or 5
// seconds fraction: 9
// AM/PM: PM or pm
// time zone offset: -07:00 or -0700 or -07
// time zone name: MST

// string formats for dates/times copied from github.com/unispaces/carbon
// const (
// 	DefaultFormat       = "2006-01-02 15:04:05"
// 	DateFormat          = "2006-01-02"
// 	FormattedDateFormat = "Jan 2, 2006"
// 	TimeFormat          = "15:04:05"
// 	HourMinuteFormat    = "15:04"
// 	HourFormat          = "15"
// 	DayDateTimeFormat   = "Mon, Aug 2, 2006 3:04 PM"
// 	CookieFormat        = "Monday, 02-Jan-2006 15:04:05 MST"
// 	RFC822Format        = "Mon, 02 Jan 06 15:04:05 -0700"
// 	RFC1036Format       = "Mon, 02 Jan 06 15:04:05 -0700"
// 	RFC2822Format       = "Mon, 02 Jan 2006 15:04:05 -0700"
// 	RFC3339Format       = "2006-01-02T15:04:05-07:00"
// 	RSSFormat           = "Mon, 02 Jan 2006 15:04:05 -0700"
// )

// copied from https://golang.org/pkg/time/
const (
	Format_ANSIC       = "Mon Jan _2 15:04:05 2006"
	Format_UnixDate    = "Mon Jan _2 15:04:05 MST 2006"
	Format_RubyDate    = "Mon Jan 02 15:04:05 -0700 2006"
	Format_RFC822      = "02 Jan 06 15:04 MST"
	Format_RFC822Z     = "02 Jan 06 15:04 -0700" // RFC822 with numeric zone
	Format_RFC850      = "Monday, 02-Jan-06 15:04:05 MST"
	Format_RFC1123     = "Mon, 02 Jan 2006 15:04:05 MST"
	Format_RFC1123Z    = "Mon, 02 Jan 2006 15:04:05 -0700" // RFC1123 with numeric zone
	Format_RFC3339     = "2006-01-02T15:04:05Z07:00"
	Format_RFC3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
	Format_Kitchen     = "3:04PM"
	// Handy time stamps.
	Format_Stamp      = "Jan _2 15:04:05"
	Format_StampMilli = "Jan _2 15:04:05.000"
	Format_StampMicro = "Jan _2 15:04:05.000000"
	Format_StampNano  = "Jan _2 15:04:05.000000000"
)

const secondsPerHour int = 3600

func offsetSecondsToString(offset int) string {
	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}
	hours := offset / secondsPerHour
	minutes := offset % secondsPerHour / 60

	return fmt.Sprintf("%s%02d:%02d", sign, hours, minutes)
}

func stringsToOffset(hours, minutes string) (int, error) {
	h, err := str.StrToInt(hours, 10)
	if err != nil {
		return 0, err
	}
	if minutes == "" {
		return h * secondsPerHour, nil
	}
	m, err := str.StrToInt(minutes, 10)
	if err != nil {
		return 0, err
	}
	sign := 1
	if h < 0 {
		sign = -1
	}
	return h*secondsPerHour + sign*m*secondsPerHour/60, nil
}

func stringToLocation(s string, literal bool) (*time.Location, error) {
	if s == "Z" {
		return time.UTC, nil
	} else if s == "" {
		return time.Local, nil
	}

	loc, err := time.LoadLocation(s)
	if err == nil {
		return loc, nil
	}

	var m, names []string

	if literal {
		m = dtRegexTZLiteral.FindStringSubmatch(s)
	} else {
		m = dtRegexTZ.FindStringSubmatch(s)
	}
	if m == nil {
		return nil, fmt.Errorf("Invalid time zone string")
	}
	names = dtRegexTZ.SubexpNames()

	tzhours := subMatchByName("tzhours", m, names)
	tzminutes := subMatchByName("tzminutes", m, names)

	if tzminutes == "" && tzhours == "-00" {
		// no -00 offset (+00 okay)
		// For RFC 3339, a -00 offset is an "Unknown Local Offset Convention." For ISO 8601, it is invalid.
		return nil, fmt.Errorf("Invalid time zone string; no -00 offset without minutes")
	}

	offset, err := stringsToOffset(tzhours, tzminutes)
	if err != nil {
		return nil, err
	}
	return time.FixedZone("", offset), nil
}

func ToDateTime(obj Object, formatOrLocStr string) (*DateTime, error) {
	switch obj := obj.(type) {
	case *Number:
		ns, err := obj.ToInt64()
		if err != nil {
			return nil, fmt.Errorf("Error converting to integer for date-time")
		}
		return NewDateTimeFromNanoseconds(ns), nil

	case *String:
		if formatOrLocStr == "" {
			return NewDateTimeFromString(obj.String(), true, false)
		}
		return NewDateTimeFromStringFormat(obj.String(), formatOrLocStr)

	case *DateTime:
		return obj.ToLocation(formatOrLocStr)

	case *Hash:
		return obj.ToDateTime()

	}
	return nil, fmt.Errorf("Unknown type (%s) for date-time conversion", obj.TypeString())
}
