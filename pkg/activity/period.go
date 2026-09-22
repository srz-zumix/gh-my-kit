package activity

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// dateLayouts are the accepted formats for --since/--until.
var dateLayouts = []string{time.RFC3339, "2006-01-02"}

// ParseTimeFlag parses a --since/--until value in RFC3339 or YYYY-MM-DD format.
func ParseTimeFlag(value string) (time.Time, error) {
	t, _, err := parseTimeFlag(value)
	return t, err
}

// parseTimeFlag parses a --since/--until value and reports whether it matched
// the date-only (YYYY-MM-DD) layout, which callers use to treat the value as a
// whole calendar day rather than an instant at midnight.
func parseTimeFlag(value string) (time.Time, bool, error) {
	var lastErr error
	for _, layout := range dateLayouts {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t, layout == "2006-01-02", nil
		}
		lastErr = err
	}
	return time.Time{}, false, fmt.Errorf("invalid time %q: expected RFC3339 or YYYY-MM-DD: %w", value, lastErr)
}

// ParsePeriod parses a relative period such as "7d", "3w", "6m" or "1y" into a
// (since, until) window ending at now. Supported units are d(day), w(week),
// m(month, treated as 30 days) and y(year, treated as 365 days).
func ParsePeriod(period string, now time.Time) (time.Time, time.Time, error) {
	if len(period) < 2 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period %q: expected format <N>d|w|m|y", period)
	}
	unit := period[len(period)-1]
	n, err := strconv.Atoi(period[:len(period)-1])
	if err != nil || n <= 0 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period %q: expected format <N>d|w|m|y", period)
	}

	var days int
	switch unit {
	case 'd':
		days = n
	case 'w':
		days = n * 7
	case 'm':
		days = n * 30
	case 'y':
		days = n * 365
	default:
		return time.Time{}, time.Time{}, fmt.Errorf("invalid period %q: expected format <N>d|w|m|y", period)
	}

	return now.AddDate(0, 0, -days), now, nil
}

// fiscalYearStartMonth is the first month of a fiscal year (April, matching
// the Japanese fiscal year convention).
const fiscalYearStartMonth = time.April

// fiscalPeriodPattern matches fiscal year period expressions such as "FY26",
// "FY2026", "FY26H1" and "FY26Q3". The year is the calendar year in which the
// fiscal year starts (FY26 starts April 2026), and H/Q select a half or
// quarter within that fiscal year.
var fiscalPeriodPattern = regexp.MustCompile(`(?i)^FY(\d{2}|\d{4})(H[12]|Q[1-4])?$`)

// fiscalPeriodOffsets maps an H/Q suffix to the number of months after the
// fiscal year start and the span of the period in months.
var fiscalPeriodOffsets = map[string][2]int{
	"":   {0, 12},
	"H1": {0, 6},
	"H2": {6, 6},
	"Q1": {0, 3},
	"Q2": {3, 3},
	"Q3": {6, 3},
	"Q4": {9, 3},
}

// ParseFiscalPeriod parses a fiscal year period such as "FY26" (the full
// fiscal year), "FY26H1"/"FY26H2" (half), or "FY26Q1".."FY26Q4" (quarter)
// into a (since, until) window. The fiscal year starts in April, so FY26
// spans 2026-04-01 to 2027-03-31.
func ParseFiscalPeriod(period string) (time.Time, time.Time, error) {
	m := fiscalPeriodPattern.FindStringSubmatch(period)
	if m == nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid fiscal period %q: expected format FY<YY>[H1|H2|Q1..Q4]", period)
	}

	year, err := strconv.Atoi(m[1])
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid fiscal period %q: %w", period, err)
	}
	if len(m[1]) == 2 {
		year += 2000
	}

	offsets := fiscalPeriodOffsets[strings.ToUpper(m[2])]
	fyStart := time.Date(year, fiscalYearStartMonth, 1, 0, 0, 0, 0, time.Local)
	since := fyStart.AddDate(0, offsets[0], 0)
	until := since.AddDate(0, offsets[1], 0).Add(-time.Nanosecond)
	return since, until, nil
}

// ResolvePeriod determines the (since, until) window from --period and
// --since/--until inputs. --period is mutually exclusive with --since/--until.
// When none are given, the window defaults to the last 30 days.
func ResolvePeriod(period, since, until string, now time.Time) (time.Time, time.Time, error) {
	if period != "" && (since != "" || until != "") {
		return time.Time{}, time.Time{}, errors.New("--period cannot be used together with --since or --until")
	}
	if period != "" {
		if fiscalPeriodPattern.MatchString(period) {
			return ParseFiscalPeriod(period)
		}
		return ParsePeriod(period, now)
	}

	end := now
	if until != "" {
		t, dateOnly, err := parseTimeFlag(until)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		if dateOnly {
			// A date-only upper bound is inclusive of the whole calendar day,
			// so extend it to the last instant of that day.
			t = t.AddDate(0, 0, 1).Add(-time.Nanosecond)
		}
		end = t
	}

	// The default 30-day window is relative to the resolved end, so an
	// explicit --until in the past is not rejected just because it predates
	// "now minus 30 days".
	start := end.AddDate(0, 0, -30)
	if since != "" {
		t, err := ParseTimeFlag(since)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		start = t
	}
	if !start.Before(end) {
		return time.Time{}, time.Time{}, errors.New("--since must be before --until")
	}
	return start, end, nil
}
