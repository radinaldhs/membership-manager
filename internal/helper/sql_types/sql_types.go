package sqltypes

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strings"
	"time"
)

func createDateTimeWithTimezoneErr(msg string) error {
	return fmt.Errorf("can not parse value to time.Time: %s", msg)
}

// DateTimeWithTimezone parse DATETIME data type from MySQL (or any RDBMS that has similar type)
// but with additional timezone information, example: '2025-03-22 13:26:53 UTC'
type DateTimeWithTimezone time.Time

func (dt DateTimeWithTimezone) Time() time.Time {
	return time.Time(dt)
}

// Scan implements [database/sql.Scanner] interface
func (dt *DateTimeWithTimezone) Scan(src any) error {
	if src == nil {
		errMsg := "nil value"
		return createDateTimeWithTimezoneErr(errMsg)
	}

	b, ok := src.([]byte)
	if !ok {
		errMsg := fmt.Sprintf("expecting value to be of []byte type, got %T instead", src)
		return createDateTimeWithTimezoneErr(errMsg)
	}

	str := string(b)
	t, err := time.Parse("2006-01-02 15:04:05 MST", str)
	if err != nil {
		// Sometimes, the timezone is a location (example: Asia/Bangkok),
		// so we gotta be clever here
		split := strings.Split(str, " ")
		if len(split) < 3 {
			createDateTimeWithTimezoneErr(fmt.Sprintf("can not parse %s to time.Time", str))
		}

		// On index 2 is where the location string is
		loc, err := time.LoadLocation(split[2])
		if err != nil {
			createDateTimeWithTimezoneErr(err.Error())
		}

		str = strings.Join(split[0:2], " ")

		t, err = time.ParseInLocation("2006-01-02 15:04:05", str, loc)
		if err != nil {
			return createDateTimeWithTimezoneErr(err.Error())
		}
	}

	*dt = DateTimeWithTimezone(t)

	return nil
}

// Timestamp parses time in RFC3339 format.
type Timestamp struct {
	t       time.Time
	isValid bool
}

func NewTimestamp(t time.Time) Timestamp {
	return Timestamp{t: t, isValid: true}
}

func (tms Timestamp) SetTime(t time.Time) Timestamp {
	tms.t = t
	return tms
}

func (tms Timestamp) SetValid(b bool) Timestamp {
	tms.isValid = b
	return tms
}

func (tms Timestamp) Time() time.Time {
	return tms.t
}

func (tms Timestamp) Value() (driver.Value, error) {
	if !tms.isValid {
		return nil, nil
	}

	return tms.t.Format(time.RFC3339), nil
}

func (tms *Timestamp) Scan(src any) error {
	if src == nil {
		return nil
	}

	str, ok := src.(string)
	if !ok {
		return errors.New("can not scan value error: not a string")
	}

	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return fmt.Errorf("can not scan value error: can not parse '%s' to RFC3339", str)
	}

	tms.t = t
	tms.isValid = true

	return nil
}
