package htempl

import (
	"bytes"
	"fmt"
	"io"
	"os"
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

func TestMapError(t *testing.T) {
	_, err := DefaultTemplFuncs["map"].(func(...interface{}) (map[string]interface{}, error))("foo")
	if err == nil {
		t.Error("expected error for odd number of args")
	}
}

func TestWithDefaultEdgeCases(t *testing.T) {
	m := make(map[string]interface{})
	f := DefaultTemplFuncs["withDefault"].(func(map[string]interface{}, string, interface{}) map[string]interface{})

	m1 := f(m, "", "val")
	if len(m1) != 0 {
		t.Error("expected empty map for empty key")
	}

	m2 := f(m, "key", nil)
	if len(m2) != 0 {
		t.Error("expected empty map for nil value")
	}
}

func TestArrayMerge(t *testing.T) {
	// Create temporary files for testing include merge
	includeFile := "test_include_merge.yaml"
	content := "list:\n  - item2\n"
	err := os.WriteFile(includeFile, []byte(content), 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(includeFile)
	}()

	mainContent := `---
include: test_include_merge.yaml
list:
  - item1
...
{{range .list}}{{.}}{{end}}`

	tmpl, err := NewString("TestArrayMerge", mainContent)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Template.Execute(&buf, tmpl.Vars)
	if err != nil {
		t.Fatal(err)
	}
	// Note: item2 from include is processed FIRST in NewWithTemplFuncsReader loop
	// then it's merged into templ.Vars which ALREADY has item1 from the main header.
	// Wait, let's re-read the code.
	/*
		if header != nil {
			includeFiles, templateFiles, templ.Vars, err = processYamlVars(header)
		}
		for len(includeFiles) > 0 {
			...
			nincs, ntempls, nvars, err := processYamlVars(f)
			...
			for k, v := range nvars {
				if oldval, ok := templ.Vars[k]; ok {
					oldarray, isArrayOld := oldval.([]interface{})
					newarray, isArrayNew := v.([]interface{})
					if isArrayOld && isArrayNew {
						templ.Vars[k] = append(oldarray, newarray...)
					} else {
						templ.Vars[k] = v
					}
				} else {
					templ.Vars[k] = v
				}
			}
		}
	*/
	// Header is processed first, so templ.Vars["list"] = ["item1"]
	// Then include file is processed, nvars["list"] = ["item2"]
	// oldarray = ["item1"], newarray = ["item2"]
	// templ.Vars["list"] = ["item1", "item2"]

	expected := "item1item2"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestTemplateKeywords(t *testing.T) {
	// template and templates keywords
	content := `---
template: testdata/simple.htempl
templates:
  - testdata/markdown.htempl
...
Body`
	tmpl, err := NewString("TestTemplateKeywords", content)
	if err != nil {
		t.Fatal(err)
	}
	if tmpl.Template.Lookup("simple.htempl") == nil {
		t.Error("simple.htempl not found in templates")
	}
	if tmpl.Template.Lookup("markdown.htempl") == nil {
		t.Error("markdown.htempl not found in templates")
	}
}

func TestSplitHeaderEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"ShortInput", "---", true},
		{"NoEndMarker", "---\nfoo: bar\n", true}, // it returns EOF if it doesn't find ... after \n
		{"NoHeader", "Body only", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := splitHeader(strings.NewReader(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("splitHeader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestVarOverwrite(t *testing.T) {
	// Test overwriting non-array variables in include
	includeFile := "test_include_overwrite.yaml"
	content := "var: newvalue\n"
	err := os.WriteFile(includeFile, []byte(content), 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove(includeFile)
	}()

	mainContent := `---
include: test_include_overwrite.yaml
var: oldvalue
...
{{.var}}`

	tmpl, err := NewString("TestVarOverwrite", mainContent)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Template.Execute(&buf, tmpl.Vars)
	if err != nil {
		t.Fatal(err)
	}
	// The include's vars overwrite the main ones if they are not both arrays.
	// Actually, looking at the code:
	/*
		for k, v := range nvars {
			if oldval, ok := templ.Vars[k]; ok {
				...
				if isArrayOld && isArrayNew {
					templ.Vars[k] = append(oldarray, newarray...)
				} else {
					templ.Vars[k] = v
				}
			} else {
				templ.Vars[k] = v
			}
		}
	*/
	// So newvalue should overwrite oldvalue.
	expected := "newvalue"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestNewWithTemplFuncs(t *testing.T) {
	tmpl, err := NewWithTemplFuncs("testdata/simple.htempl", DefaultTemplFuncs)
	if err != nil {
		t.Fatal(err)
	}
	if tmpl == nil {
		t.Fatal("expected template, got nil")
	}
}

func TestNestedIncludes(t *testing.T) {
	err := os.WriteFile("level2.yaml", []byte("val: final"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove("level2.yaml")
	}()
	err = os.WriteFile("level1.yaml", []byte("include: level2.yaml"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove("level1.yaml")
	}()

	content := `---
include: level1.yaml
...
{{.val}}`
	tmpl, err := NewString("TestNestedIncludes", content)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Template.Execute(&buf, tmpl.Vars)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "final" {
		t.Errorf("expected final, got %q", buf.String())
	}
}

func TestParseFilesError(t *testing.T) {
	content := `---
template: non_existent_template.htempl
...
Body`
	_, err := NewString("TestParseFilesError", content)
	if err == nil {
		t.Error("expected error for non-existent template file")
	}
}

func TestIncludeProcessYamlError(t *testing.T) {
	err := os.WriteFile("invalid.yaml", []byte("[invalid"), 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = os.Remove("invalid.yaml")
	}()

	content := `---
include: invalid.yaml
...
Body`
	_, err = NewString("TestIncludeProcessYamlError", content)
	if err == nil {
		t.Error("expected error for invalid YAML in include")
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("forced error")
}

func TestReaderError(t *testing.T) {
	_, err := NewReader("Error", &errorReader{})
	if err == nil {
		t.Error("expected error from errorReader")
	}
}

type errorAtByteReader struct {
	count int
}

func (e *errorAtByteReader) Read(p []byte) (n int, err error) {
	if e.count == 0 {
		content := "---\n"
		copy(p, content)
		e.count++
		return len(content), nil
	}
	return 0, fmt.Errorf("forced error at byte")
}

func TestSplitHeaderReadByteError(t *testing.T) {
	_, _, err := splitHeader(&errorAtByteReader{})
	if err == nil || !strings.Contains(err.Error(), "forced error at byte") {
		t.Errorf("expected forced error at byte, got %v", err)
	}
}

type errorInBodyReader struct {
	count int
}

func (e *errorInBodyReader) Read(p []byte) (n int, err error) {
	if e.count == 0 {
		content := "---\nfoo: bar\n...\nBody"
		copy(p, content)
		e.count++
		return len(content), nil
	}
	return 0, fmt.Errorf("forced error in body")
}

func TestReaderErrorInBody(t *testing.T) {
	_, err := NewReader("ErrorInBody", &errorInBodyReader{})
	if err == nil || !strings.Contains(err.Error(), "forced error in body") {
		t.Errorf("expected forced error in body, got %v", err)
	}
}

func TestIncludesSuccess(t *testing.T) {
	content := `---
includes:
  - testdata/testinclude.yaml
...
{{.var}}`
	tmpl, err := NewString("TestIncludesSuccess", content)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	err = tmpl.Template.Execute(&buf, tmpl.Vars)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "Value" {
		t.Errorf("expected Value, got %q", buf.String())
	}
}

func TestSplitHeaderEOFWithoutNewline(t *testing.T) {
	input := "---\nfoo: bar"
	header, _, err := splitHeader(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	_, err = io.Copy(&buf, header)
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "foo: bar" {
		t.Errorf("expected 'foo: bar', got %q", buf.String())
	}
}

func TestSplitHeaderNewlineNotEnd(t *testing.T) {
	input := "---\nfoo\nbar\n...\nBody"
	header, body, err := splitHeader(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	var hBuf, bBuf bytes.Buffer
	_, err = io.Copy(&hBuf, header)
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.Copy(&bBuf, body)
	if err != nil {
		t.Fatal(err)
	}
	if hBuf.String() != "foo\nbar\n" {
		t.Errorf("expected 'foo\nbar\n', got %q", hBuf.String())
	}
	if bBuf.String() != "Body" {
		t.Errorf("expected 'Body', got %q", bBuf.String())
	}
}

func TestNewWithTemplFuncsError(t *testing.T) {
	_, err := NewWithTemplFuncs("non_existent", DefaultTemplFuncs)
	if err == nil {
		t.Error("expected error")
	}
}
