zpack is yet another tool to pack static data in Go binaries.

Why? Because I don't like relying on external binaries, and most other tools do.


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
		"./db/pack.go": map[string]string{
			"Schema": "./db/schema.sql",
		},
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
