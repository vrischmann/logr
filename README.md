logr
====

logr is a simplistic type which implements log rotating, suitable for use with Go's `log` package.

[![Build Status](https://travis-ci.org/vrischmann/logr.svg?branch=master)](https://travis-ci.org/vrischmann/logr)
[![GoDoc](https://godoc.org/github.com/vrischmann/logr?status.svg)](https://godoc.org/github.com/vrischmann/logr)

Usage
-----

Use it like this.

```go
w := logr.NewWriter("/var/log/mylog.log")
log.SetOutput(w)

log.Println("foobar")
```

Future works
------------

 * Custom time suffix formats
