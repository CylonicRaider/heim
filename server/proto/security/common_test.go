package security

import "fmt"

func ShouldBeDistinctByteSlice(a interface{}, b ...interface{}) string {
	fail := func(context string, expected, got interface{}) string {
		return fmt.Sprintf("ShouldBeDistinctByteSlice: %s: expected %s, got %s", context, expected, got)
	}

	if len(b) != 1 {
		return fail("extra arguments", 1, len(b))
	}
	sa, ok := a.([]byte)
	if !ok {
		return fail("first argument", "[]byte", fmt.Sprintf("%T", a))
	}
	sb, ok := b[0].([]byte)
	if !ok {
		return fail("second argument", "[]byte", fmt.Sprintf("%T", b[0]))
	}
	if &sa[0] == &sb[0] {
		return "ShouldBeDistinctByteSlice: slices point to same data"
	}
	return ""
}
