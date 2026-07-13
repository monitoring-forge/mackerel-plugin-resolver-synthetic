package main

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func duration(s string) time.Duration {
	t, _ := time.ParseDuration(s)
	return t
}
func TestResolvTimeout(t *testing.T) {
	opt := &Opt{
		Timeout: duration("5s"),
		Hosts:   []string{"a", "b"},
	}
	assert.Equal(t, duration("5s"), opt.resolvTimeout(0))
	assert.Equal(t, duration("5s"), opt.resolvTimeout(1))

	opt = &Opt{
		Timeout: duration("1s"),
		Hosts:   []string{"a", "b"},
	}
	assert.Equal(t, duration("1s"), opt.resolvTimeout(0))
	assert.Equal(t, duration("1s"), opt.resolvTimeout(1))

	opt = &Opt{
		Timeout: duration("1s"),
		Hosts:   []string{"a", "b", "c"},
	}
	assert.Equal(t, duration("1s"), opt.resolvTimeout(0))
	assert.Equal(t, duration("1s"), opt.resolvTimeout(1))
	assert.Equal(t, duration("1s"), opt.resolvTimeout(2))

	opt = &Opt{
		Timeout: duration("2s"),
		Hosts:   []string{"a", "b", "c"},
	}
	assert.Equal(t, duration("2s"), opt.resolvTimeout(0))
	assert.Equal(t, duration("1s"), opt.resolvTimeout(1))
	assert.Equal(t, duration("2s"), opt.resolvTimeout(2))

	opt = &Opt{
		Timeout: duration("5s"),
		Hosts:   []string{"a", "b", "c"},
	}
	assert.Equal(t, duration("5s"), opt.resolvTimeout(0))
	assert.Equal(t, duration("3s"), opt.resolvTimeout(1))
	assert.Equal(t, duration("6s"), opt.resolvTimeout(2))
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
		conn.Close()
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
	err = opt.resolveOnce(ctx, "8.8.8.1", duration("10s"))
	assert.Error(t, err)
	assert.NotNil(t, ctx.Err())

}
