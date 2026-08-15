// Package tuesday implements strftime formatting compatible with Ruby 3.4.
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

const maxFieldWidth = 1 << 20

var directiveRE = regexp.MustCompile(`%([-_^#0]*)(\d+)?([EO]?)(:{1,3})?([a-zA-Z\+nt%])`)

// Formatter is an immutable, reusable compiled strftime format. A Formatter is
// safe for concurrent use by multiple goroutines.
type Formatter struct {
	parts      []formatPart
	formatSize int
}

type partKind uint8

const (
	literalPart partKind = iota
	directivePart
)

type formatPart struct {
	kind      partKind
	literal   string
	directive directive
}

type directive struct {
	flags    string
	width    int
	hasWidth bool
	colons   string
	code     rune
}

// Compile parses a strftime format for repeated use. Unsupported directives
// are retained as literal text. Compile returns an error when a supported
// directive requests a field width large enough to risk excessive allocation.
func Compile(format string) (*Formatter, error) {
	formatter := &Formatter{formatSize: len(format)}
	last := 0
	for _, match := range directiveRE.FindAllStringSubmatchIndex(format, -1) {
		formatter.appendLiteral(format[last:match[0]])
		raw := format[match[0]:match[1]]
		flags := capture(format, match[2], match[3])
		widthText := capture(format, match[4], match[5])
		colons := capture(format, match[8], match[9])
		code, _ := utf8.DecodeRuneInString(format[match[10]:match[11]])

		if !supportedConversion(code) || colons != "" && code != 'z' {
			formatter.appendLiteral(raw)
			last = match[1]
			continue
		}

		d := directive{flags: flags, colons: colons, code: code}
		if widthText != "" {
			width, err := strconv.Atoi(widthText)
			if err != nil || width > maxFieldWidth {
				return nil, fmt.Errorf("tuesday: field width in %q exceeds %d", raw, maxFieldWidth)
			}
			d.width, d.hasWidth = width, true
		}
		formatter.parts = append(formatter.parts, formatPart{kind: directivePart, directive: d})
		last = match[1]
	}
	formatter.appendLiteral(format[last:])
	return formatter, nil
}

func capture(s string, start, end int) string {
	if start < 0 {
		return ""
	}
	return s[start:end]
}

func (f *Formatter) appendLiteral(s string) {
	if s == "" {
		return
	}
	if n := len(f.parts); n > 0 && f.parts[n-1].kind == literalPart {
		f.parts[n-1].literal += s
		return
	}
	f.parts = append(f.parts, formatPart{kind: literalPart, literal: s})
}

// Format formats t using the compiled format.
func (f *Formatter) Format(t time.Time) string {
	var builder strings.Builder
	builder.Grow(f.formatSize)
	for _, part := range f.parts {
		if part.kind == literalPart {
			builder.WriteString(part.literal)
			continue
		}
		builder.WriteString(formatDirective(part.directive, t))
	}
	return builder.String()
}

// Strftime formats t according to a Ruby-compatible strftime format.
// Unsupported directives are retained as literal text. Strftime returns an
// error if a supported directive requests a field width greater than 1 MiB.
func Strftime(format string, t time.Time) (string, error) {
	formatter, err := Compile(format)
	if err != nil {
		return "", err
	}
	return formatter.Format(t), nil
}

func supportedConversion(code rune) bool {
	switch code {
	case 'Y', 'y', 'C', 'm', 'B', 'b', 'h', 'd', 'e', 'j',
		'H', 'k', 'I', 'l', 'M', 'S', 'L', 'N', 'P', 'p',
		'z', 'Z', 'A', 'a', 'u', 'w', 'G', 'g', 'V', 'U', 'W',
		's', 'Q', 'n', 't', '%', 'c', 'D', 'x', 'F', 'v', 'r',
		'R', 'T', 'X', '+':
		return true
	default:
		return false
	}
}

var amPmTable = map[bool]string{true: "AM", false: "PM"}
var amPmLowerTable = map[bool]string{true: "am", false: "pm"}

type padding struct {
	char  rune
	width int
}

// The default for numeric directives not listed here is zero-padding to two characters.
var defaultPadding = map[rune]padding{
	'e': {' ', 2},
	'j': {'0', 3},
	'k': {' ', 2},
	'l': {' ', 2},
	'G': {'0', 4},
	'Q': {'0', 0},
	's': {'0', 0},
	'u': {'0', 0},
	'w': {'0', 0},
	'Y': {'0', 4},
}

func formatDirective(d directive, t time.Time) string {
	if d.code == 'z' {
		return formatZone(t, d)
	}
	if d.code == 'L' || d.code == 'N' {
		defaultWidth := 9
		if d.code == 'L' {
			defaultWidth = 3
		}
		return formatFraction(t.Nanosecond(), d, defaultWidth)
	}

	value := convert(t, d.code)
	switch value.kind {
	case textValue:
		text := applyCaseFlags(d.flags, d.code, value.text)
		return padString(text, d)
	case numberValue:
		return formatNumber(value.number, d)
	default:
		panic("unsupported compiled strftime directive")
	}
}

func formatNumber(number int64, d directive) string {
	pad, width := rune('0'), 2
	if defaults, ok := defaultPadding[d.code]; ok {
		pad, width = defaults.char, defaults.width
	}
	if d.hasWidth {
		width = d.width
	}
	if strings.Contains(d.flags, "-") {
		width = 0
	} else {
		pad = paddingFlag(d.flags, pad)
	}

	var result string
	switch {
	case width == 0:
		result = strconv.FormatInt(number, 10)
	case pad == '0':
		result = fmt.Sprintf("%0*d", width, number)
	default:
		result = fmt.Sprintf("%*d", width, number)
	}
	return applyCaseFlags(d.flags, d.code, result)
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

func padString(s string, d directive) string {
	if !d.hasWidth || strings.Contains(d.flags, "-") || len(s) >= d.width {
		return s
	}
	pad := paddingFlag(d.flags, ' ')
	return strings.Repeat(string(pad), d.width-len(s)) + s
}

func formatFraction(ns int, d directive, defaultWidth int) string {
	width := defaultWidth
	if d.hasWidth {
		width = d.width
	}
	digits := fmt.Sprintf("%09d", ns)
	if width <= len(digits) {
		return digits[:width]
	}
	return digits + strings.Repeat("0", width-len(digits))
}

func formatZone(t time.Time, d directive) string {
	_, offset := t.Zone()
	sign := '+'
	if offset < 0 {
		offset, sign = -offset, '-'
	}
	h, m, s := offset/3600, (offset/60)%60, offset%60

	var body string
	defaultWidth := 5
	switch d.colons {
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
		body = fmt.Sprintf("%03d", h*100+m)
	}

	width := defaultWidth
	if d.hasWidth && d.width > defaultWidth {
		width = d.width
	}
	pad := zonePaddingFlag(d.flags)
	if n := width - 1 - len(body); n > 0 {
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

type valueKind uint8

const (
	unsupportedValue valueKind = iota
	textValue
	numberValue
)

type conversionValue struct {
	kind   valueKind
	text   string
	number int64
}

func text(s string) conversionValue {
	return conversionValue{kind: textValue, text: s}
}

func number(n int64) conversionValue {
	return conversionValue{kind: numberValue, number: n}
}

func convert(t time.Time, code rune) conversionValue { // nolint: gocyclo
	switch code {
	case 'Y':
		return number(int64(t.Year()))
	case 'y':
		return number(int64(t.Year() % 100))
	case 'C':
		return number(int64(t.Year() / 100))
	case 'm':
		return number(int64(t.Month()))
	case 'B':
		return text(t.Month().String())
	case 'b', 'h':
		return text(t.Month().String()[:3])
	case 'd', 'e':
		return number(int64(t.Day()))
	case 'j':
		return number(int64(t.YearDay()))
	case 'H', 'k':
		return number(int64(t.Hour()))
	case 'I', 'l':
		return number(int64((t.Hour()+11)%12 + 1))
	case 'M':
		return number(int64(t.Minute()))
	case 'S':
		return number(int64(t.Second()))
	case 'P':
		return text(amPmLowerTable[t.Hour() < 12])
	case 'p':
		return text(amPmTable[t.Hour() < 12])
	case 'Z':
		zone, _ := t.Zone()
		return text(zone)
	case 'A':
		return text(t.Weekday().String())
	case 'a':
		return text(t.Weekday().String()[:3])
	case 'u':
		return number(int64((t.Weekday()+6)%7 + 1))
	case 'w':
		return number(int64(t.Weekday()))
	case 'G':
		year, _ := t.ISOWeek()
		return number(int64(year))
	case 'g':
		year, _ := t.ISOWeek()
		return number(int64(year % 100))
	case 'V':
		_, week := t.ISOWeek()
		return number(int64(week))
	case 'U':
		day := t.YearDay() - int(t.Weekday())
		return number(int64((day + 6) / 7))
	case 'W':
		day := t.YearDay() - int(t.Weekday()) + 1
		if t.Weekday() == time.Sunday {
			day -= 7
		}
		return number(int64((day + 6) / 7))
	case 's':
		return number(t.Unix())
	case 'Q':
		return number(t.UnixMilli())
	case 'n':
		return text("\n")
	case 't':
		return text("\t")
	case '%':
		return text("%")
	case 'c':
		hour, minute, second := t.Clock()
		return text(fmt.Sprintf("%s %s %2d %02d:%02d:%02d %04d", t.Weekday().String()[:3], t.Month().String()[:3], t.Day(), hour, minute, second, t.Year()))
	case 'D', 'x':
		year, month, day := t.Date()
		return text(fmt.Sprintf("%02d/%02d/%02d", month, day, year%100))
	case 'F':
		year, month, day := t.Date()
		return text(fmt.Sprintf("%04d-%02d-%02d", year, month, day))
	case 'v':
		return text(fmt.Sprintf("%2d-%s-%04d", t.Day(), strings.ToUpper(t.Month().String()[:3]), t.Year()))
	case 'r':
		hour, minute, second := t.Clock()
		hour12 := (hour+11)%12 + 1
		return text(fmt.Sprintf("%02d:%02d:%02d %s", hour12, minute, second, amPmTable[hour < 12]))
	case 'R':
		hour, minute, _ := t.Clock()
		return text(fmt.Sprintf("%02d:%02d", hour, minute))
	case 'T', 'X':
		hour, minute, second := t.Clock()
		return text(fmt.Sprintf("%02d:%02d:%02d", hour, minute, second))
	case '+':
		hour, minute, second := t.Clock()
		zone, _ := t.Zone()
		return text(fmt.Sprintf("%s %s %2d %02d:%02d:%02d %s %04d", t.Weekday().String()[:3], t.Month().String()[:3], t.Day(), hour, minute, second, zone, t.Year()))
	default:
		return conversionValue{kind: unsupportedValue}
	}
}
