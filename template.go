package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type importValues struct {
	Name string
	Path string
}

const importTemplate = `
{{if ne .Name "timestamp" -}}
import * as {{.Name}} from './{{.Path}}'
{{end}}
`

func (iv *importValues) Compile() (string, error) {
	return compileAndExecute(importTemplate, iv)
}

type enumKeyVal struct {
	Name  string
	Value int32
}

type enumValues struct {
	Name   string
	Values []*enumKeyVal
}

const enumTemplate = `
{{$enumName := .Name}}
export enum {{$enumName}} {
  {{- range $i, $v := .Values}}
  {{- if $i}},{{end}}
  {{$v.Name}} = '{{$v.Name}}'
  {{- end}}
}
`

func (ev *enumValues) Compile() (string, error) {
	return compileAndExecute(enumTemplate, ev)
}

type messageValues struct {
	Name          string
	Interface     string
	JSONInterface string

	Fields      []*fieldValues
	NestedTypes []*messageValues
	NestedEnums []*enumValues
}

var messageTemplate = `
export interface {{.Interface}} {
  {{- if .Fields }}
  {{- range .Fields}}
  {{.Field }}: {{. | fieldType}}
  {{- end}}
  {{end}}
}

{{- if .NestedEnums}}
{{range .NestedEnums}}
{{. | compile}}
{{end}}
{{else}}

{{ end -}}
`

func (mv *messageValues) Compile() (string, error) {
	return compileAndExecute(messageTemplate, mv)
}

type fieldValues struct {
	Name       string
	Field      string
	Type       string
	IsEnum     bool
	IsRepeated bool
}

type serviceValues struct {
	Package   string
	Name      string
	Interface string
	Methods   []*serviceMethodValues
}

var serviceTemplate = `
export interface {{.Interface}} {
  {{- range .Methods}}
  {{.Name | methodName}}: (data: {{.InputType}}, headers?: object) => Promise<{{.OutputType}}>
  {{- end}}
}

export class {{.Name}}Impl implements {{.Interface}} {
  private hostname: string
  protected fetch: Fetch
  private path = '/twirp/{{.Package}}.{{.Name}}/'

  constructor(hostname: string, fetch: Fetch) {
    this.hostname = hostname
    this.fetch = fetch
  }

  protected url(name: string): string {
    return this.hostname + this.path + name
  }

  {{range .Methods}}
  public {{.Name | methodName}}(params: {{.InputType}}, headers: object = {}): Promise<{{.OutputType}}> {
    return this.fetch(
      this.url('{{.Name}}'),
      createTwirpRequest(params, headers)
    ).then((res) => {
      if (!res.ok) {
        return throwTwirpError(res)
      }
      return res.json()
    })
  }
  {{end}}
}
`

func (sv *serviceValues) Compile() (string, error) {
	return compileAndExecute(serviceTemplate, sv)
}

type serviceMethodValues struct {
	Name string

	Path       string
	InputType  string
	OutputType string
}

type protoFile struct {
	Messages []*messageValues
	Services []*serviceValues
	Enums    []*enumValues
	Imports  map[string]*importValues
}

var protoTemplate = `
/* tslint:disable */

// This file has been generated by https://github.com/h2oai/protoc-gen-twirp_ts.
// Do not edit.

{{- if .Imports}}
{{range .Imports -}}
{{. | compile}}
{{end}}
{{end -}}

{{- if .Services}}
import {
  createTwirpRequest,
  Fetch,
  throwTwirpError
} from './twirp'
{{end}}

{{- if .Enums}}
{{range .Enums -}}
{{. | compile}}
{{end -}}
{{end}}

{{- if .Messages}}

{{range .Messages -}}
{{. | compile}}

{{end -}}
{{end}}

{{- if .Services}}

// Services
{{range .Services}}
{{- . | compile}}
{{- end}}
{{- end}}
`

func (pf *protoFile) Compile() (string, error) {
	return compileAndExecute(protoTemplate, pf)
}

func compileAndExecute(tpl string, data interface{}) (string, error) {
	funcMap := template.FuncMap{
		"compile":       compile,
		"fieldType":     fieldType,
		"methodName":    methodName,
		"objectToField": objectToField,
	}

	t, err := template.New("").Funcs(funcMap).Parse(tpl)
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(nil)
	if err := t.Execute(buf, data); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}

func objectToField(fv fieldValues) string {
	t := fv.Type

	if t == "Date" {
		t = "string"
	}

	if fv.IsRepeated {
		switch t {
		case "string", "number", "boolean":
			return fmt.Sprintf("(m['%s']! || []).map((v) => { return %s(v)})", fv.Name, upperCaseFirst(t))
		}
		return fmt.Sprintf("(m['%s']! || []).map((v) => { return %s.fromJSON(v) })", fv.Name, t)
	}

	switch t {
	case "string", "number", "boolean":
		return fmt.Sprintf("m['%s']!", fv.Name)
	}

	if fv.IsEnum {
		return fmt.Sprintf("(<any>%s)[m['%s']!]!", fv.Type, fv.Name)
	}

	return fmt.Sprintf("%s.fromJSON(m['%s']!)", t, fv.Name)
}

func typeToInterface(typeName string) string {
	return typeName
}

func typeToJSONInterface(typeName string) string {
	return typeName + "JSON"
}

func methodName(method string) string {
	return strings.ToLower(method[0:1]) + method[1:]
}
