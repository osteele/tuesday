# Tuesday: Ruby-Compatible Strftime for Go

 [![][travis-svg]][travis-url]
 [![][appveyor-svg]][appveyor-url]
 [![][coveralls-svg]][coveralls-url]
 [![][go-report-card-svg]][go-report-card-url]
 [![][godoc-svg]][godoc-url]
 [![][license-svg]][license-url]

This package provides a function `Strftime` that is compatible with Ruby's [`Time.strftime`](https://ruby-doc.org/core-2.4.1/Time.html#method-i-strftime).

It provides additional flags and conversions beyond C stdlib-like `strftime`s:

* padding flags, *e.g.* `%-m`, `%_m`, `%0e`
* case change flags, *e.g.* `%^A`, `%#b`
* field widths: `%03e`, `%3N`, `%9N`
* Ruby-specific conversions such as `%s`, `%N`, `%:z`, `%::z`

It was developed for use with in [Liquid](https://github.com/osteele/liquid) and [Gojekyll](https://github.com/osteele/gojekyll).

## Install

`go get gopkg.in/osteele/tuesday.v1` # latest snapshot

`go get -u github.com/osteele/tuesday` # development version

## References

* [Ruby Date.strftime](https://ruby-doc.org/stdlib-2.4.1/libdoc/date/rdoc/Date.html#method-i-strftime)
* [Ruby DateTime.strftime](https://ruby-doc.org/stdlib-2.4.1/libdoc/date/rdoc/DateTime.html#method-i-strftime)
* [Ruby Time.strftime](https://ruby-doc.org/core-2.4.1/Time.html#method-i-strftime)

## License

MIT License

[coveralls-url]: https://coveralls.io/r/osteele/tuesday?branch=master
[coveralls-svg]: https://img.shields.io/coveralls/osteele/tuesday.svg?branch=master

[godoc-url]: https://godoc.org/github.com/osteele/tuesday
[godoc-svg]: https://godoc.org/github.com/osteele/tuesday?status.svg

[license-url]: https://github.com/osteele/tuesday/blob/master/LICENSE
[license-svg]: https://img.shields.io/badge/license-MIT-blue.svg

[go-report-card-url]: https://goreportcard.com/report/github.com/osteele/tuesday
[go-report-card-svg]: https://goreportcard.com/badge/github.com/osteele/tuesday

[travis-url]: https://travis-ci.org/osteele/tuesday
[travis-svg]: https://img.shields.io/travis/osteele/tuesday.svg?branch=master

[appveyor-url]: https://ci.appveyor.com/project/osteele/tuesday
[appveyor-svg]: https://ci.appveyor.com/api/projects/status/y9cyh4e30yjxshtm?svg=true
