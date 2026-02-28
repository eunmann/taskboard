package ids

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
)

// ErrUnsupportedScanType is returned when Scan receives an unsupported type.
var ErrUnsupportedScanType = errors.New("unsupported scan type")

// baseID wraps a UUID and provides common operations for all ID types.
type baseID struct {
	uuid UUID
}

func (b baseID) String() string {
	return b.uuid.String()
}

func (b baseID) IsZero() bool {
	return b.uuid.IsZero()
}

func (b baseID) UUID() UUID {
	return b.uuid
}

func (b baseID) Value() (driver.Value, error) {
	if b.uuid.IsZero() {
		return nil, nil //nolint:nilnil // SQL NULL convention for zero IDs
	}

	return b.uuid.String(), nil
}

func (b *baseID) scanString(src any) error {
	switch v := src.(type) {
	case string:
		u, err := ParseUUID(v)
		if err != nil {
			return fmt.Errorf("parse UUID: %w", err)
		}

		b.uuid = u
	case []byte:
		u, err := ParseUUID(string(v))
		if err != nil {
			return fmt.Errorf("parse UUID bytes: %w", err)
		}

		b.uuid = u
	case nil:
		b.uuid = UUID{}
	default:
		return fmt.Errorf("%w: %T", ErrUnsupportedScanType, src)
	}

	return nil
}

func (b baseID) marshalJSON() ([]byte, error) {
	data, err := json.Marshal(b.uuid.String())
	if err != nil {
		return nil, fmt.Errorf("marshal ID: %w", err)
	}

	return data, nil
}

func (b *baseID) unmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("unmarshal ID: %w", err)
	}

	u, err := ParseUUID(s)
	if err != nil {
		return fmt.Errorf("parse ID: %w", err)
	}

	b.uuid = u

	return nil
}

// UserID identifies a user.
type UserID struct{ baseID }

func NewUserID() UserID { return UserID{baseID{uuid: NewV7()}} }
func ParseUserID(s string) (UserID, error) {
	u, err := ParseUUID(s)

	return UserID{baseID{uuid: u}}, err
}

func MustParseUserID(s string) UserID {
	u, err := ParseUserID(s)
	if err != nil {
		panic(err)
	}

	return u
}
func UserIDFromUUID(u UUID) UserID              { return UserID{baseID{uuid: u}} }
func (id UserID) MarshalJSON() ([]byte, error)  { return id.marshalJSON() }
func (id *UserID) UnmarshalJSON(b []byte) error { return id.unmarshalJSON(b) }
func (id UserID) Value() (driver.Value, error)  { return id.baseID.Value() }
func (id *UserID) Scan(src any) error           { return id.scanString(src) }
func (id UserID) Compare(other UserID) int      { return id.uuid.Compare(other.uuid) }
