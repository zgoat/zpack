package zpack

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestPack(t *testing.T) {
	tmp, err := ioutil.TempDir("", "testpack")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { os.RemoveAll(tmp) }()

	err = Pack(map[string]map[string]string{
		tmp + "/pack.go": map[string]string{
			"zpack":    "./zpack.go",
			"zpackDir": ".",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile(tmp+"/pack_test.go",
		[]byte(fmt.Sprintf(test, filepath.Base(tmp), "zpack")),
		0644)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: Check files hashes
	// TODO: test large compressed files.
	// cmd := exec.Command("go", "test")
	// cmd.Dir = tmp
	// out, err := cmd.CombinedOutput()
	// if err != nil {
	// 	t.Fatalf("go test: %s: %s", err, out)
	// }
}

func TestVaraname(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"ab", "ab"},
		{"a.b", "a_b"},
		{"a..b", "a_b"},
		{"a__b", "a__b"},
		{"1ab", "v1ab"},
		{"αβ", "αβ"},
		{"α.β", "α_β"},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			out := Varname(tt.in)
			if out != tt.want {
				t.Errorf("\nout:  %q\nwant: %q", out, tt.want)
			}
		})
	}
}

const test = `
package %s

import (
    "fmt"
    "testing"
)

func TestA(t *testing.T) {
    fmt.Println(string(%s))
}`
