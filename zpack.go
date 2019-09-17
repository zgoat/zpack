package zpack

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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
func Pack(data map[string]map[string]string) error {
	for out, content := range data {
		fp, err := os.Create(out)
		if err != nil {
			return err
		}
		defer func() { fp.Close() }()

		err = Header(fp, filepath.Dir(out))
		if err != nil {
			return err
		}

		for varname, files := range content {
			st, err := os.Stat(files)
			if err != nil {
				return err
			}

			if st.IsDir() {
				err = Dir(fp, varname, files)
			} else {
				err = File(fp, varname, files)
			}
		}

		err = fp.Close()
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
	_, err = fp.Write([]byte("package " + pkg + "\n\n"))
	return err
}

// File writes a single file as a variable.
func File(fp io.Writer, varname, path string) error {
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(fp, "var %s = %s\n", varname, asbyte(d))
	return err
}

// Dir recursively writes all files in a directory as variables.
func Dir(fp io.Writer, varname, dir string) error {
	_, err := fp.Write([]byte("var " + varname + " = map[string][]byte{\n"))
	if err != nil {
		return err
	}

	err = filepath.Walk(dir, func(path string, st os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if st.IsDir() {
			return nil
		}

		d, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(fp, "\t\"%s\": %s,\n\n", path, asbyte(d))
		return err
	})
	if err != nil {
		return err
	}

	_, err = fp.Write([]byte("}\n\n"))
	return err
}

func asbyte(s []byte) string {
	var b strings.Builder
	for i, c := range s {
		if i%19 == 0 {
			b.WriteString("\n\t\t")
		}
		b.WriteString(fmt.Sprintf("%#x, ", c))
	}
	return "[]byte{" + b.String() + "}"
}
