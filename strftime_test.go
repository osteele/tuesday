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

// fills in the gaps from gen.rb
var conversionTests = []struct{ format, expect string }{
	// prefix and suffix
	{"pre%m", "pre01"},
	{"%m.post", "01.post"},
	// empty
	{"", ""},
	// unicode
	{"⌘%m⌘", "⌘01⌘"},

	{"%1N", "1"},
	{"%3N", "123"},
	{"%6N", "123456"},
	{"%9N", "123456789"},
	{"%12N", "123456789000"},

	// width
	{"%1H", "15"},
	{"%2H", "15"},
	{"%3H", "015"},

	// flags, width override zero-padded conversion
	{"%1m", "1"},
	{"%2m", "01"},
	{"%-2m", "1"},
	{"%_2m", " 1"},
	{"%02m", "01"},

	{"%3m", "001"},
	{"%-3m", "1"},
	{"%_3m", "  1"},
	{"%03m", "001"},

	// flags, width override blank-padded conversion
	{"%2e", " 2"},
	{"%-2e", "2"},
	{"%_2e", " 2"},
	{"%02e", "02"},

	{"%3e", "  2"},
	{"%-3e", "2"},
	{"%_3e", "  2"},
	{"%03e", "002"},

	{"%-3H", "15"},
	{"%_3H", " 15"},
	{"%03H", "015"},
	{"%03e", "002"},

	// time zone
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
	{"%Z", "EST"},

	// spot checks
	{"%a, %b %d, %Y", "Mon, Jan 02, 2006"},
	{"%Y/%m/%d", "2006/01/02"},
	{"%Y/%m/%e", "2006/01/ 2"},
	{"%Y/%-m/%-d", "2006/1/2"},
	{"%a, %b %d, %Y %z", "Mon, Jan 02, 2006 -0500"},
	{"%a, %b %d, %Y %Z", "Mon, Jan 02, 2006 EST"},
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
	var hourTests = []struct {
		hour       int
		directives string
	}{
		{0, "%H=00 %k= 0 %I=12 %l=12 %P=am %p=AM"},
		{1, "%H=01 %k= 1 %I=01 %l= 1 %P=am %p=AM"},
		{11, "%H=11 %k=11 %I=11 %l=11 %P=am %p=AM"},
		{12, "%H=12 %k=12 %I=12 %l=12 %P=pm %p=PM"},
		{13, "%H=13 %k=13 %I=01 %l= 1 %P=pm %p=PM"},
		{23, "%H=23 %k=23 %I=11 %l=11 %P=pm %p=PM"},
	}
	for _, test := range hourTests {
		dt := time.Date(2006, 1, 2, test.hour, 4, 5, 0, time.UTC)
		testDirectives(t, test.hour, dt, test.directives)
	}
}

func TestStrftime_dow(t *testing.T) {
	// indexed by 0-based day of month
	var tests = []string{
		"%A=Sunday %a=Sun %u=7 %w=0 %d=01 %e= 1 %j=001 %U=01 %V=52 %W=00",
		"%A=Monday %a=Mon %u=1 %w=1 %d=02 %e= 2 %j=002 %U=01 %V=01 %W=01",
		"%A=Tuesday %a=Tue %u=2 %w=2 %d=03 %e= 3 %j=003 %U=01 %V=01 %W=01",
		"%A=Wednesday %a=Wed %u=3 %w=3 %d=04 %e= 4 %j=004 %U=01 %V=01 %W=01",
		"%A=Thursday %a=Thu %u=4 %w=4 %d=05 %e= 5 %j=005 %U=01 %V=01 %W=01",
		"%A=Friday %a=Fri %u=5 %w=5 %d=06 %e= 6 %j=006 %U=01 %V=01 %W=01",
		"%A=Saturday %a=Sat %u=6 %w=6 %d=07 %e= 7 %j=007 %U=01 %V=01 %W=01",
	}
	for day, tests := range tests {
		dt := time.Date(2006, 1, day+1, 15, 4, 5, 0, time.UTC)
		testDirectives(t, day+1, dt, tests)
	}
}

func TestStrftime_weeks(t *testing.T) {
	var tests = map[int]string{
		2017: "%a=Sun %G=2016 %g=16 %U=01 %V=52 %W=00",
		2007: "%a=Mon %G=2007 %g=07 %U=00 %V=01 %W=01",
		2013: "%a=Tue %G=2013 %g=13 %U=00 %V=01 %W=00",
		2014: "%a=Wed %G=2014 %g=14 %U=00 %V=01 %W=00",
		2015: "%a=Thu %G=2015 %g=15 %U=00 %V=01 %W=00",
		2016: "%a=Fri %G=2015 %g=15 %U=00 %V=53 %W=00",
		2011: "%a=Sat %G=2010 %g=10 %U=00 %V=52 %W=00",
	}
	for year, directives := range tests {
		dt := time.Date(year, 1, 1, 15, 4, 5, 0, time.UTC)
		testDirectives(t, year, dt, directives)
	}
}

func TestStrftime_zones(t *testing.T) {
	tests := []struct{ source, expect string }{
		{"02 Jan 06 15:04 UTC", "%z=+0000 %Z=UTC"},
		{"02 Jan 06 15:04 EST", "%z=-0500 %Z=EST"},
		{"02 Jul 06 15:04 EDT", "%z=-0400 %Z=EDT"},
	}
	for _, test := range tests {
		dt := timeMustParse(time.RFC822, test.source)
		testDirectives(t, test.source, dt, test.expect)
	}

	tests2 := []struct {
		source string
		sec    int // overrides source TZ, if non-zero
		expect string
	}{
		{"02 Jan 06 15:04 +0500", 0, "%z=+0500 %:z=+05:00 %::z=+05:00:00 %:::z=+05"},
		{"02 Jan 06 15:04 +0530", 0, "%z=+0530 %:z=+05:30 %::z=+05:30:00 %:::z=+05:30"},
		{"02 Jan 06 15:04 -0700", 0, "%z=-0700 %:z=-07:00 %::z=-07:00:00 %:::z=-07"},
		{"02 Jan 06 15:04 -0730", 0, "%z=-0730 %:z=-07:30 %::z=-07:30:00 %:::z=-07:30"},
		{"02 Jan 06 15:04 +0000", 45, "%z=+0000 %:z=+00:00 %::z=+00:00:45 %:::z=+00:00:45"},
		{"02 Jan 06 15:04 +0000", 60, "%z=+0001 %:z=+00:01 %::z=+00:01:00 %:::z=+00:01"},
		{"02 Jan 06 15:04 +0000", 105, "%z=+0001 %:z=+00:01 %::z=+00:01:45 %:::z=+00:01:45"},
		{"02 Jan 06 15:04 +0000", -45, "%z=-0000 %:z=-00:00 %::z=-00:00:45 %:::z=-00:00:45"},
		{"02 Jan 06 15:04 +0000", -60, "%z=-0001 %:z=-00:01 %::z=-00:01:00 %:::z=-00:01"},
		{"02 Jan 06 15:04 +0000", -27000, "%z=-0730 %:z=-07:30 %::z=-07:30:00 %:::z=-07:30"},
		{"02 Jan 06 15:04 +0000", -18015, "%z=-0500 %:z=-05:00 %::z=-05:00:15 %:::z=-05:00:15"},
		{"02 Jan 06 15:04 -0730", -27045, "%z=-0730 %:z=-07:30 %::z=-07:30:45 %:::z=-07:30:45"},
	}
	for i, test := range tests2 {
		dt := timeMustParse(time.RFC822Z, test.source)
		if test.sec != 0 {
			loc := time.FixedZone("FTZ", test.sec)
			dt = time.Date(dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second(), dt.Nanosecond(), loc)
		}
		testDirectives(t, i, dt, test.expect)
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

func init() {
	if err := os.Setenv("TZ", "America/New_York"); err != nil {
		log.Fatalf("set timezone %s\n", err)
	}
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

// runs a separate test for each e.g. "%a=Mon" in directives.
func testDirectives(t *testing.T, label interface{}, dt time.Time, directives string) {
	var fieldRE = regexp.MustCompile(`(\S+)=(\s*\S+)`)
	for _, m := range fieldRE.FindAllStringSubmatch(directives, -1) {
		format, expect := m[1], m[2]
		t.Run(fmt.Sprintf("%v.Strftime(%q)", label, format), func(t *testing.T) {
			actual, err := Strftime(format, dt)
			require.NoError(t, err)
			require.Equal(t, expect, actual)
		})
	}
}

func timeMustParse(f, s string) time.Time {
	t, err := time.ParseInLocation(f, s, time.Local)
	if err != nil {
		log.Fatalf("time.ParseInLocation %s\n", err)
	}
	return t
}
