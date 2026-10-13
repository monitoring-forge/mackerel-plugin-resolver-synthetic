package main

import (
	"context"
	"log"
	"strings"
	"time"

	mp "github.com/mackerelio/go-mackerel-plugin"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

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
