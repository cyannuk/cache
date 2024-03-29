package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	var buffer bytes.Buffer

	flags := flag.NewFlagSet("github.com/cyannuk/cache", flag.ContinueOnError)
	flags.SetOutput(&buffer)

	keyType := flags.String("key-type", "", "Cache key type")
	valueType := flags.String("value-type", "", "Cache value type")
	pkgPath := flags.String("package", "", "Cache package path")

	err := flags.Parse(os.Args[1:])
	if err != nil {
		if err == flag.ErrHelp {
			fmt.Println(buffer.String())
			return
		}
		panic(err)
	}

	funcMap := template.FuncMap{
		"capitalize":   capitalize,
		"uncapitalize": uncapitalize,
		"join":         join,
	}
	tmpl, err := template.New("cacheTmpl").Funcs(funcMap).Parse(cacheTmpl)
	if err != nil {
		panic(err)
	}

	buffer.Grow(64 * 1024)

	importPath, typeName, cacheValueType := getTypeInfo(*valueType)
	packageName, fileName := getPath(*pkgPath, typeName)
	params := map[string]string{
		"Package":   packageName,
		"Import":    importPath,
		"KeyType":   *keyType,
		"ValueType": cacheValueType,
		"TypeName":  typeName,
	}

	if err := tmpl.Execute(&buffer, params); err != nil {
		panic(err)
	}

	if err := write(buffer.Bytes(), fileName); err != nil {
		panic(err)
	}
}

func getPath(pkgPath string, typeName string) (string, string) {
	if strings.HasSuffix(pkgPath, ".go") {
		i := strings.LastIndexByte(pkgPath, '/')
		if i >= 0 {
			return filepath.Base(pkgPath[:i]), pkgPath
		} else {
			var packageName string
			dir, _ := os.Getwd()
			if _, err := os.Stat(path.Join(dir, "go.mod")); err == nil {
				packageName = "main"
			} else {
				packageName = filepath.Base(dir)
			}
			return packageName, filepath.Join(dir, pkgPath)
		}
	} else {
		return filepath.Base(pkgPath), filepath.Join(pkgPath, strings.ToLower(typeName)+".cache.go")
	}
}

func getTypeInfo(valueType string) (string, string, string) {
	importPath := ""
	prefix := ""
	if strings.HasPrefix(valueType, "*") {
		valueType = valueType[1:]
		prefix = "*"
	}
	typeName := valueType
	cacheValueType := valueType
	i := strings.LastIndexByte(valueType, '/')
	if i >= 0 {
		p := valueType[:i]
		importPath = `"` + p + `"`
		typeName = valueType[i+1:]
		cacheValueType = path.Base(p) + "." + typeName
	}
	return importPath, typeName, prefix + cacheValueType
}

func write(bytes []byte, fileName string) error {
	if b, err := format.Source(bytes); err != nil {
		_ = writeFile(bytes, fileName)
		return fmt.Errorf("format template: %w", err)
	} else {
		return writeFile(b, fileName)
	}
}

func writeFile(b []byte, fileName string) error {
	dir := path.Dir(fileName)
	if _, err := os.Stat(dir); err != nil {
		_ = os.MkdirAll(dir, 0)
	}
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	_, err = file.Write(b)
	_ = file.Close()
	if err != nil {
		return fmt.Errorf("write template: %w", err)
	}
	return nil
}

func join(strings ...string) string {
	b := make([]byte, 0, 512)
	for _, s := range strings {
		b = append(b, s...)
	}
	return string(b)
}

func capitalize(str string) string {
	if str == "" {
		return str
	}
	bb := []byte(str)
	var b byte
	var i int
	for i, b = range bb {
		if b != '_' {
			break
		}
	}
	if b >= 'a' && b <= 'z' {
		buffer := make([]byte, 0, len(bb))
		buffer = append(buffer, b-0x20)
		buffer = append(buffer, bb[i+1:]...)
		return string(buffer)
	} else {
		if i == 0 {
			return str
		} else {
			return string(bb[i:])
		}
	}
}

func uncapitalize(str string) string {
	if str == "" {
		return str
	}
	bb := []byte(str)
	var b byte
	var i int
	for i, b = range bb {
		if b != '_' {
			break
		}
	}
	if b >= 'A' && b <= 'Z' {
		buffer := make([]byte, 0, len(bb))
		buffer = append(buffer, b+0x20)
		buffer = append(buffer, bb[i+1:]...)
		return string(buffer)
	} else {
		if i == 0 {
			return str
		} else {
			return string(bb[i:])
		}
	}
}
