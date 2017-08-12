package tuesday

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func init() {
	if err := os.Setenv("TZ", "America/New_York"); err != nil {
		log.Fatal(err)
	}
}

func timeMustParse(f, s string) time.Time {
	t, err := time.ParseInLocation(f, s, time.Local)
	if err != nil {
		panic(err)
	}
	return t
}

// fills in the gaps from gen.rb
var conversionTests = []struct{ format, expect string }{
	// prefix and suffix
	{"pre%m", "pre01"},
	{"%mpost", "01post"},
	{"⌘%m⌘", "⌘01⌘"},
	{"", ""},

	{"%1N", "1"},
	{"%3N", "123"},
	{"%6N", "123456"},
	{"%9N", "123456789"},
	{"%12N", "123456789000"},

	// flags and width override zero-padded conversion
	{"%1m", "1"},
	{"%2m", "01"},
	{"%3m", "001"},
	{"%-2m", "1"},
	{"%_2m", " 1"},
	{"%02m", "01"},

	// flags and width override blank-padded conversion
	{"%2e", " 2"},
	{"%-2e", "2"},
	{"%_2e", " 2"},
	{"%02e", "02"},

	// width
	{"%1H", "15"},
	{"%2H", "15"},
	{"%3H", "015"},

	{"%:z", "-05:00"},
	{"%::z", "-05:00:00"},

	{"%%", "%"},
	{"%t", "\t"},
	{"%n", "\n"},

	// other runes are passed through
	{"%&", "%&"},
	{"%J", "%J"},
	{"%⌘", "%⌘"},

	// Date.strftime uses these, but the test table is generated from Time
	{"%Q", "1136232245123456"},
	{"%_Q", "1136232245123456"},
	{"%+", "Mon Jan  2 15:04:05 EST 2006"},

	// do what Ruby says it does, rather than what it does
	{"%v", " 2-Jan-2006"},
	{"%Z", "EST"},

	// spot checks
	{"%a, %b %d, %Y", "Mon, Jan 02, 2006"},
	{"%Y/%m/%d", "2006/01/02"},
	{"%Y/%m/%e", "2006/01/ 2"},
	{"%Y/%-m/%-d", "2006/1/2"},
	{"%a, %b %d, %Y %z", "Mon, Jan 02, 2006 -0500"},
	{"%a, %b %d, %Y %Z", "Mon, Jan 02, 2006 EST"},
}

var fieldRE = regexp.MustCompile(`(\S+)=(\s*\S+)`)

var hourTests = []struct {
	hour  int
	tests string
}{
	{0, "%H=00 %k= 0 %I=12 %l=12 %P=am %p=AM"},
	{1, "%H=01 %k= 1 %I=01 %l= 1 %P=am %p=AM"},
	{12, "%H=12 %k=12 %I=12 %l=12 %P=pm %p=PM"},
	{13, "%H=13 %k=13 %I=01 %l= 1 %P=pm %p=PM"},
	{23, "%H=23 %k=23 %I=11 %l=11 %P=pm %p=PM"},
}

var dayOfWeekTests = []string{
	"%A=Sunday %a=Sun %u=7 %w=0 %d=01 %e= 1 %j=001 %U=01 %V=52 %W=00",
	"%A=Monday %a=Mon %u=1 %w=1 %d=02 %e= 2 %j=002 %U=01 %V=01 %W=01",
	"%A=Tuesday %a=Tue %u=2 %w=2 %d=03 %e= 3 %j=003 %U=01 %V=01 %W=01",
	"%A=Wednesday %a=Wed %u=3 %w=3 %d=04 %e= 4 %j=004 %U=01 %V=01 %W=01",
	"%A=Thursday %a=Thu %u=4 %w=4 %d=05 %e= 5 %j=005 %U=01 %V=01 %W=01",
	"%A=Friday %a=Fri %u=5 %w=5 %d=06 %e= 6 %j=006 %U=01 %V=01 %W=01",
	"%A=Saturday %a=Sat %u=6 %w=6 %d=07 %e= 7 %j=007 %U=01 %V=01 %W=01",
}

func readConversionTests() ([][]string, error) {
	skip := map[string]bool{"%_z": true}
	for _, test := range conversionTests {
		skip[test.format] = true
	}

	f, err := os.Open("testdata/tests.csv")
	if err != nil {
		return nil, err
	}
	defer f.Close() // nolint: errcheck

	r := csv.NewReader(f)
	recs, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	tests := make([][]string, 0, len(recs))
	for _, row := range recs {
		if !skip[row[0]] {
			tests = append(tests, row)
		}
	}
	return tests, nil
}

func TestStrftime(t *testing.T) {
	dt := timeMustParse(time.RFC3339Nano, "2006-01-02T15:04:05.123456789-05:00")
	for _, test := range conversionTests {
		name := fmt.Sprintf("Strftime %q", test.format)
		actual, err := Strftime(test.format, dt)
		require.NoErrorf(t, err, name)
		require.Equalf(t, test.expect, actual, name)
	}

	tests, err := readConversionTests()
	if err != nil {
		log.Fatal(err)
	}
	for _, row := range tests {
		format, expect := row[0], row[1]
		name := fmt.Sprintf("Strftime %q", format)
		actual, err := Strftime(format, dt)
		require.NoErrorf(t, err, name)
		require.Equalf(t, expect, actual, name)
	}
}

func TestStrftime_hours(t *testing.T) {
	for _, hour := range hourTests {
		dt := time.Date(2006, 01, 2, hour.hour, 4, 5, 0, time.UTC)
		for _, m := range fieldRE.FindAllStringSubmatch(hour.tests, -1) {
			format, expect := m[1], m[2]
			t.Run(fmt.Sprintf("hour(%v).Strftime(%q)", hour.hour, format), func(t *testing.T) {
				actual, err := Strftime(format, dt)
				require.NoError(t, err)
				require.Equal(t, expect, actual)
			})
		}
	}
}

func TestStrftime_dow(t *testing.T) {
	for day, tests := range dayOfWeekTests {
		dt := time.Date(2006, 01, day+1, 15, 4, 5, 0, time.UTC)
		for _, m := range fieldRE.FindAllStringSubmatch(tests, -1) {
			format, expect := m[1], m[2]
			t.Run(fmt.Sprintf("day(%v).Strftime(%q)", day, format), func(t *testing.T) {
				actual, err := Strftime(format, dt)
				require.NoError(t, err)
				require.Equal(t, expect, actual)
			})
		}
	}
}

func TestStrftime_zones(t *testing.T) {
	tests := []struct{ source, expect string }{
		{"02 Jan 06 15:04 UTC", "%z=+0000 %Z=UTC"},
		{"02 Jan 06 15:04 EST", "%z=-0500 %Z=EST"},
		{"02 Jul 06 15:04 EDT", "%z=-0400 %Z=EDT"},
	}
	for _, test := range tests {
		rt := timeMustParse(time.RFC822, test.source)
		actual, err := Strftime("%%z=%z %%Z=%Z", rt)
		require.NoErrorf(t, err, test.source)
		require.Equalf(t, test.expect, actual, test.source)
	}
}

func ExampleStrftime_flags() {
	t, _ := time.Parse(time.RFC822, "10 Jul 17 18:45 EDT")
	s, _ := Strftime("%B %^B %m %_m %-m %6Y", t)
	fmt.Println(s)
	// Output: July JULY 07  7 7 002017
}

func ExampleStrftime_timezone() {
	t, _ := time.Parse(time.RFC822, "10 Jul 17 18:45 EDT")
	s, _ := Strftime("%Z %z %:z %::z", t)
	fmt.Println(s)
	// Output: EDT -0400 -04:00 -04:00:00
}
