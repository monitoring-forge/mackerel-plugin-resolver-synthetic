package main

import (
	"context"
	"net"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func duration(s string) time.Duration {
	t, _ := time.ParseDuration(s)
	return t
}

func TestResolvTimeout(t *testing.T) {
	testCases := []struct {
		timeout   string
		hosts     []string
		durations []string
	}{
		{"5s", []string{"a", "b"}, []string{"5s", "5s"}},
		{"1s", []string{"a", "b"}, []string{"1s", "1s"}},
		{"2s", []string{"a", "b", "c"}, []string{"2s", "1s", "2s"}},
		{"5s", []string{"a", "b", "c"}, []string{"5s", "3s", "6s"}},
	}
	for _, tc := range testCases {
		assert.Len(t, tc.hosts, len(tc.durations))
		opt := &Opt{
			Timeout: duration(tc.timeout),
			Hosts:   tc.hosts,
		}
		for i, expected := range tc.durations {
			assert.Equal(t, duration(expected), opt.resolvTimeout(i))
		}
	}
}

func TestResolveOnce(t *testing.T) {
	ctx := context.Background()
	opt := &Opt{
		Question: "example.com.",
	}
	err := opt.resolveOnce(ctx, "8.8.8.8", duration("5s"))
	assert.NoError(t, err)

	opt = &Opt{
		Question: "example-hoge-fuga-mitsukaranai.com.",
	}
	err = opt.resolveOnce(ctx, "8.8.8.8", duration("5s"))
	assert.Error(t, err)

	opt = &Opt{
		Question: "example.com.",
	}
	err = opt.resolveOnce(ctx, "8.8.8.1", duration("1s"))
	assert.Error(t, err)

	// test if can connect to 8.8.8.8 with tcp
	if conn, err := net.DialTimeout("tcp", "8.8.8.8:53", duration("5s")); err == nil {
		_ = conn.Close()
		opt = &Opt{
			Question: "tcp-fallback.kazeburo.work.",
			Expect:   "192.168.77.1",
		}
		err = opt.resolveOnce(ctx, "8.8.8.8", duration("10s"))
		assert.NoError(t, err)
	}

	ctx, cancel := context.WithTimeout(ctx, duration("3s"))
	defer cancel()
	opt = &Opt{
		Question: "example.com.",
	}
	err = opt.resolveOnce(ctx, "192.0.2.42", duration("10s"))
	assert.Error(t, err)
	assert.NotNil(t, ctx.Err())

}

func TestExpectAny(t *testing.T) {
	assert.True(t, expectAny("hello world", []string{"hello world"}))
	assert.True(t, expectAny("hello", []string{"hello world"}))
	assert.True(t, expectAny("hello world", []string{"world", "hello world"}))
	assert.True(t, expectAny("world", []string{"world", "hello world"}))
	assert.False(t, expectAny("hello world", []string{"foo"}))
	assert.False(t, expectAny("hello world", []string{}))
	assert.False(t, expectAny("", []string{"hello"}))
}

func TestExpectMatchAny(t *testing.T) {
	reg := FlagRegexp{
		Value:  "hello",
		Regexp: regexp.MustCompile("hello"),
	}
	assert.True(t, expectMatchAny(reg, []string{"hello world"}))
	assert.True(t, expectMatchAny(reg, []string{"world", "hello world"}))
	assert.False(t, expectMatchAny(reg, []string{"world"}))
	assert.False(t, expectMatchAny(reg, []string{}))
	assert.False(t, expectMatchAny(FlagRegexp{}, []string{"hello"}))
}

func TestResolveOnceExpect(t *testing.T) {
	ctx := context.Background()
	opt := &Opt{
		Question: "example.com.",
		Expect:   "127.0.0.1",
	}
	err := opt.resolveOnce(ctx, "8.8.8.8", duration("5s"))
	assert.Error(t, err)

	opt = &Opt{
		Question: "example.com.",
		ExpectRegexp: FlagRegexp{
			Value:  `\d+\.\d+\.\d+\.\d+`,
			Regexp: regexp.MustCompile(`\d+\.\d+\.\d+\.\d+`),
		},
	}
	err = opt.resolveOnce(ctx, "8.8.8.8", duration("5s"))
	assert.NoError(t, err)

	opt = &Opt{
		Question: "example.com.",
		ExpectRegexp: FlagRegexp{
			Value:  `\d+\.\d+\.\d+\.A`,
			Regexp: regexp.MustCompile(`\d+\.\d+\.\d+\.A`),
		},
	}
	err = opt.resolveOnce(ctx, "8.8.8.8", duration("5s"))
	assert.Error(t, err)
}
