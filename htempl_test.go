package htempl

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestNewString(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    string
		wantErr bool
	}{
		{"SimpleVar", `---
var: "value"
...
{{.var}}`, "value", false},
		{"Markdown", `---
var: "**value**"
...
{{md2html .var}}`, "<p><strong>value</strong></p>\n", false},
		{"NestedMap", `---
data:
  user:
    name: "John Doe"
    age: 30
...
{{.data.user.name}} - {{.data.user.age}}`, "John Doe - 30", false},
		{"ListInMap", `---
items:
  - apple
  - banana
...
{{range .items}}{{.}}{{end}}`, "applebanana", false},
		{"IntegerAndBoolean", `---
count: 10
enabled: true
...
{{.count}} {{.enabled}}`, "10 true", false},
		{"IncludeDirective", `---
include: testdata/b.htempl
...
{{.foo}} {{.bar}}`, "foo bar", true},
		{"IncludesDirective", `---
includes:
  - testdata/b.htempl
...
{{.foo}} {{.bar}}`, "foo bar", true},
		{"SafeAttr", `{{safeattr "foo"}}`, `foo`, false},
		{"SafeHTML", `{{safehtml "<b>bold</b>"}}`, `<b>bold</b>`, false},
		{"SafeJS", `{{safejs "alert('hi')"}}`, `alert(&#39;hi&#39;)`, false},
		{"SafeCSS", `{{safecss "color: red;"}}`, `color: red;`, false},
		{"SafeURL", `{{safeurl "http://example.com"}}`, `http://example.com`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewString(tt.name, tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			var buf bytes.Buffer
			err = got.Template.Execute(&buf, got.Vars)
			if err != nil {
				t.Errorf("Executing %v: %v", tt.name, err)
				return
			}
			if buf.String() != tt.want {
				t.Errorf("NewString() = %v, want %v", buf.String(), tt.want)
				return
			}
		})
	}
}

func TestNewReader(t *testing.T) {
	type args struct {
		tmplName string
		in       io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"SimpleVar", args{"SimpleVar", strings.NewReader(`---
var: "value"
...
{{.var}}`)}, "value", false},
		{"Markdown", args{"Markdown", strings.NewReader(`---
var: "**value**"
...
{{md2html .var}}`)}, "<p><strong>value</strong></p>\n", false},
		{"NestedMap", args{"NestedMap", strings.NewReader(`---
data:
  user:
    name: "John Doe"
    age: 30
...
{{.data.user.name}} - {{.data.user.age}}`)}, "John Doe - 30", false},
		{"ListInMap", args{"ListInMap", strings.NewReader(`---
items:
  - apple
  - banana
...
{{range .items}}{{.}}{{end}}`)}, "applebanana", false},
		{"IntegerAndBoolean", args{"IntegerAndBoolean", strings.NewReader(`---
count: 10
enabled: true
...
{{.count}} {{.enabled}}`)}, "10 true", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewReader(tt.args.tmplName, tt.args.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var buf bytes.Buffer
			err = got.Template.Execute(&buf, got.Vars)
			if err != nil {
				t.Errorf("Executing %v: %v", tt.name, err)
				return
			}
			if buf.String() != tt.want {
				t.Errorf("NewReader() = %v, want %v", buf.String(), tt.want)
				return
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		fname string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"SimpleVar", args{"testdata/simple.htempl"}, "value", false},
		{"Markdown", args{"testdata/markdown.htempl"}, "<p><strong>value</strong></p>\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.fname)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			var buf bytes.Buffer
			err = got.Template.Execute(&buf, got.Vars)
			if err != nil {
				t.Errorf("Executing %v: %v", tt.name, err)
				return
			}
			if buf.String() != tt.want {
				t.Errorf("New() = %v, want %v", buf.String(), tt.want)
				return
			}
		})
	}
}

func TestWithDefault(t *testing.T) {
	m := make(map[string]interface{})
	m["foo"] = "bar"
	m = DefaultTemplFuncs["withDefault"].(func(map[string]interface{}, string, interface{}) map[string]interface{})(m, "baz", "bam")
	if m["foo"] != "bar" {
		t.Errorf("foo should be bar but is %v", m["foo"])
		return
	}
	if m["baz"] != "bam" {
		t.Errorf("baz should be bam but is %v", m["baz"])
		return
	}
}

func TestMap(t *testing.T) {
	m, err := DefaultTemplFuncs["map"].(func(...interface{}) (map[string]interface{}, error))("foo", "bar", "baz", "bam")
	if err != nil {
		t.Error(err)
	}
	if m["foo"] != "bar" {
		t.Errorf("foo should be bar but is %v", m["foo"])
	}
	if m["baz"] != "bam" {
		t.Errorf("baz should be bam but is %v", m["baz"])
	}
}

func TestWithDefaultAndMap(t *testing.T) {
	tmpl, err := NewString("map", `---
map:
  foo: bar
...
{{- $M := map "baz" "bam" -}}
{{- $M := withDefault $M "foo" .map.foo -}}
{{ $M.foo }} {{ $M.baz}}`)
	if err != nil {
		t.Error(err)
	}
	var buf bytes.Buffer
	err = tmpl.Template.Execute(&buf, tmpl.Vars)
	if err != nil {
		t.Error(err)
	}
	if buf.String() != "bar bam" {
		t.Errorf("expected 'bar bam' but got '%s'", buf.String())
	}
}

func TestInclude(t *testing.T) {
	result := `<h1 id="anchor">Test</h1>

<p>Text</p>
Value

`
	tmpl, err := New("testdata/testinclude.htempl")
	if err != nil {
		t.Error(err)
		return
	}
	var buf bytes.Buffer
	err = tmpl.Template.Execute(&buf, tmpl.Vars)
	if err != nil {
		t.Error(err)
		return
	}
	if buf.String() != result {
		t.Errorf("expected '%s' but got '%s'", result, buf.String())
		return
	}
}

func TestIncludeError(t *testing.T) {
	aHTemplContent := `---
include: non_existent.htempl
---
{{.foo}}`
	_, err := NewString("TestIncludeError", aHTemplContent)
	if err == nil {
		t.Errorf("expected an error for non-existent include, but got none")
	}
}
