package snowflake

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMask(t *testing.T) {
	Convey("Verify masks", t, func() {
		tests := map[uint]int64{
			1: 0x1,
			2: 0x3,
			3: 0x7,
			4: 0xf,
			5: 0x1f,
			6: 0x3f,
			7: 0x7f,
			8: 0xff,
		}

		for b, m := range tests {
			So(mask(b), ShouldEqual, m)
		}
	})
}

func TestIdN(t *testing.T) {
	Convey("Verify ids are unique", t, func() {
		generator := New(Options{
			ServerBits:   4,
			SequenceBits: 2,
		})
		n := 2500
		ids := generator.IdN(n)
		So(len(ids), ShouldEqual, n)

		uniques := map[int64]int64{}
		for _, id := range ids {
			uniques[id] = id
		}
		So(len(uniques), ShouldEqual, n)
	})
}

func TestStringN(t *testing.T) {
	Convey("Verify ids are unique", t, func() {
		generator := New(Options{
			ServerBits:   4,
			SequenceBits: 2,
		})

		c := 5
		n := 2500
		uniques := map[string]string{}

		for i := 0; i < c; i++ {
			ids := generator.StringN(n)
			So(len(ids), ShouldEqual, n)

			for _, id := range ids {
				uniques[id] = id
			}
		}
		So(len(uniques), ShouldEqual, n * c)
	})
}
