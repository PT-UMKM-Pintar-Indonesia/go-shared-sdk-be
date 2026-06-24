package sdk_helper

import (
	"time"
)

type Date struct {
	Native      time.Time
	Timestamp   string
	Year        int
	Month       int
	Day         int
	Hour        int
	Minute      int
	Second      int
	Millisecond int
	Nanosecond  int
	Weekday     time.Weekday
	ISOWeek     [2]int
}

func newDate(t time.Time) *Date {
	isoYear, isoWeek := t.ISOWeek()
	return &Date{
		Native:      t,
		Timestamp:   t.Format(time.RFC3339Nano),
		Year:        t.Year(),
		Month:       int(t.Month()),
		Day:         t.Day(),
		Hour:        t.Hour(),
		Minute:      t.Minute(),
		Second:      t.Second(),
		Millisecond: t.Nanosecond() / int(time.Millisecond),
		Nanosecond:  t.Nanosecond(),
		Weekday:     t.Weekday(),
		ISOWeek:     [2]int{isoYear, isoWeek},
	}
}

func Now() *Date {
	return newDate(time.Now())
}

func NowUTC() *Date {
	return newDate(time.Now().UTC())
}

func (h *Date) Format(layout string) string {
	return h.Native.Format(layout)
}

func (h *Date) Unix() int64 {
	return h.Native.Unix()
}

func (h *Date) UnixMilli() int64 {
	return h.Native.UnixMilli()
}

func (h *Date) UnixNano() int64 {
	return h.Native.UnixNano()
}

func (h *Date) Add(duration time.Duration) *Date {
	return newDate(h.Native.Add(duration))
}

func (h *Date) StartOfDay() *Date {
	t := h.Native
	return newDate(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()))
}

func (h *Date) EndOfDay() *Date {
	t := h.Native
	// Menggunakan 999,999,999 nanodetik untuk presisi akhir hari
	return newDate(time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location()))
}

func (h *Date) IsWeekend() bool {
	return h.Weekday == time.Saturday || h.Weekday == time.Sunday
}

func (h *Date) DaysInMonth() int {
	// Trik: Tanggal 0 dari bulan berikutnya adalah hari terakhir bulan ini
	return time.Date(h.Year, time.Month(h.Month+1), 0, 0, 0, 0, 0, h.Native.Location()).Day()
}

// Age dihitung berdasarkan perbandingan tanggal absolut
func Age(birthYear, birthMonth, birthDay int) int {
	now := time.Now()
	age := now.Year() - birthYear

	// Cek apakah ulang tahun sudah lewat di tahun ini
	if now.Month() < time.Month(birthMonth) ||
		(now.Month() == time.Month(birthMonth) && now.Day() < birthDay) {
		age--
	}
	return age
}

func ParseDate(layout, value string) (*Date, error) {
	t, err := time.Parse(layout, value)
	if err != nil {
		return nil, err
	}
	return newDate(t), nil
}

func ParseInLocation(layout, value string, loc *time.Location) (*Date, error) {
	t, err := time.ParseInLocation(layout, value, loc)
	if err != nil {
		return nil, err
	}
	return newDate(t), nil
}
