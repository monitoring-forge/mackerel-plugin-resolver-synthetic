package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	mp "github.com/mackerelio/go-mackerel-plugin"
	"github.com/miekg/dns"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	StatusCodeOK      = 0
	StatusCodeWARNING = 1
)

var version string
var commit string

type Opt struct {
	Version bool   `short:"v" long:"version" description:"Show version"`
	Prefix  string `long:"prefix" default:"resolver" description:"Metric key prefix"`

	Hosts    []string      `short:"H" long:"hostname" default:"127.0.0.1" description:"DNS server hostnames"`
	Question string        `short:"Q" long:"question" default:"example.com." description:"Question hostname"`
	Expect   string        `short:"E" long:"expect" default:"" description:"Expect string in result"`
	Timeout  time.Duration `long:"timeout" default:"5s" description:"Timeout"`
	Attempts int           `long:"attempts" default:"2" description:"Number of resoluitions"`
	Deadline time.Duration `long:"deadline" default:"20s" description:"Deadline timeout"`
}

func (o *Opt) MetricKeyPrefix() string {
	if o.Prefix == "" {
		return "resolver-synthetic"
	}
	return o.Prefix + "-synthetic"
}

func (o *Opt) GraphDefinition() map[string]mp.Graphs {
	labelPrefix := cases.Title(language.Und, cases.NoLower).String(o.MetricKeyPrefix())
	return map[string]mp.Graphs{
		"service": {
			Label: labelPrefix + ": Available",
			Unit:  mp.UnitPercentage,
			Metrics: []mp.Metrics{
				{Name: "available", Label: "available", Diff: false, Stacked: true},
			},
		},
		"rtt": {
			Label: labelPrefix + ": RTT",
			Unit:  mp.UnitInteger,
			Metrics: []mp.Metrics{
				{Name: "milliseconds", Label: "milliseconds", Diff: false, Stacked: false},
			},
		},
	}
}

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
	if len(o.Expect) > 0 && !strings.Contains(strings.Join(answer, "|"), o.Expect) {
		return fmt.Errorf("dns answer does not contain '%s' in '%s'",
			o.Expect,
			strings.Join(answer, "\t"))
	}
	return nil
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

func (o *Opt) FetchMetrics() (map[string]float64, error) {
	result := map[string]float64{}

	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(o.Deadline),
	)
	defer cancel()

	if !strings.HasSuffix(o.Question, ".") {
		o.Question = o.Question + "."
	}
	rtt := time.Duration(0)
	var err error
OUTLOOP:
	for i := 0; i < o.Attempts; i++ {
		for k, h := range o.Hosts {
			timeout := o.resolvTimeout(k)
			n := time.Now()
			err = o.resolveOnce(ctx, h, timeout)
			rtt += time.Since(n)
			if err == nil {
				break OUTLOOP
			}
			log.Printf("failed to resolv on %s with timeout %fs: %v", h, timeout.Seconds(), err)
			if ctx.Err() != nil {
				break OUTLOOP
			}
		}
	}

	if err != nil {
		result["available"] = 0
	} else {
		result["available"] = 100
	}
	result["milliseconds"] = float64(rtt.Milliseconds())

	return result, nil
}

func (o *Opt) Run() {
	plugin := mp.NewMackerelPlugin(o)
	plugin.Run()
}

func main() {
	opt := &Opt{}
	psr := flags.NewParser(opt, flags.HelpFlag|flags.PassDoubleDash)
	_, err := psr.Parse()
	if opt.Version {
		if commit == "" {
			commit = "dev"
		}
		fmt.Printf(
			"%s-%s\n%s/%s, %s, %s\n",
			filepath.Base(os.Args[0]),
			version,
			runtime.GOOS,
			runtime.GOARCH,
			runtime.Version(),
			commit)
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(StatusCodeWARNING)
	}

	opt.Run()
}
