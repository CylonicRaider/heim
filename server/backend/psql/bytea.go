package psql

/* Some time between time immemorial (i.e. ca. 2016) and now, the pq library
 * changed its handling of byte slices to treat nil slices as SQL NULLs rather
 * than empty arrays. This breaks legacy code in this package, which uses NOT
 * NULL columns but actually treats them in a nullable manner with the empty
 * byte array as the null/zero/nil value and assumes that nil byte slices are
 * equivalent to empty ones. The two types introduced in this file allow
 * remediating this mess by clearly distinguishing nullable, non-nullable, and
 * not-audited-yet byte arrays. */

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
		return ByteAOrNull{[]byte{}}
	}
	return ByteAOrNull{data}
}

func EmptyByteAOrNull() ByteAOrNull {
	return ByteAOrNull{[]byte{}}
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
		result := bytes.Clone(src.([]byte))
		if result == nil {
			result = []byte{}
		}
		b.v = result
		return nil
	case nil:
		b.v = []byte{}
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
