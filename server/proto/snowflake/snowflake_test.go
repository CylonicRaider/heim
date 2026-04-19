package snowflake

import (
	"time"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const MANY = 1000000

func TestSnowflakes(t *testing.T) {
	Convey("Snowflakes increase", t, func() {
		sf1, err := New()
		So(err, ShouldBeNil)

		sf2, err := New()
		So(err, ShouldBeNil)

		So(sf1, ShouldBeLessThan, sf2)
	})

	// GoConvey's DSL is nice, but the raw conditionals allow me to actually exercise the
	// wait-for-the-next-millisecond case. Also, this is ~25 times faster.
	Convey("Snowflakes are unique", t, func() {
		seen := map[Snowflake]bool{}
		var prev Snowflake

		for i := 0; i < MANY; i++ {
			sf, err := New()
			if err != nil {
				t.Errorf("snowflake generation failed: %s", err)
			}

			if sf <= prev {
				t.Errorf("snowflake did not increase: %s <= %s", sf, prev)
			}
			prev = sf

			if seen[sf] {
				t.Errorf("snowflake is not unique: %s", sf)
			}
			seen[sf] = true
		}

		if len(seen) != MANY {
			t.Errorf("sanity check failed: expected %d snowflakes, got %d", MANY, len(seen))
		}
	})

	Convey("Snowflake times are correct", t, func() {
		now := time.Now()
		sf, err := New()
		So(err, ShouldBeNil)

		// One millisecond ought to be enough to execute the above.
		So(sf.Time().UnixNano(), ShouldAlmostEqual, now.UnixNano(), 1000000)
	})

	Convey("Snowflake string conversion works", t, func() {
		sf, err := NewFromString("094d7n6z80f0g")
		So(err, ShouldBeNil)

		So(sf.Time(), ShouldEqual, time.Date(2023, 12, 25, 22, 4, 1, 605000000, time.UTC))

		So(sf.String(), ShouldEqual, "094d7n6z80f0g")
	})
}
