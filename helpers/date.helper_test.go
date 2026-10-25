package sdk_helper

import (
	"testing"
	"time"
)

func mustParseDate(t *testing.T, value string) *Date {
	t.Helper()

	d, err := ParseDate("2006-01-02", value)
	if err != nil {
		t.Fatalf("parse %q: %v", value, err)
	}

	return d
}

func TestDateFields(t *testing.T) {
	d := mustParseDate(t, "2024-02-29")

	if d.Year != 2024 || d.Month != 2 || d.Day != 29 {
		t.Fatalf("unexpected fields %+v", d)
	}

	if d.Weekday != time.Thursday {
		t.Fatalf("expected Thursday, got %v", d.Weekday)
	}

	if d.IsWeekend() {
		t.Fatal("Thursday is not a weekend")
	}

	if got := d.Format("2006/01/02"); got != "2024/02/29" {
		t.Fatalf("Format got %q", got)
	}

	if d.UnixMilli() != d.Unix()*1000 {
		t.Fatal("UnixMilli consistency")
	}

	if d.UnixNano()/1e6 != d.UnixMilli() {
		t.Fatal("UnixNano consistency")
	}
}

func TestDateWeekend(t *testing.T) {
	sat := mustParseDate(t, "2024-03-02")
	sun := mustParseDate(t, "2024-03-03")

	if !sat.IsWeekend() || !sun.IsWeekend() {
		t.Fatal("expected weekend detection")
	}
}

func TestDaysInMonthLeap(t *testing.T) {
	feb2024 := mustParseDate(t, "2024-02-01")
	if feb2024.DaysInMonth() != 29 {
		t.Fatalf("leap year Feb should be 29, got %d", feb2024.DaysInMonth())
	}

	feb2023 := mustParseDate(t, "2023-02-01")
	if feb2023.DaysInMonth() != 28 {
		t.Fatalf("non-leap Feb should be 28, got %d", feb2023.DaysInMonth())
	}

	dec := mustParseDate(t, "2024-12-15")
	if dec.DaysInMonth() != 31 {
		t.Fatalf("Dec should be 31, got %d", dec.DaysInMonth())
	}
}

func TestStartEndOfDay(t *testing.T) {
	d := mustParseDate(t, "2024-06-15")

	sod := d.StartOfDay()
	if sod.Hour != 0 || sod.Minute != 0 || sod.Second != 0 || sod.Day != 15 {
		t.Fatalf("StartOfDay %+v", sod)
	}

	eod := d.EndOfDay()
	if eod.Hour != 23 || eod.Minute != 59 || eod.Second != 59 || eod.Millisecond != 999 {
		t.Fatalf("EndOfDay %+v", eod)
	}
}

func TestDateAdd(t *testing.T) {
	d := mustParseDate(t, "2024-12-31")

	n := d.Add(24 * time.Hour)
	if n.Year != 2025 || n.Month != 1 || n.Day != 1 {
		t.Fatalf("Add crossed year boundary incorrectly: %+v", n)
	}
}

func TestAge(t *testing.T) {
	now := time.Now()

	if got := Age(now.Year(), int(now.Month()), now.Day()); got != 0 {
		t.Fatalf("age born today should be 0, got %d", got)
	}

	if got := Age(2000, 1, 1); got != now.Year()-2000 {
		t.Fatalf("expected %d got %d", now.Year()-2000, got)
	}

	futureYear := now.Year() + 10
	if got := Age(futureYear, 1, 1); got != -10 {
		t.Fatalf("expected -10 for future birth year, got %d", got)
	}
}

func TestParseHelpers(t *testing.T) {
	if _, err := ParseDate("garbage-layout", "2024-01-01"); err == nil {
		t.Fatal("expected parse error")
	}

	utc, err := ParseInLocation(time.RFC3339, "2024-01-01T12:00:00Z", time.UTC)
	if err != nil {
		t.Fatal(err)
	}

	if utc.Hour != 12 || utc.Native.Location() != time.UTC {
		t.Fatalf("unexpected %+v", utc)
	}
}

func TestNowHelpers(t *testing.T) {
	if Now().Year < 2024 {
		t.Fatal("Now looks stale")
	}

	u := NowUTC()
	if u.Native.Location() != time.UTC {
		t.Fatal("NowUTC must be UTC")
	}

	if len(u.Timestamp) == 0 {
		t.Fatal("timestamp string missing")
	}

	if u.ISOWeek[1] < 1 || u.ISOWeek[1] > 53 {
		t.Fatalf("ISOWeek out of range: %v", u.ISOWeek)
	}
}
