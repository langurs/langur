// langur/str/str_test.go

package str

import (
	"runtime"
	"testing"
)

func TestSystemSpecificLineReturn(t *testing.T) {
	var expected string

	// https://golang.org/doc/install/source#environment
	if runtime.GOOS == "windows" {
		expected = "\x0D\x0A"
	} else {
		expected = "\x0A"
	}

	if SysNewLine != expected {
		t.Fatalf(`System-specific newline (on %q) expected=%q, received=%q
It may be that this test needs to be updated, or that the method for setting the system-specific newline needs to be updated (or both).`,
			runtime.GOOS, expected, SysNewLine)
	}
}

func TestCpSubString(t *testing.T) {
	tests := []struct {
		input    string
		start    int
		limit    int
		expected string
	}{
		{"abcd", 2, 2, "bc"},
		{"abשלםcd", 2, 2, "bש"},
		{"abשלםcd", 2, -1, "bשלםcd"},
		{"abשלםcd", 1, -1, "abשלםcd"},
		{"abשלםcd", 3, 3, "שלם"},
	}

	for _, tt := range tests {
		result, err := SubStr(tt.input, tt.start, tt.limit)
		if err != nil {
			t.Errorf(err.Error())
		}

		if result != tt.expected {
			t.Errorf("expected=%q, received=%q", tt.expected, result)
		}
	}
}

func TestCpLimit(t *testing.T) {
	tests := []struct {
		input    string
		limit    int
		expected string
	}{
		{"abcdefg", 8, "abcdefg"},
		{"abcdefg", 7, "abcdefg"},
		{"abcdefg", 6, "abc..."},
		{"abcdefg", 5, "ab..."},
		{"abcdefg", 4, "a..."},
		{"abcdefg", 3, "..."},
		{"abcdefg", 2, "..."},
		{"abcdefg", 1, "..."},
		{"abcdefg", 0, "..."},

		{"abcdefg", -8, "abcdefg"},
		{"abcdefg", -7, "abcdefg"},
		{"abcdefg", -6, "...efg"},
		{"abcdefg", -5, "...fg"},
		{"abcdefg", -4, "...g"},
		{"abcdefg", -3, "..."},
		{"abcdefg", -2, "..."},
		{"abcdefg", -1, "..."},
		{"abcdefg", -0, "..."},
		{"bשלםcdefg", 7, "bשלם..."},
		{"bשלםcdefg", -7, "...defg"},
	}

	for _, tt := range tests {
		result := Limit(tt.input, tt.limit, "...")

		if result != tt.expected {
			t.Errorf("(%q, %d) expected=%q, received=%q", tt.input, tt.limit, tt.expected, result)
		}
	}
}

func TestBalanced(t *testing.T) {
	tests := []struct {
		s      string
		result bool
	}{
		// LTR/RTL embedding/isolate marks balanced?
		{"", true},
		{"abcd", true},
		{"\u2066abcd\u2069", true},
		{"\u2066abcd", false},
		{"abcd\u2069", false},

		{"\u2067abcd\u2069", true},
		{"\u2067abcd", false},
		{"abcd\u2069", false},

		{"\u2068abcd\u2069", true},
		{"\u2068abcd", false},
		{"abcd\u2069", false},
		{"abcd\u202C", false},

		{"\u2066a\u202Abc\u202Cd\u2069", true},
		{"\u2066a\u202Abc\u2069d\u202C", false},

		{"\u202Aa\u2066bc\u2069d\u202C", true},
		{"\u202Aa\u2066bc\u202Cd\u2069", false},
	}

	for _, tt := range tests {
		result := Balanced(tt.s)
		if result != tt.result {
			t.Errorf("String balanced expected to be %t", tt.result)
		}
	}
}

func TestRemoveTrailing(t *testing.T) {
	tests := []struct {
		s       string
		trailer byte
		result  string
	}{
		{"1.9", '0', "1.9"},
		{"1.90", '0', "1.9"},
		{"1.90000000000", '0', "1.9"},
	}

	for _, tt := range tests {
		result := RemoveTrailing(tt.s, tt.trailer)
		if result != tt.result {
			t.Errorf("remove trailing %b from %q expected %q, received %q",
				tt.trailer, tt.s, tt.result, result)
		}
	}
}

func TestStrToInt64(t *testing.T) {
	tests := []struct {
		s      string
		result int64
		ok     bool
	}{
		{"789456123", 789456123, true},
		{"255", 255, true},
		{"7", 7, true},
		{"1", 1, true},
		{"0", 0, true},

		{"-789456123", -789456123, true},
		{"-255", -255, true},
		{"-7", -7, true},
		{"-1", -1, true},
	}

	for _, tt := range tests {
		result, ok := StrToInt64(tt.s, 10)
		if result != tt.result {
			t.Errorf("String to int64 expected %d, received %d", tt.result, result)
		}
		if (ok == nil) != tt.ok {
			t.Errorf("String to int64 ok status expected %t, received %t", tt.ok, (ok == nil))
		}
	}
}

// func TestChanges(t *testing.T) {
// 	tests := []struct {
// 		s      string
// 		result bool
// 	}{}

// 	for _, tt := range tests {
// 		result := Balanced(tt.s)
// 		if result != tt.result {
// 			t.Errorf("String balanced expected to be %t", tt.result)
// 		}
// 	}
// }
