package main

import (
	"bytes"
	"encoding/json"
	"go/format"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:generate go run gen.go

var consoleServices []string = []string{
	"auth",
	"payments",
}

type Endpoint struct {
	Path     string
	Method   string
	FuncName string
	Request  map[string]interface{}
	Response map[string]interface{}
}

type ApiDef map[string][]Endpoint

func main() {
	data, err := ioutil.ReadFile("api.json")
	check(err)

	var api ApiDef
	check(json.Unmarshal(data, &api))

	genInfos := []struct {
		TemplatePath string
		OutputPath   string
		IsGo         bool
	}{
		// TODO set IsGo to true when formatting works correctly
		{TemplatePath: "templates/consoleweb.tmpl", OutputPath: "satellite/console/consoleweb/api.go"},
		// TODO set IsGo to true when formatting works correctly
		{TemplatePath: "templates/consoleapi.tmpl", OutputPath: "satellite/console/consoleweb/consoleapi/api.go"},
		{TemplatePath: "templates/web.tmpl", OutputPath: "web/satellite/src/api/api.ts"},
	}

	for _, genInfo := range genInfos {
		tmpl, err := template.New(filepath.Base(genInfo.TemplatePath)).Funcs(template.FuncMap{
			"sub": func(a, b int) int {
				return a - b
			},
			"strtitle": func(s string) string {
				return strings.Title(s)
			},
			"substr": func(s string, start, end int) string {
				return s[start:end]
			},
			"tolower": func(s string) string {
				return strings.ToLower(s)
			},
		}).ParseFiles(genInfo.TemplatePath)
		check(err)

		check(os.MkdirAll(filepath.Dir(genInfo.OutputPath), os.ModePerm))

		out, err := os.OpenFile(genInfo.OutputPath, os.O_CREATE, os.ModePerm)
		check(err)

		check(out.Truncate(0))

		if genInfo.IsGo {
			buf := bytes.NewBuffer([]byte{})
			check(tmpl.Execute(buf, struct {
				Services []string
				ApiDef   ApiDef
			}{
				Services: consoleServices,
				ApiDef:   api,
			}))
			// TODO do this without pulling everything into memory
			formattedBuf, err := format.Source(buf.Bytes())
			check(err)
			io.Copy(out, bytes.NewReader(formattedBuf))
		} else {
			check(tmpl.Execute(out, struct {
				Services []string
				ApiDef   ApiDef
			}{
				Services: consoleServices,
				ApiDef:   api,
			}))
		}
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func (e Endpoint) GetGoFuncName() (name string) {
	name = e.Method
	pathParts := strings.Split(strings.ReplaceAll(e.Path[1:], "/", "-"), "-")
	for i, part := range pathParts {
		pathParts[i] = strings.Title(strings.ToLower(part))
	}
	return name + strings.Join(pathParts, "_")
}
