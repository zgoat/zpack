package zpack

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Pack data to a Go file.
//
// The data is a map where the key is the filename to store the packed data, and
// the contents a variable -> contents mapping.
//
// Example:
//
//     err := zpack.Pack(map[string]map[string]string{
//         "./db/pack.go": map[string]string{
//             "Schema": "./db/schema.sql",
//         },
//         "./handlers/pack.go": map[string]string{
//             "packPublic": "./public",
//             "packTpl":    "./tpl",
//         },
//     })
//     if err != nil {
//         fmt.Fprintln(os.Stderr, err)
//         os.Exit(1)
//     }
//
// The ignore patterns are matched by strings.HasSuffix().
func Pack(data map[string]map[string]string, ignore ...string) error {
	for out, content := range data {
		// TODO: be atomic; that is, we don't want to clobber anything existing
		// unless we're sure we'll be creating valid Go files.
		fp, err := os.Create(out)
		if err != nil {
			return err
		}
		defer func() { fp.Close() }()

		err = Header(fp, filepath.Base(filepath.Dir(out)))
		if err != nil {
			return err
		}

		var varnames []string
		for v := range content {
			varnames = append(varnames, v)
		}
		sort.Strings(varnames)

		for _, varname := range varnames {
			files := content[varname]
			st, err := os.Stat(files)
			if err != nil {
				return err
			}

			if st.IsDir() {
				err = Dir(fp, varname, files, ignore...)
			} else {
				err = File(fp, varname, files)
			}
			if err != nil {
				return err
			}
		}

		err = fp.Close()
		if err != nil {
			return err
		}
		err = Format(out)
		if err != nil {
			return err
		}
	}
	return nil
}

// Header writes a file header, which is a code generation comment and package
// declaration.
func Header(fp io.Writer, pkg string) error {
	_, err := fp.Write([]byte("// Code generated by pack.go; DO NOT EDIT.\n\n"))
	if err != nil {
		return err
	}
	_, err = fp.Write([]byte("package " + pkg + "\n\nimport (" + `
		"bytes"
		"compress/zlib"
		"encoding/base64"
		"io/ioutil"
	` + ")\n\nvar _, _, _, _ = zlib.BestSpeed, base64.NoPadding, ioutil.Discard, bytes.Join\n\n"))
	return err
}

// File writes a single file as a variable.
func File(fp io.Writer, varname, path string) error {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(fp, "var %s = %s\n", varname, enc(d))
	return err
}

// Dir recursively writes all files in a directory as variables.
//
// The ignore patterns are matched by strings.HasSuffix().
func Dir(fp io.Writer, varname, dir string, ignore ...string) error {
	_, err := fp.Write([]byte("var " + varname + " = map[string][]byte{\n"))
	if err != nil {
		return err
	}

	// Make sure to walk the contents of the directory in case of a link,
	// instead of the link itself.
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}

	err = filepath.Walk(dir, func(path string, st os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk error: %s", err)
		}
		if st.IsDir() {
			return nil
		}
		for _, ig := range ignore {
			// Special case to exclude VCS "keep" files.
			if strings.HasSuffix(path, ig) {
				return nil
			}
		}

		d, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("ioutil.ReadFile(%q): %s", path, err)
		}

		_, err = fmt.Fprintf(fp, "\t\"%s\": %s,\n", path, enc(d))
		return err
	})
	if err != nil {
		return err
	}

	_, err = fp.Write([]byte("}\n\n"))
	return err
}

// Format the given file with gofmt.
func Format(path string) error {
	// TODO: can also use "go/format.Source(data)"
	out, err := exec.Command("gofmt", "-w", path).CombinedOutput()
	if err != nil {
		return fmt.Errorf("gofmt: %s: %s", err, string(out))
	}
	return nil
}

// Varname replaces any sequence of invalid identifier characters with an _.
func Varname(s string) string {
	if s == "" {
		// TODO: possibly random generate?
		return ""
	}

	var n []rune
	var replaced bool
	for _, c := range s {
		if unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' {
			n = append(n, c)
			replaced = false
			continue
		}

		if !replaced {
			n = append(n, '_')
			replaced = true
		}
	}

	if n[0] != '_' && !unicode.IsLetter(n[0]) {
		n = append([]rune{'v'}, n...)
	}

	return string(n)
}

func enc(s []byte) string {
	if bytes.IndexByte(s, 0) == -1 && utf8.Valid(s) {
		return fmt.Sprintf("[]byte(`%s`)", bytes.Replace(s, []byte("`"), []byte("` + \"`\" + `"), -1))
	}

	// Compress files larger than 100K
	if len(s) > 1024*100 {
		var b bytes.Buffer
		w := zlib.NewWriter(&b)
		w.Write(s)
		w.Close()

		return fmt.Sprintf(`func() []byte {
			z, err := base64.StdEncoding.DecodeString("%s")
			if err != nil {
				panic(err)
			}
			r, err := zlib.NewReader(bytes.NewReader(z))
			if err != nil {
				panic(err)
			}

			s, err := ioutil.ReadAll(r)
			if err != nil {
				panic(err)
			}
			r.Close()
			return s
		}()`, base64.StdEncoding.EncodeToString(b.Bytes()))
	}

	// TODO: maybe wrap?
	return fmt.Sprintf(`func() []byte {
		s, err := base64.StdEncoding.DecodeString("%s")
		if err != nil {
			panic(err)
		}
		return s
	}()`, base64.StdEncoding.EncodeToString(s))
}
