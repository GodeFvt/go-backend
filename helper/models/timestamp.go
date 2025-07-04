package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"4d63.com/tz"
)

const (
	TimestampLayout = "2006-01-02 15:04:05"
)

var (
	Location, _ = tz.LoadLocation("Asia/Bangkok")
)

type Timestamp time.Time

/*
------------------------
Timestamp Function
------------------------
*/
func NewTimestampFromNow() Timestamp {
	return NewTimestampFromTime(time.Now())
}

func NewTimestampFromString(dateString string) Timestamp {
	if dateString == "" {
		return Timestamp(time.Time{})
	}
	d, err := time.ParseInLocation(TimestampLayout, dateString, Location)
	if err != nil {
		panic(err)
	}
	return Timestamp(d.In(Location))
}

func NewTimestampFromTime(t time.Time) Timestamp {
	d, err := time.ParseInLocation(TimestampLayout, t.In(Location).Format(TimestampLayout), Location)
	if err != nil {
		panic(err)
	}

	return Timestamp(d.In(Location))
}

func ParseTimestampFromString(dateString string) (Timestamp, error) {
	if dateString == "" {
		return Timestamp(time.Time{}), nil
	}
	d, err := time.ParseInLocation(TimestampLayout, dateString, Location)
	if err != nil {
		return Timestamp(time.Time{}), err
	}
	return Timestamp(d.In(Location)), nil
}

func ParseTimestampFromTime(t time.Time) (Timestamp, error) {
	d, err := time.ParseInLocation(TimestampLayout, t.In(Location).Format(TimestampLayout), Location)
	if err != nil {
		return Timestamp(time.Time{}), err
	}

	return Timestamp(d.In(Location)), nil
}

func (t Timestamp) ToUnix() int64 {
	tt := time.Time(t)
	return tt.Unix()
}

func (j *Timestamp) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")

	// Define timestamp formats (including microseconds & milliseconds)
	layouts := []string{
		"2006-01-02 15:04:05.999999", // "2025-04-03 12:05:12.510131" (microseconds)
		"2006-01-02 15:04:05.999",    // "2025-04-03 12:05:12.510" (milliseconds)
		"2006-01-02 15:04:05",        // "2025-04-03 12:05:12"
		"2006-01-02T15:04:05",        // "2024-06-25T14:30:00"
		"2006-01-02",                 // "2024-06-25"
		"02 Jan 2006 15:04:05 MST",   // "25 Jun 2024 14:30:00 UTC"
		"02 Jan 2006",                // "25 Jun 2024"
		time.RFC3339,                 // "2025-04-03T12:05:12Z"
		time.RFC3339Nano,             // "2025-04-03T12:05:12.510131Z"
		time.RFC1123,                 // "Mon, 25 Jun 2024 14:30:00 UTC"
		time.RFC1123Z,                // "Mon, 25 Jun 2024 14:30:00 +0000"
		time.RFC850,                  // "Monday, 25-Jun-24 14:30:00 UTC"
		time.ANSIC,                   // "Mon Jan _2 15:04:05 2006"
		time.UnixDate,                // "Mon Jan 2 15:04:05 MST 2006"
		time.RubyDate,                // "Mon Jan 02 15:04:05 -0700 2006"
	}

	var t time.Time
	var err error

	// Try each format until one succeeds
	for _, layout := range layouts {
		t, err = time.Parse(layout, s)
		if err == nil {
			*j = Timestamp(t)
			return nil
		}
	}

	return fmt.Errorf("invalid timestamp format: %s", s)
}

func (j Timestamp) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Format(TimestampLayout))
}

func (j Timestamp) Format(s string) string {
	t := time.Time(j)
	return t.Format(s)
}

func (j Timestamp) YearDay() int {
	t := time.Time(j)
	return t.YearDay()
}

func (j Timestamp) String() string {
	return j.Format(TimestampLayout)
}
func (j *Timestamp) Interface() interface{} {
	if j == nil {
		return nil
	}

	return j.Format(TimestampLayout)
}
func (j *Timestamp) GetBSON() (interface{}, error) {
	if j == nil {
		return nil, nil
	}
	if (*j) == Timestamp(time.Time{}) {
		return nil, nil
	}
	loc, _ := tz.LoadLocation("Asia/Bangkok")
	t := time.Time(*j)
	d, err := time.ParseInLocation(TimestampLayout, t.Format(TimestampLayout), loc)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (j *Timestamp) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		*j = Timestamp(v)
	case string:
		parsedTime, err := time.ParseInLocation(TimestampLayout, v, Location)
		if err != nil {
			return err
		}
		*j = Timestamp(parsedTime)
	case []byte:
		parsedTime, err := time.ParseInLocation(TimestampLayout, string(v), Location)
		if err != nil {
			return err
		}
		*j = Timestamp(parsedTime)
	default:
		return fmt.Errorf("cannot scan type %T into Timestamp", value)
	}

	return nil
}

func (j Timestamp) Value() (driver.Value, error) {
	if j == (Timestamp{}) {
		return nil, nil
	}
	return j.String(), nil
}

func (j Timestamp) ValueOrZero() string {
	if j == (Timestamp{}) {
		return ""
	}
	return j.String()
}

func (j Timestamp) ToTime() time.Time {
	return time.Time(j)
}

func (j Timestamp) ToPointer() *Timestamp {
	return &j
}
