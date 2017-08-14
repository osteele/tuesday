// Package tuesday implements a Strftime function that is compatible with Ruby's Time.strftime.
package tuesday

//go:generate ruby testdata/gen.rb

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Strftime is compatible with Ruby's Time.strftime.
//
// See https://ruby-doc.org/core-2.4.1/Time.html#method-i-strftime
//
// Strftime returns an error for compatibility with other formatting functions and for future compatibility,
// but in the current implementation this error is always nil.
func Strftime(format string, t time.Time) (string, error) {
	return re.ReplaceAllStringFunc(format, func(directive string) string {
		var (
			pad, w        = '0', 2
			m             = re.FindAllStringSubmatch(directive, 1)[0]
			flags, width  = m[1], m[2]
			conversion, _ = utf8.DecodeRuneInString(m[3])
			c             = convert(t, conversion, flags, width)
		)
		if s, ok := c.(string); ok {
			return applyFlags(flags, s)
		}
		if f, ok := defaultPadding[conversion]; ok {
			pad, w = f.c, f.w
		}
		if len(width) > 0 {
			w, _ = strconv.Atoi(width) // nolint: gas
		}
		switch flags {
		case "-":
			w = 0
		case "_":
			pad = '-'
		case "0":
			pad = '0'
		}
		var fm string
		switch {
		// Hardcode the defaults:
		case pad == '-' && w == 2:
			fm = "%2d"
		case pad == '0' && w == 2:
			fm = "%02d"
		case pad == '0' && w == 3:
			fm = "%03d"
		case pad == '-':
			fm = fmt.Sprintf("%%%dd", w)
		default:
			fm = fmt.Sprintf("%%%c%dd", pad, w)
		}
		s := fmt.Sprintf(fm, c)
		return applyFlags(flags, s)
	}), nil
}

var re = regexp.MustCompile(`%([-_^#0]|:{1,3})?(\d+)?[EO]?([a-zA-Z\+nt%])`)

// Test whether a string is uppercase, for purpose of applying the # case reversal flag.
// This is ASCII-only and is foiled by spaces and punctuation, but is sufficient for this context.
var isUpperRE = regexp.MustCompile(`^[[:upper:]]+$`).MatchString

var amPmTable = map[bool]string{true: "AM", false: "PM"}
var amPmLowerTable = map[bool]string{true: "am", false: "pm"}

// Default padding character and width, by conversion rune.
// The default default is pad='0', width=2
var defaultPadding = map[rune]struct {
	c rune
	w int
}{
	'e': {'-', 2},
	'f': {'0', 6},
	'j': {'0', 3},
	'k': {'-', 2},
	'L': {'0', 3},
	'l': {'-', 2},
	'N': {'0', 9},
	'u': {'-', 0},
	'w': {'-', 0},
	'Y': {'0', 4},
}

func applyFlags(flags, s string) string {
	switch flags {
	case "^":
		return strings.ToUpper(s)
	case "#":
		if isUpperRE(s) {
			return strings.ToLower(s)
		}
		return strings.ToUpper(s)
	default:
		return s
	}
}

func convert(t time.Time, c rune, flags, width string) interface{} { // nolint: gocyclo
	switch c {

	// Date
	case 'Y':
		return t.Year()
	case 'y':
		return t.Year() % 100
	case 'C':
		return t.Year() / 100

	case 'm':
		return t.Month()
	case 'B':
		return t.Month().String()
	case 'b', 'h':
		return t.Month().String()[:3]

	case 'd', 'e':
		return t.Day()

	case 'j':
		return t.YearDay()

	// Time
	case 'H', 'k':
		return t.Hour()
	case 'I', 'l':
		return (t.Hour()+11)%12 + 1
	case 'M':
		return t.Minute()
	case 'S':
		return t.Second()
	case 'L':
		return t.Nanosecond() / 1e6
	case 'N':
		ns := t.Nanosecond()
		if len(width) > 0 {
			w, _ := strconv.Atoi(width) // nolint: gas
			if w <= 9 {
				return fmt.Sprintf("%09d", ns)[:w]
			}
			return fmt.Sprintf(fmt.Sprintf("%%09d%%0%dd", w-9), ns, 0)
		}
		return ns

	case 'P':
		return amPmLowerTable[t.Hour() < 12]
	case 'p':
		return amPmTable[t.Hour() < 12]

	// Time zone
	case 'z':
		_, offset := t.Zone()
		sign := '+'
		if offset < 0 {
			offset, sign = -offset, '-'
		}
		var (
			h = offset / 3600
			m = (offset / 60) % 60
			s = offset % 60
		)
		if flags == ":::" {
			switch {
			case s != 0:
				flags = "::"
			case m != 0:
				flags = ":"
			default:
				flags = "H" // not a real flag; only used to talk to next switch
			}
		}
		switch flags {
		case "H":
			return fmt.Sprintf("%c%02d", sign, h)
		case ":":
			return fmt.Sprintf("%c%02d:%02d", sign, h, m)
		case "::":
			return fmt.Sprintf("%c%02d:%02d:%02d", sign, h, m, s)
		default:
			return fmt.Sprintf("%c%02d%02d", sign, h, m)
		}
	case 'Z':
		z, _ := t.Zone()
		return z

	// Weekday
	case 'A':
		return t.Weekday().String()
	case 'a':
		return t.Weekday().String()[:3]
	case 'u':
		return (t.Weekday()+6)%7 + 1
	case 'w':
		return t.Weekday()

	// ISO week and year
	case 'G':
		y, _ := t.ISOWeek()
		return y
	case 'g':
		y, _ := t.ISOWeek()
		return y % 100
	case 'V':
		_, wn := t.ISOWeek()
		return wn

	// Ruby week
	case 'U':
		// day of year of first day of week (might be negative)
		d := t.YearDay() - int(t.Weekday())
		return (d + 6) / 7
	case 'W':
		// day of year of first day of (Monday-based) week
		d := t.YearDay() - int(t.Weekday()) + 1
		if t.Weekday() == time.Sunday {
			d -= 7
		}
		return (d + 6) / 7

	// Epoch seconds
	case 's':
		return t.Unix()
	case 'Q':
		return t.UnixNano() / 1000

	// Literals
	case 'n':
		return "\n"
	case 't':
		return "\t"
	case '%':
		return "%"

	// Combinations
	case 'c':
		// date and time (%a %b %e %T %Y)
		h, m, s := t.Clock()
		return fmt.Sprintf("%s %s %2d %02d:%02d:%02d %04d", t.Weekday().String()[:3], t.Month().String()[:3], t.Day(), h, m, s, t.Year())
	case 'D', 'x':
		// Date (%m/%d/%y)
		y, m, d := t.Date()
		return fmt.Sprintf("%02d/%02d/%02d", m, d, y%100)
	case 'F':
		// The ISO 8601 date format (%Y-%m-%d)
		y, m, d := t.Date()
		return fmt.Sprintf("%04d-%02d-%02d", y, m, d)
	case 'v':
		// VMS date (%e-%^b-%4Y)
		return fmt.Sprintf("%2d-%s-%04d", t.Day(), strings.ToUpper(t.Month().String()[:3]), t.Year())
	case 'r':
		// 12-hour time (%I:%M:%S %p)
		h, m, s := t.Clock()
		h12 := (h+11)%12 + 1
		return fmt.Sprintf("%02d:%02d:%02d %s", h12, m, s, amPmTable[h < 12])
	case 'R':
		// 24-hour time (%H:%M)
		h, m, _ := t.Clock()
		return fmt.Sprintf("%02d:%02d", h, m)
	case 'T', 'X':
		// 24-hour time (%H:%M:%S)
		h, m, s := t.Clock()
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	case '+':
		// date(1) (%a %b %e %H:%M:%S %Z %Y)
		s, _ := Strftime("%a %b %e %H:%M:%S %Z %Y", t) // nolint: gas
		return s
	default:
		return string([]rune{'%', c})
	}
}
