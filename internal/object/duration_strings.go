// langur/object/duration_strings.go

package object

import (
	"langur/regexp"
)

/*
almost ISO 8601 durations
As of 0.13, the "P" is optional and spaces are allowed between each number and after the "T".
As of 0.13.8, lowercase letters may be used (except for the "T").

The default conversion to string is in ISO 8601 format.

If given weeks, they are converted to days.

P[n]W
(weeks)

or ...
P[n]Y[n]M[n]DT[n]H[n]M[n[.n]]S

must include at least one of...

years, months, days
hours, minutes, seconds

allowing 1 to 9 fractional digits on seconds for langur (9 digits representing nanoseconds)
as of langur 0.12.6, following the lead of others in using fractional seconds on durations...
12 digits: https://www.ibm.com/docs/en/i/7.2?topic=types-xsduration
 6 digits: https://docs.oracle.com/en/database/oracle/oracle-database/21/adjsn/iso-8601-date-time-and-duration-support.html
*/

var dtPatternDurationDaysString = `(?x:
	(?:(?P<years>[0-9]+)[Yy])?\ ?
	(?:(?P<months>[0-9]+)[Mm])?\ ?
	(?:(?P<days>[0-9]+)[Dd])?\ ?
)`

var dtPatternDurationTimeString = `(?x:T\ ?
	(?:(?P<hours>[0-9]+)[Hh])?\ ?
	(?:(?P<minutes>[0-9]+)[Mm])?\ ?
	(?:(?P<seconds>[0-9]+) (?:\.(?P<secondsfraction>[0-9]{1,9}))? [Ss])?
)`

var dtPatternDurationWeeksString = `(?P<weeks>[0-9]+)[Ww]`

var dtPatternDuration = regexp.MustCompile(
	"^P?(?:" + dtPatternDurationDaysString + ")?(?:" + dtPatternDurationTimeString + ")?$|^P?" +
		dtPatternDurationWeeksString + "$")

// A "month" and a "year" are hard to define in terms of nanoseconds.
const (
	nsPerYear int64 = 31557600000000000 // 365.25 days
	nsPerMon        = 2629800000000000  // 30.4375 days
	nsPerWeek       = 604800000000000
	nsPerDay        = 86400000000000
	nsPerHour       = 3600000000000
	nsPerMin        = 60000000000
	nsPerSec        = 1000000000
)

var (
	nsPerYearN *Number = NumberFromInt64(nsPerYear)
	nsPerMonN          = NumberFromInt64(nsPerMon)
	nsPerWeekN         = NumberFromInt64(nsPerWeek)
	nsPerDayN          = NumberFromInt64(nsPerDay)
	nsPerHourN         = NumberFromInt64(nsPerHour)
	nsPerMinN          = NumberFromInt64(nsPerMin)
	nsPerSecN          = NumberFromInt64(nsPerSec)
)

func IsValidDurationString(s string) bool {
	if s == "" {
		// allowing zero-length string to represent a 0 duration
		return true
	}

	m := dtPatternDuration.FindStringSubmatch(s)
	if m == nil {
		return false
	}
	names := dtPatternDuration.SubexpNames()

	// must have at least one submatch
	_, ok := subMatchToInt64("years", m, names)
	if ok {
		return true
	}
	_, ok = subMatchToInt64("months", m, names)
	if ok {
		return true
	}
	_, ok = subMatchToInt64("weeks", m, names)
	if ok {
		return true
	}
	_, ok = subMatchToInt64("days", m, names)
	if ok {
		return true
	}
	_, ok = subMatchToInt64("hours", m, names)
	if ok {
		return true
	}
	_, ok = subMatchToInt64("minutes", m, names)
	if ok {
		return true
	}
	_, ok = subMatchToInt64("seconds", m, names)
	if ok {
		return true
	}

	return false
}
