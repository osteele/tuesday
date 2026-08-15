//go:build ignore
// +build ignore

// compare_report.go reads a ruby_behavior_report.json file and compares each
// reported strftime result against tuesday.Strftime. Run it as:
//
//	go run testdata/compare_report.go < ruby_report.json
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/osteele/tuesday"
)

type report struct {
	RubyVersion     string    `json:"ruby_version"`
	RubyDescription string    `json:"ruby_description"`
	GeneratedAt     string    `json:"generated_at"`
	Note            string    `json:"note"`
	Times           []timeRep `json:"times"`
	Datetime        *timeRep  `json:"datetime,omitempty"`
}

type timeRep struct {
	Source            string            `json:"source"`
	ISO8601           string            `json:"iso8601"`
	ZoneName          string            `json:"zone_name"`
	ZoneOffsetSeconds int               `json:"zone_offset_seconds"`
	Formats           map[string]string `json:"formats"`
}

func main() {
	var report report
	dec := json.NewDecoder(os.Stdin)
	if err := dec.Decode(&report); err != nil {
		fmt.Fprintf(os.Stderr, "decode report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("# Comparing tuesday against Ruby %s\n", report.RubyVersion)
	fmt.Printf("# %s\n", report.RubyDescription)

	var mismatches []string
	var skipped int

	for i, tr := range report.Times {
		// Reconstruct the Go time using the offset so that %z/%Z behavior is
		// comparable. Ruby zone names come from the location database, while Go
		// FixedZone names are supplied by the caller, so %Z comparisons are
		// inherently dependent on how the reference time was built.
		loc := time.FixedZone(tr.ZoneName, tr.ZoneOffsetSeconds)
		value, err := time.Parse(time.RFC3339Nano, tr.ISO8601)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse time %d: %v\n", i, err)
			continue
		}
		value = value.In(loc)

		for fmtStr, expected := range tr.Formats {
			actual, err := tuesday.Strftime(fmtStr, value)
			if err != nil {
				mismatches = append(mismatches, fmt.Sprintf("time %d %s: Go error: %v", i, fmtStr, err))
				continue
			}
			if actual != expected {
				// Zone-name output depends on how the reference time was built. When
				// Ruby has no zone name, comparing %Z or %+ composites is not
				// meaningful.
				if tr.ZoneName == "" && (strings.HasSuffix(fmtStr, "Z") || strings.Contains(fmtStr, "+")) {
					skipped++
					continue
				}
				mismatches = append(mismatches, fmt.Sprintf("time %d %%%s: Go=%q Ruby=%q", i, fmtStr[1:], actual, expected))
			}
		}
	}

	if report.Datetime != nil {
		for fmtStr, expected := range report.Datetime.Formats {
			// %Q is only meaningful when interpreted as DateTime milliseconds.
			value, _ := time.Parse(time.RFC3339Nano, report.Datetime.ISO8601)
			actual, err := tuesday.Strftime(fmtStr, value)
			if err != nil {
				mismatches = append(mismatches, fmt.Sprintf("datetime %%%s: Go error: %v", fmtStr[1:], err))
				continue
			}
			if actual != expected {
				mismatches = append(mismatches, fmt.Sprintf("datetime %%%s: Go=%q Ruby=%q", fmtStr[1:], actual, expected))
			}
		}
	}

	fmt.Printf("# Total mismatches: %d (skipped: %d)\n", len(mismatches), skipped)
	for _, m := range mismatches {
		fmt.Println(m)
	}
}
