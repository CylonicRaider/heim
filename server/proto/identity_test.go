package proto

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNormalizeNick(t *testing.T) {
	pass := func(name string) string {
		name, err := NormalizeNick(name)
		So(err, ShouldBeNil)
		return name
	}

	reject := func(name string) error {
		name, err := NormalizeNick(name)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "invalid nick")
		So(name, ShouldEqual, "")
		return err
	}

	Convey("Spaces are stripped", t, func() {
		So(pass("test"), ShouldEqual, "test")
		So(pass(" test"), ShouldEqual, "test")
		So(pass("\r  test"), ShouldEqual, "test")
		So(pass("test "), ShouldEqual, "test")
		So(pass("test\v  "), ShouldEqual, "test")
		So(pass("  test  "), ShouldEqual, "test")
	})

	Convey("Non-spaces are required", t, func() {
		reject("")
		reject(" ")
		reject(" \t\n\v\f ")
	})

	Convey("Internal spaces are collapsed", t, func() {
		So(pass("test test"), ShouldEqual, "test test")
		So(pass("test \r \n \v test"), ShouldEqual, "test test")
		So(pass(" test\ntest test "), ShouldEqual, "test test test")
	})

	Convey("UTF-8 is handled", t, func() {
		input := `
        ᕦ( ͡° ͜ʖ ͡°)ᕤ    


                   ─=≡Σᕕ( ͡° ͜ʖ ͡°)ᕗ
            ` + "\t\r\n:)"
		expected := "ᕦ( ͡° ͜ʖ ͡°)ᕤ ─=≡Σᕕ( ͡° ͜ʖ ͡°)ᕗ :)"
		So(pass(input), ShouldEqual, expected)
	})

	Convey("Max length is enforced", t, func() {
		name := make([]byte, MaxNickLength)
		for i := 0; i < MaxNickLength; i++ {
			name[i] = 'a'
		}
		name[1] = ' '
		name[4] = ' '
		name[5] = ' '
		expected := "a aa " + string(name[6:])
		So(pass(string(name)), ShouldEqual, expected)
		So(pass(string(name)+"a"), ShouldEqual, expected+"a")
		reject(string(name) + "aa")
	})

	Convey("Bidi isolates and overrides are popped", t, func() {
		testCases := [][]string{
			{"\u202Btest", "\u202Btest\u202C"},
			{"\u202B\u202Bte\u202Cst", "\u202B\u202Bte\u202Cst\u202C"},
			{"\u2067\u202Atest\u2069test", "\u2067\u202Atest\u2069test\u202C"},
			{"\u2067\u202Atest\u202Ctest", "\u2067\u202Atest\u202Ctest\u2069"},
		}

		for _, testCase := range testCases {
			output, err := NormalizeNick(testCase[0])
			So(err, ShouldBeNil)
			So(output, ShouldEqual, testCase[1])
		}
	})

	Convey("Emoji are collapsed", t, func() {
		validEmoji["+1"] = "~plusone"
		name := make([]byte, MaxNickLength+len(":+1:")-1)
		for i := 0; i < MaxNickLength+len(":+1:")-1; i++ {
			name[i] = 'a'
		}
		name[1] = ' '
		name[4] = ' '
		name[5] = ' '
		copy(name[len(name)-len(":+1:"):], ":+1:")
		expected := "a aa " + string(name[6:])
		So(pass(string(name)), ShouldEqual, expected)
	})
}

func TestNickLen(t *testing.T) {
	validEmoji["greenduck"] = "~greenduck"
	validEmoji["apple"] = "1f34e"
	validEmoji["astronaut"] = "1f9d1-200d-1f680"

	Convey("Length of nicks without emoji are correct", t, func() {
		name := ":greenduck"
		So(nickLen(name), ShouldEqual, len(name))
		name = ":not an emoji:"
		So(nickLen(name), ShouldEqual, len(name))
	})

	Convey("Length of nicks with emoji are correct", t, func() {
		name := ":greenduck:greenduck:"
		So(nickLen(name), ShouldEqual, 1+len("greenduck:"))
		name = "foo:greenduck::greenduck:bar"
		So(nickLen(name), ShouldEqual, len("foobar")+2)
	})

	Convey("Length of nicks with multi-codepoint emoji are correct", t, func() {
		name := ":astronaut:"
		So(nickLen(name), ShouldEqual, 3)
	})

	Convey("Testing degenerate case", t, func() {
		name := strings.Repeat(":", 1000000)
		So(nickLen(name), ShouldEqual, len(name))
	})

	Convey("Testing another degenerate case", t, func() {
		name := strings.Repeat(":!", 500000)
		So(nickLen(name), ShouldEqual, len(name))
	})

	Convey("Testing boundary cases", t, func() {
		name := ":greenduck:f"
		So(nickLen(name), ShouldEqual, 2)
	})

	Convey("Testing overlapping shortcodes", t, func() {
		name := ":greenduck:apple:"
		So(nickLen(name), ShouldEqual, 1 + len("apple:"))
	})

	Convey("Testing partially-valid overlapping shortcodes", t, func() {
		name := ":no!such!emoji:greenduck:"
		So(nickLen(name), ShouldEqual, len(":no!such!emoji") + 1)
	})
}

func TestNormalizeEmoji(t *testing.T) {
	validEmoji["greenduck"] = "~greenduck"
	validEmoji["apple"] = "1f34e"
	validEmoji["astronaut"] = "1f9d1-200d-1f680"

	Convey("Plain nicks pass through", t, func() {
		name := "greenduck"
		So(normalizeEmoji(name), ShouldEqual, name)
	})

	Convey("Unicode emoji shortcodes are normalized", t, func() {
		name := "green:apple:duck"
		So(normalizeEmoji(name), ShouldEqual, "green\U0001F34Educk")
	})

	Convey("Multi-codepoint Unicode emoji shortcuts are normalized", t, func() {
		name := ":astronaut:"
		So(normalizeEmoji(name), ShouldEqual, "\U0001F9D1\u200D\U0001F680")
	})

	Convey("Custom emoji are unmodified", t, func() {
		name := "greenduck:greenduck:"
		So(normalizeEmoji(name), ShouldEqual, name)
	})

	Convey("Testing degenerate case", t, func() {
		name := strings.Repeat(":", 1000000)
		So(normalizeEmoji(name), ShouldEqual, name)
	})

	Convey("Testing another degenerate case", t, func() {
		name := strings.Repeat(":!", 500000)
		So(normalizeEmoji(name), ShouldEqual, name)
	})

	Convey("Testing boundary case", t, func() {
		name := ":apple:apple:apple:"
		So(normalizeEmoji(name), ShouldEqual, "\U0001F34Eapple\U0001F34E")
	})

	Convey("Testing overlapping shortcodes", t, func() {
		name := ":greenduck:apple:"
		So(normalizeEmoji(name), ShouldEqual, name)
	})

	Convey("Testing partially-valid overlapping shortcodes", t, func() {
		name := ":no!such!emoji:apple:greenduck:apple:"
		So(normalizeEmoji(name), ShouldEqual, ":no!such!emoji\U0001F34Egreenduck\U0001F34E")
	})
}
