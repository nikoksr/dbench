package duration

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Duration time.Duration

func (d Duration) Value() (driver.Value, error) {
	return int64(d), nil
}

func (d *Duration) Scan(src interface{}) error {
	switch v := src.(type) {
	case int64:
		*d = Duration(v)
		return nil
	default:
		return fmt.Errorf("cannot scan %T", src)
	}
}

func (d Duration) String() string {
	return time.Duration(d).String()
}
