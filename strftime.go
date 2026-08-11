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
		m := re.FindStringSubmatch(directive)
		flags, width, colons := m[1], m[2], m[4]
		conversion, _ := utf8.DecodeRuneInString(m[5])

		if colons != "" && conversion != 'z' {
			return directive
		}
		if conversion == 'z' {
			return formatZone(t, flags, width, colons)
		}
		if conversion == 'L' || conversion == 'N' {
			defaultWidth := 9
			if conversion == 'L' {
				defaultWidth = 3
			}
			return formatFraction(t.Nanosecond(), width, defaultWidth)
		}

		c := convert(t, conversion)
		if _, ok := c.(unsupportedConversion); ok {
			return directive
		}
		if s, ok := c.(string); ok {
			s = applyCaseFlags(flags, conversion, s)
			return padString(s, flags, width)
		}

		pad, w := '0', 2
		if f, ok := defaultPadding[conversion]; ok {
			pad, w = f.c, f.w
		}
		if width != "" {
			var err error
			w, err = strconv.Atoi(width)
			if err != nil {
				return directive
			}
		}
		if strings.Contains(flags, "-") {
			w = 0
		} else {
			pad = paddingFlag(flags, pad)
		}
		var s string
		switch {
		case w == 0:
			s = fmt.Sprintf("%d", c)
		case pad == '0':
			s = fmt.Sprintf("%0*d", w, c)
		default:
			s = fmt.Sprintf("%*d", w, c)
		}
		return applyCaseFlags(flags, conversion, s)
	}), nil
}

var re = regexp.MustCompile(`%([-_^#0]*)(\d+)?([EO]?)(:{1,3})?([a-zA-Z\+nt%])`)

type unsupportedConversion struct{}

var amPmTable = map[bool]string{true: "AM", false: "PM"}
var amPmLowerTable = map[bool]string{true: "am", false: "pm"}

// Default padding character and width, by conversion rune.
// The default default is pad='0', width=2
var defaultPadding = map[rune]struct {
	c rune
	w int
}{
	'e': {'-', 2},
	'j': {'0', 3},
	'k': {'-', 2},
	'l': {'-', 2},
	'G': {'0', 4},
	'Q': {'-', 0},
	's': {'-', 0},
	'u': {'-', 0},
	'w': {'-', 0},
	'Y': {'0', 4},
}

func applyCaseFlags(flags string, conversion rune, s string) string {
	if strings.Contains(flags, "#") && strings.ContainsRune("ABPZabhp", conversion) {
		upper, lower := strings.ToUpper(s), strings.ToLower(s)
		if s == upper && s != lower {
			return lower
		}
		return upper
	}
	if strings.Contains(flags, "^") {
		return strings.ToUpper(s)
	}
	return s
}

func paddingFlag(flags string, defaultPad rune) rune {
	pad := defaultPad
	for _, flag := range flags {
		switch flag {
		case '_':
			pad = ' '
		case '0':
			pad = '0'
		}
	}
	return pad
}

func padString(s, flags, width string) string {
	if width == "" || strings.Contains(flags, "-") {
		return s
	}
	w, err := strconv.Atoi(width)
	if err != nil || len(s) >= w {
		return s
	}
	pad := paddingFlag(flags, ' ')
	return strings.Repeat(string(pad), w-len(s)) + s
}

func formatFraction(ns int, width string, defaultWidth int) string {
	w := defaultWidth
	if width != "" {
		var err error
		w, err = strconv.Atoi(width)
		if err != nil {
			return ""
		}
	}
	digits := fmt.Sprintf("%09d", ns)
	if w <= len(digits) {
		return digits[:w]
	}
	return digits + strings.Repeat("0", w-len(digits))
}

func formatZone(t time.Time, flags, width, colons string) string {
	_, offset := t.Zone()
	sign := '+'
	if offset < 0 {
		offset, sign = -offset, '-'
	}
	h, m, s := offset/3600, (offset/60)%60, offset%60

	var body string
	defaultWidth := 5
	switch colons {
	case ":":
		body, defaultWidth = fmt.Sprintf("%d:%02d", h, m), 6
	case "::":
		body, defaultWidth = fmt.Sprintf("%d:%02d:%02d", h, m, s), 9
	case ":::":
		switch {
		case s != 0:
			body, defaultWidth = fmt.Sprintf("%d:%02d:%02d", h, m, s), 9
		case m != 0:
			body, defaultWidth = fmt.Sprintf("%d:%02d", h, m), 6
		default:
			body, defaultWidth = strconv.Itoa(h), 3
		}
	default:
		body = strconv.Itoa(h*100 + m)
	}

	w := defaultWidth
	if width != "" {
		var err error
		w, err = strconv.Atoi(width)
		if err != nil {
			return ""
		}
	}
	pad := zonePaddingFlag(flags)
	if n := w - 1 - len(body); n > 0 {
		if pad == '0' {
			body = strings.Repeat("0", n) + body
		} else {
			return strings.Repeat(" ", n) + string(sign) + body
		}
	}
	return string(sign) + body
}

func zonePaddingFlag(flags string) rune {
	pad := '0'
	for _, flag := range flags {
		switch flag {
		case '_':
			pad = ' '
		case '-', '0':
			pad = '0'
		}
	}
	return pad
}

func convert(t time.Time, c rune) interface{} { // nolint: gocyclo
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
	case 'P':
		return amPmLowerTable[t.Hour() < 12]
	case 'p':
		return amPmTable[t.Hour() < 12]

	// Time zone
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

	// Epoch time
	case 's':
		return t.Unix()
	case 'Q':
		// Milliseconds since epoch (Ruby DateTime#strftime)
		return t.UnixMilli()

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
		return unsupportedConversion{}
	}
}
