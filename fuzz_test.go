package tuesday

import (
	"testing"
	"time"
)

func FuzzStrftime(f *testing.F) {
	for _, format := range []string{
		"",
		"%Y-%m-%dT%H:%M:%S.%N%:z",
		"%-^A %_10z %12L",
		"%_J %EJ %:J",
		"⌘%m⌘",
		"%999999999999999999Y",
	} {
		f.Add(format, int64(0), int32(0))
	}

	f.Fuzz(func(t *testing.T, format string, unixSeconds int64, nanoseconds int32) {
		if len(format) > 1024 {
			t.Skip()
		}
		value := time.Unix(unixSeconds, int64(nanoseconds)%int64(time.Second)).UTC()
		formatter, err := Compile(format)
		if err != nil {
			return
		}
		compiled := formatter.Format(value)
		direct, err := Strftime(format, value)
		if err != nil {
			t.Fatalf("Strftime after successful Compile: %v", err)
		}
		if direct != compiled {
			t.Fatalf("Strftime result %q differs from compiled result %q", direct, compiled)
		}
	})
}
