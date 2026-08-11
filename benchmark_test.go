package tuesday

import (
	"testing"
	"time"
)

var benchmarkResult string

func BenchmarkStrftime(b *testing.B) {
	value := time.Date(2006, 1, 2, 15, 4, 5, 123456789, time.FixedZone("EST", -5*60*60))
	const format = "%a, %b %d, %Y %H:%M:%S.%6N %:z"
	b.ReportAllocs()
	for b.Loop() {
		var err error
		benchmarkResult, err = Strftime(format, value)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFormatter(b *testing.B) {
	value := time.Date(2006, 1, 2, 15, 4, 5, 123456789, time.FixedZone("EST", -5*60*60))
	formatter, err := Compile("%a, %b %d, %Y %H:%M:%S.%6N %:z")
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		benchmarkResult = formatter.Format(value)
	}
}
