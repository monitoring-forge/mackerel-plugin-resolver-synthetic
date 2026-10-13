package main

import (
	"regexp"
	"testing"
)

func TestFlagRegexp(t *testing.T) {
	testCases := []struct {
		input    string
		err      bool
		expected bool
	}{
		{"", false, false},
		{"^example", false, true},
		{"invalid[", true, false},
	}
	for _, tc := range testCases {
		fr := &FlagRegexp{}
		err := fr.UnmarshalFlag(tc.input)
		if !tc.err && err != nil {
			t.Errorf("unexpected error for input %q: %v", tc.input, err)
		}
		if tc.err && err == nil {
			t.Errorf("expected error for input %q, but got none", tc.input)
		}
		if fr.IsSet() != tc.expected {
			t.Errorf("input: %q, expected IsSet: %v, got: %v", tc.input, tc.expected, fr.IsSet())
		}
	}
}

func TestOptValidator(t *testing.T) {
	testFlagRegexp := FlagRegexp{
		Value:  "^example",
		Regexp: regexp.MustCompile("^example"),
	}
	testCases := []struct {
		expect       string
		expectRegexp FlagRegexp
		err          bool
	}{
		{"", FlagRegexp{}, false},
		{"value", FlagRegexp{}, false},
		{"", testFlagRegexp, false},
		{"value", testFlagRegexp, true},
	}
	for _, tc := range testCases {
		opt := &Opt{
			Expect:       tc.expect,
			ExpectRegexp: tc.expectRegexp,
		}
		err := opt.Validator(nil)
		if !tc.err && err != nil {
			t.Errorf("unexpected error for expect %q and expect-regexp %q: %v", tc.expect, tc.expectRegexp.Value, err)
		}
		if tc.err && err == nil {
			t.Errorf("expected error for expect %q and expect-regexp %q, but got none", tc.expect, tc.expectRegexp.Value)
		}
	}
}

func TestEmptyOptValiddator(t *testing.T) {
	opt := &Opt{}
	err := opt.Validator(nil)
	if err != nil {
		t.Errorf("unexpected error for empty Opt: %v", err)
	}
}
