// tomltestgen retrieves a given version of the language-agnostic TOML test suite in
// https://github.com/BurntSushi/toml-test and generates go-toml unit tests.
//
// Within the go-toml package, run `go generate`.  Otherwise, use:
//
//	go run github.com/pelletier/go-toml/cmd/tomltestgen -o toml_testgen_test.go
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type invalid struct {
	Name  string
	Input string
}

type valid struct {
	Name    string
	Input   string
	JsonRef string
}

type testsCollection struct {
	Ref       string
	Timestamp string
	Invalid   []invalid
	Valid     []valid
	Count     int
}

const srcTemplate = "// Generated by tomltestgen for toml-test ref {{.Ref}} on {{.Timestamp}}\n" +
	"package toml_test\n" +
	" import (\n" +
	"	\"testing\"\n" +
	")\n" +

	"{{range .Invalid}}\n" +
	"func TestTOMLTest_Invalid_{{.Name}}(t *testing.T) {\n" +
	"	input := {{.Input|gostr}}\n" +
	"	testgenInvalid(t, input)\n" +
	"}\n" +
	"{{end}}\n" +
	"\n" +
	"{{range .Valid}}\n" +
	"func TestTOMLTest_Valid_{{.Name}}(t *testing.T) {\n" +
	"   input := {{.Input|gostr}}\n" +
	"   jsonRef := {{.JsonRef|gostr}}\n" +
	"   testgenValid(t, input, jsonRef)\n" +
	"}\n" +
	"{{end}}\n"

func kebabToCamel(kebab string) string {
	camel := ""
	nextUpper := true
	for _, c := range kebab {
		if nextUpper {
			camel += strings.ToUpper(string(c))
			nextUpper = false
		} else if c == '-' {
			nextUpper = true
		} else if c == '/' {
			nextUpper = true
			camel += "_"
		} else {
			camel += string(c)
		}
	}
	return camel
}

func templateGoStr(input string) string {
	return strconv.Quote(input)
}

var (
	ref = flag.String("r", "master", "git reference")
	out = flag.String("o", "", "output file")
)

func usage() {
	_, _ = fmt.Fprintf(os.Stderr, "usage: tomltestgen [flags]\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	collection := testsCollection{
		Ref:       *ref,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	dirContent, _ := filepath.Glob("tests/invalid/**/*.toml")
	for _, f := range dirContent {
		filename := strings.TrimPrefix(f, "tests/valid/")
		name := kebabToCamel(strings.TrimSuffix(filename, ".toml"))

		log.Printf("> [%s] %s\n", "invalid", name)

		tomlContent, err := os.ReadFile(f)
		if err != nil {
			fmt.Printf("failed to read test file: %s\n", err)
			os.Exit(1)
		}

		collection.Invalid = append(collection.Invalid, invalid{
			Name:  name,
			Input: string(tomlContent),
		})
		collection.Count++
	}

	dirContent, _ = filepath.Glob("tests/valid/**/*.toml")
	for _, f := range dirContent {
		filename := strings.TrimPrefix(f, "tests/valid/")
		name := kebabToCamel(strings.TrimSuffix(filename, ".toml"))

		log.Printf("> [%s] %s\n", "valid", name)

		tomlContent, err := os.ReadFile(f)
		if err != nil {
			fmt.Printf("failed reading test file: %s\n", err)
			os.Exit(1)
		}

		filename = strings.TrimSuffix(f, ".toml")
		jsonContent, err := os.ReadFile(filename + ".json")
		if err != nil {
			fmt.Printf("failed reading validation json: %s\n", err)
			os.Exit(1)
		}

		collection.Valid = append(collection.Valid, valid{
			Name:    name,
			Input:   string(tomlContent),
			JsonRef: string(jsonContent),
		})
		collection.Count++
	}

	log.Printf("Collected %d tests from toml-test\n", collection.Count)

	funcMap := template.FuncMap{
		"gostr": templateGoStr,
	}
	t := template.Must(template.New("src").Funcs(funcMap).Parse(srcTemplate))
	buf := new(bytes.Buffer)
	err := t.Execute(buf, collection)
	if err != nil {
		panic(err)
	}
	outputBytes, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	if *out == "" {
		fmt.Println(string(outputBytes))
		return
	}

	err = os.WriteFile(*out, outputBytes, 0o644)
	if err != nil {
		panic(err)
	}
}
