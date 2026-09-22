package activity

import (
	"testing"
	"time"
)

func TestParseFiscalPeriod(t *testing.T) {
	tests := []struct {
		name      string
		period    string
		wantSince string
		wantUntil string
		wantErr   bool
	}{
		{name: "full fiscal year, 2-digit", period: "FY26", wantSince: "2026-04-01", wantUntil: "2027-03-31"},
		{name: "full fiscal year, 4-digit", period: "FY2026", wantSince: "2026-04-01", wantUntil: "2027-03-31"},
		{name: "lowercase", period: "fy26", wantSince: "2026-04-01", wantUntil: "2027-03-31"},
		{name: "first half", period: "FY26H1", wantSince: "2026-04-01", wantUntil: "2026-09-30"},
		{name: "second half", period: "FY26H2", wantSince: "2026-10-01", wantUntil: "2027-03-31"},
		{name: "first quarter", period: "FY26Q1", wantSince: "2026-04-01", wantUntil: "2026-06-30"},
		{name: "second quarter", period: "FY26Q2", wantSince: "2026-07-01", wantUntil: "2026-09-30"},
		{name: "third quarter", period: "FY26Q3", wantSince: "2026-10-01", wantUntil: "2026-12-31"},
		{name: "fourth quarter", period: "FY26Q4", wantSince: "2027-01-01", wantUntil: "2027-03-31"},
		{name: "invalid suffix", period: "FY26X1", wantErr: true},
		{name: "not a fiscal period", period: "30d", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			since, until, err := ParseFiscalPeriod(tt.period)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseFiscalPeriod(%q) expected an error, got none", tt.period)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseFiscalPeriod(%q) unexpected error: %v", tt.period, err)
			}
			if got := since.Format(dateLayout); got != tt.wantSince {
				t.Errorf("since = %q, want %q", got, tt.wantSince)
			}
			if got := until.Format(dateLayout); got != tt.wantUntil {
				t.Errorf("until = %q, want %q", got, tt.wantUntil)
			}
		})
	}
}

func TestResolvePeriodFiscal(t *testing.T) {
	now := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	since, until, err := ResolvePeriod("FY26H1", "", "", now)
	if err != nil {
		t.Fatalf("ResolvePeriod returned unexpected error: %v", err)
	}
	if got := since.Format(dateLayout); got != "2026-04-01" {
		t.Errorf("since = %q, want 2026-04-01", got)
	}
	if got := until.Format(dateLayout); got != "2026-09-30" {
		t.Errorf("until = %q, want 2026-09-30", got)
	}
}

func TestResolvePeriodUntilDateOnlyEndOfDay(t *testing.T) {
	now := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	since, until, err := ResolvePeriod("", "2024-01-01", "2024-03-31", now)
	if err != nil {
		t.Fatalf("ResolvePeriod returned unexpected error: %v", err)
	}
	// Date-only since stays at the start of the day.
	wantSince := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	if !since.Equal(wantSince) {
		t.Errorf("since = %v, want %v", since, wantSince)
	}
	// Date-only until is extended to the last instant of the day.
	wantUntil := time.Date(2024, 3, 31, 23, 59, 59, int(time.Second-time.Nanosecond), time.UTC)
	if !until.Equal(wantUntil) {
		t.Errorf("until = %v, want %v", until, wantUntil)
	}
	// The formatted search range keeps the same calendar day.
	if got := until.Format(dateLayout); got != "2024-03-31" {
		t.Errorf("until date = %q, want 2024-03-31", got)
	}
}

func TestResolvePeriodUntilRFC3339NotExtended(t *testing.T) {
	now := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	_, until, err := ResolvePeriod("", "", "2024-03-31T00:00:00Z", now)
	if err != nil {
		t.Fatalf("ResolvePeriod returned unexpected error: %v", err)
	}
	// An explicit RFC3339 instant is used as-is, not extended to end of day.
	wantUntil := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)
	if !until.Equal(wantUntil) {
		t.Errorf("until = %v, want %v", until, wantUntil)
	}
}

func TestResolvePeriodSameDayDateOnly(t *testing.T) {
	now := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	since, until, err := ResolvePeriod("", "2024-03-31", "2024-03-31", now)
	if err != nil {
		t.Fatalf("ResolvePeriod returned unexpected error for same-day range: %v", err)
	}
	if !since.Before(until) {
		t.Errorf("expected since %v before until %v", since, until)
	}
}

func TestResolvePeriodNoUntilKeepsNow(t *testing.T) {
	now := time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)
	_, until, err := ResolvePeriod("", "2024-01-01", "", now)
	if err != nil {
		t.Fatalf("ResolvePeriod returned unexpected error: %v", err)
	}
	if !until.Equal(now) {
		t.Errorf("until = %v, want %v", until, now)
	}
}
