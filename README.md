logr
====

logr is a simplistic type which implements log rotating, suitable for use with Go's `log` package.

[![Build Status](https://travis-ci.org/vrischmann/logr.svg?branch=master)](https://travis-ci.org/vrischmann/logr)
[![GoDoc](https://godoc.org/github.com/vrischmann/logr?status.svg)](https://godoc.org/github.com/vrischmann/logr)

Usage
-----

Daily rotation.

```go
w := logr.NewWriter("/var/log/mylog.log", &logr.Options{
    RotateDaily: true,
})
log.SetOutput(w)

log.Println("foobar")
```

Maximum size rotation.

```go
// Rotate every 500 Mib.
w := logr.NewWriter("/var/log/mylog.log", &logr.Options{
    MaximumSize: 1024 * 1024 * 500,
})
log.SetOutput(w)

log.Println("foobar")
```

Compress the rotated file.

```go
// Rotate every 500 Mib then compress the file.
w := logr.NewWriter("/var/log/mylog.log", &logr.Options{
    MaximumSize: 1024 * 1024 * 500,
    Compress: true,
})
log.SetOutput(w)

log.Println("foobar")
```

Future work
-----------

  * Background rotation
