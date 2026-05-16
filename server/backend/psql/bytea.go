package psql

/* Some time between time immemorial (i.e. ca. 2016) and now, the pq library
 * changed its handling of byte slices to treat nil slices as SQL NULLs rather
 * than empty arrays. This broke legacy code in this package that assumed that
 * nil byte slices are equivalent to empty ones. The two types introduced in
 * this file allow clearly distinguishing nullable vs. non-nullable byte
 * strings. */

import (
	"bytes"
	"database/sql/driver"
	"fmt"
)

type ByteAOrNull struct {
	v []byte
}

type ByteANonNull struct {
	v []byte
}

func NewByteAOrNull(data []byte) ByteAOrNull {
	if len(data) == 0 {
		data = nil
	}
	return ByteAOrNull{data}
}

func EmptyByteAOrNull() ByteAOrNull {
	return ByteAOrNull{}
}

func NewByteANonNull(data []byte) ByteANonNull {
	if len(data) == 0 {
		panic("trying to put null into non-nullable byte array")
	}
	return ByteANonNull{data}
}

func (b *ByteAOrNull) Scan(src interface{}) error {
	switch src.(type) {
	case []byte:
		result := src.([]byte)
		if len(result) != 0 {
			result = bytes.Clone(result)
		} else {
			result = nil
		}
		b.v = result
		return nil
	case nil:
		b.v = nil
		return nil
	}
	return fmt.Errorf("cannot convert %T to nullable byte array", src)
}

func (b ByteAOrNull) Value() (driver.Value, error) {
	return b.v, nil
}

func (b *ByteANonNull) Scan(src interface{}) error {
	switch src.(type) {
	case []byte:
		result := bytes.Clone(src.([]byte))
		if len(result) == 0 {
			return fmt.Errorf("cannot convert empty byte array to non-nullable byte array")
		}
		b.v = result
		return nil
	}
	panic(fmt.Errorf("cannot convert %T to non-nullable byte array", src))
}

func (b ByteANonNull) Value() (driver.Value, error) {
	return b.v, nil
}
