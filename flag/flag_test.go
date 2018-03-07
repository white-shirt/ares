package flag_test

import (
	"testing"

	. "github.com/sevenNt/ares/flag"
)

func TestParse(t *testing.T) {
	Convey("解析Flag", t, func() {
		fs := []Flag{
			&StringFlag{
				Name:    "s,str",
				Default: "nostring",
			},
			&BoolFlag{
				Name:    "b,bool",
				Default: true,
			},
		}

		With(fs...)
		Parse()

		So(Bool("b"), ShouldEqual, true)
		So(Bool("bool", ShouldEqual, true))

		So(String("s"), ShouldEqual, "nostring")
		So(String("str"), ShouldEqual, "nostring")
	})
}
