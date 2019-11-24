[![Build Status](https://travis-ci.org/zgoat/zpack.svg?branch=master)](https://travis-ci.org/zgoat/zpack)
[![codecov](https://codecov.io/gh/zgoat/zpack/branch/master/graph/badge.svg)](https://codecov.io/gh/zgoat/zpack)
[![GoDoc](https://godoc.org/github.com/zgoat/zpack?status.svg)](https://godoc.org/github.com/zgoat/zpack)

zpack is yet another way to pack static data in Go binaries.

Why? Because I don't like relying on external binaries, and many other solutions
do. zpack just writes data to the specified file as `[]byte()`.

Basic usage:

```go
// +build go_run_only

package main

import (
    "fmt"
    "os"

    "zgo.at/zpack"
)

func main() {
    err := zpack.Pack(map[string]map[string]string{
        // Pack ./db/schema.sql in ./db/pack.go as the variable "Schema".
        "./db/pack.go": map[string]string{
            "Schema": "./db/schema.sql",
        },

        // Pack all files in the "./public" and "./tpl" directories.
        "./handlers/pack.go": map[string]string{
            "packPublic": "./public",
            "packTpl":    "./tpl",
        },
    })
    if err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

Then `go generate ./...` and presto!
