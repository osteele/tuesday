# Tuesday: Ruby-Compatible Strftime for Go

[![Test badge][test-svg]][test-url]
[![Golangci-lint badge][golangci-lint-svg]][golangci-lint-url]
[![Coveralls badge][coveralls-svg]][coveralls-url]
[![Go Report Card badge][go-report-card-svg]][go-report-card-url]
[![Go Reference][go-reference-svg]][go-reference-url]
[![MIT License][license-svg]][license-url]

Tuesday formats Go `time.Time` values with Ruby-compatible `strftime` format
strings. It supports padding and case flags, field widths, fractional seconds,
epoch time, and colon-delimited timezone offsets.

Tuesday was developed for use by [Liquid](https://github.com/osteele/liquid)
and [Gojekyll](https://github.com/osteele/gojekyll).

## Install

```console
go get github.com/osteele/tuesday@latest
```

## Usage

```go
package main

import (
	"fmt"
	"time"

	"github.com/osteele/tuesday"
)

func main() {
	value := time.Date(2026, 8, 11, 15, 4, 5, 123456789, time.UTC)
	formatted, err := tuesday.Strftime("%Y-%m-%d %H:%M:%S.%3N %:z", value)
	if err != nil {
		panic(err)
	}
	fmt.Println(formatted)
}
```

Output:

```text
2026-08-11 15:04:05.123 +00:00
```

Compile formats that are used repeatedly. A compiled formatter is immutable
and safe for concurrent use:

```go
formatter, err := tuesday.Compile("%a, %b %d, %Y")
if err != nil {
	panic(err)
}
formatted := formatter.Format(value)
```

## Compatibility

Tuesday targets the formatting behavior of Ruby 3.4
[`Time#strftime`](https://docs.ruby-lang.org/en/3.4/Time.html). Its differential
test matrix also uses `DateTime#strftime` for `%Q`, which Ruby `Time` does not
implement. Tuesday additionally supports `%+` as an extension; Ruby `Time`
does not implement `%+`, while Ruby `DateTime` prints the UTC offset rather
than the location name.

- Month and weekday names use Go's English names, corresponding to Ruby's C
  locale. Locale-specific names are not supported; `E` and `O` modifiers use
  the corresponding unmodified conversion.
- `%Z` uses the name attached to the Go `time.Location`. A numeric offset alone
  does not imply a timezone abbreviation.
- Unsupported directives are copied to the result unchanged.
- Supported field widths greater than 1 MiB return an error to prevent
  excessive allocation.
- `%Q`, `%N`, `%L`, `%:z`, `%::z`, and `%:::z` follow the corresponding Ruby
  `DateTime`, fractional-second, and timezone conventions.

The Ruby differential test is skipped when Ruby is unavailable. The checked-in
Go tests remain sufficient to run the package test suite without Ruby.

## Development

```console
go test ./...
go test -fuzz=FuzzStrftime
go test -bench=. -benchmem
```

## References

- [Ruby 3.4 Time](https://docs.ruby-lang.org/en/3.4/Time.html)
- [Ruby 3.4 DateTime](https://docs.ruby-lang.org/en/3.4/DateTime.html)

## License

MIT License

[coveralls-url]: https://coveralls.io/r/osteele/tuesday?branch=main
[coveralls-svg]: https://img.shields.io/coveralls/osteele/tuesday.svg?branch=main
[go-reference-url]: https://pkg.go.dev/github.com/osteele/tuesday
[go-reference-svg]: https://pkg.go.dev/badge/github.com/osteele/tuesday.svg
[golangci-lint-url]: https://github.com/osteele/tuesday/actions?query=workflow%3Agolangci-lint
[golangci-lint-svg]: https://github.com/osteele/tuesday/actions/workflows/golangci-lint.yml/badge.svg
[license-url]: https://github.com/osteele/tuesday/blob/main/LICENSE
[license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
[go-report-card-url]: https://goreportcard.com/report/github.com/osteele/tuesday
[go-report-card-svg]: https://goreportcard.com/badge/github.com/osteele/tuesday
[test-url]: https://github.com/osteele/tuesday/actions?query=workflow%3Atest
[test-svg]: https://github.com/osteele/tuesday/actions/workflows/test.yml/badge.svg
