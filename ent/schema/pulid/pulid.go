// Package pulid implements the pulid type.
// A pulid is an identifier that is a two-byte prefixed ULIDs, with the first two bytes encoding the type of the entity.
package pulid

import (
	"database/sql/driver"
	"fmt"
	"go.jetpack.io/typeid"
)

// ID implements a PULID - a prefixed ULID.
type ID string

// MustNew returns a new PULID for time.Now() given a prefix. This uses the default entropy source.
func MustNew(prefix string) ID {
	return ID(typeid.Must(typeid.New(prefix)).String())
}

// Scan implements the Scanner interface.
func (u *ID) Scan(src any) error {
	if src == nil {
		return fmt.Errorf("pulid: expected a value")
	}
	switch src := src.(type) {
	case string:
		*u = ID(src)
	case ID:
		*u = src
	default:
		return fmt.Errorf("pulid: unexpected type, %T", src)
	}
	return nil
}

// Value implements the driver Valuer interface.
func (u ID) Value() (driver.Value, error) {
	return string(u), nil
}

// String implements the fmt.Stringer interface.
func (u ID) String() string {
	return string(u)
}

// TypeID returns the type id of the PULID.
func (u ID) TypeID() (typeid.TypeID, error) {
	return typeid.FromString(string(u))
}
