package adunitconfig

import (
	"regexp"
	"testing"
)

func TestContainer(t *testing.T) {
	cases := []struct {
		name         string
		exp          string
		str          string
		expect_err   bool
		expect_match bool
	}{
		{"test1", "^[hc]at", "cat", false, true},
		{"test2", "^[hc]at", "hat", false, true},
		{"test3", "^[hc]at", "hot", false, false},
		{"test4", `^^^[ddd!!\1\1\1\1`, "hot", true, true},
	}

	cont := newContainer()
	for _, c := range cases {
		re, err := cont.Get(c.exp)
		if (err != nil) != c.expect_err {
			t.Error("expect error, but got", err.Error())
		}
		if c.expect_err {
			continue
		}
		match := re.MatchString(c.str)
		if match != c.expect_match {
			t.Error("expect ", c.expect_match, ", but got ", match, "for test ", c.name)
		}
	}

}

func BenchmarkRegexpPackageCompile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		re, _ := regexp.Compile(`^[hc]at`)
		re.MatchString("cat")
	}
}

func BenchmarkRegexpCachePackageCompile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		re, _ := Compile(`^[hc]at`)
		re.MatchString("cat")
	}
}

func TestMatchString(t *testing.T) {
	testCases := []struct {
		pattern  string
		input    string
		expected bool
	}{
		{`^[a-zA-Z]+$`, "abcdef", true},
		{`^[0-9]+$`, "12345", true},
		{`^[a-zA-Z]+$`, "12345", false},
		{`^[0-9]+$`, "abcdef", false},
	}

	for _, tc := range testCases {
		t.Run(tc.pattern, func(t *testing.T) {
			matched, err := MatchString(tc.pattern, tc.input)

			if err != nil {
				t.Fatalf("Error while matching: %v", err)
			}

			if matched != tc.expected {
				t.Errorf("Expected match %v for input '%s' with pattern '%s', but got %v",
					tc.expected, tc.input, tc.pattern, matched)
			}
		})
	}
}
