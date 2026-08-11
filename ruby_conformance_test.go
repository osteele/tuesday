package tuesday

import (
	"bytes"
	"encoding/csv"
	"os/exec"
	"testing"
	"time"
)

func TestRubyConformance(t *testing.T) {
	ruby, err := exec.LookPath("ruby")
	if err != nil {
		t.Skip("Ruby is not installed")
	}

	output, err := exec.Command(ruby, "testdata/conformance.rb").Output()
	if err != nil {
		t.Fatalf("generate Ruby conformance matrix: %v", err)
	}
	records, err := csv.NewReader(bytes.NewReader(output)).ReadAll()
	if err != nil {
		t.Fatalf("read Ruby conformance matrix: %v", err)
	}

	for i, record := range records {
		if len(record) != 3 {
			t.Fatalf("row %d has %d fields", i, len(record))
		}
		source, format, expect := record[0], record[1], record[2]
		value, err := time.Parse(time.RFC3339Nano, source)
		if err != nil {
			t.Fatalf("parse row %d time: %v", i, err)
		}
		actual, err := Strftime(format, value)
		if err != nil {
			t.Fatalf("row %d Strftime(%q): %v", i, format, err)
		}
		if actual != expect {
			t.Errorf("row %d Strftime(%q, %s) = %q, want %q", i, format, source, actual, expect)
		}
	}

	if len(records) < 500 {
		t.Fatalf("conformance matrix has only %d rows", len(records))
	}
	t.Logf("compared tuesday.Strftime against Ruby across %d cases", len(records))
}
