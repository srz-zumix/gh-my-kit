package activity

import (
	"errors"
	"fmt"
	"time"

	"github.com/srz-zumix/go-gh-extension/pkg/parser"
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

// ResolvePeriod determines the (since, until) window from --period and
// --since/--until inputs. --period is mutually exclusive with --since/--until.
// When none are given, the window defaults to the last 30 days.
func ResolvePeriod(period, since, until string, now time.Time) (time.Time, time.Time, error) {
	if period != "" && (since != "" || until != "") {
		return time.Time{}, time.Time{}, errors.New("--period cannot be used together with --since or --until")
	}
	if period != "" {
		if parser.IsFiscalPeriod(period) {
			return parser.ParseFiscalPeriod(period)
		}
		duration, err := parser.ParsePeriod(period)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
		// parser.ParsePeriod returns whole-day multiples; convert back to
		// calendar days and use AddDate so the window keeps the same
		// wall-clock time across daylight-saving transitions.
		days := int(duration / (24 * time.Hour))
		return now.AddDate(0, 0, -days), now, nil
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
