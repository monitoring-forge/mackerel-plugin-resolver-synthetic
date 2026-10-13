package main

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/miekg/dns"
)

func (o *Opt) resolveOnce(ctx context.Context, host string, timeout time.Duration) error {
	address := net.JoinHostPort(host, "53")

	c := &dns.Client{Net: "udp", Timeout: timeout}
	m := new(dns.Msg)
	m.SetQuestion(o.Question, dns.StringToType["A"])
	r, _, err := c.ExchangeContext(ctx, m, address)
	if err != nil {
		return err
	}
	if r.Truncated {
		c = &dns.Client{Net: "tcp", Timeout: timeout}
		r, _, err = c.ExchangeContext(ctx, m, address)
		if err != nil {
			return err
		}
	}
	if r.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("failed to resolve '%s'. rcode:%s",
			o.Question,
			dns.RcodeToString[r.Rcode],
		)
	}
	answer := make([]string, 0)
	for _, a := range r.Answer {
		if aa, ok := a.(*dns.A); ok {
			answer = append(answer, aa.A.String())
		}
	}
	if len(o.Expect) > 0 && !expectAny(o.Expect, answer) {
		return fmt.Errorf("dns answer does not contain '%s' in '%s'",
			o.Expect,
			strings.Join(answer, "\t"))
	}
	if o.ExpectRegexp.IsSet() && !expectMatchAny(o.ExpectRegexp, answer) {
		return fmt.Errorf("dns answer does not match regexp '%s' in '%s'",
			o.ExpectRegexp.Value,
			strings.Join(answer, "\t"))
	}
	return nil
}

func expectAny(haystack string, needles []string) bool {
	if len(needles) == 0 {
		return false
	}
	for _, needle := range needles {
		if strings.Contains(haystack, needle) {
			return true
		}
	}
	return false
}

func expectMatchAny(reg FlagRegexp, needles []string) bool {
	if !reg.IsSet() {
		return false
	}
	for _, needle := range needles {
		if reg.Regexp.MatchString(needle) {
			return true
		}
	}
	return false
}

func (o *Opt) resolvTimeout(k int) time.Duration {
	t := int(o.Timeout.Seconds())
	if k == 1 {
		t = int((t * 2) / len(o.Hosts))
	} else if k > 1 {
		t = int((t*2)/len(o.Hosts)) * 2
	}
	if t < 1 {
		t = 1
	}
	return time.Duration(t) * time.Second
}
