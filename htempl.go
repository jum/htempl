// Package htempl combines go html/templates with YAML configuration in one
// file.
package htempl

import (
	"bufio"
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"os"

	"github.com/gomarkdown/markdown"
	"gopkg.in/yaml.v2"
)

const (
	yamlStart = "---\n"
	yamlEnd   = "...\n"
)

// HTempl denotes a template combined by variables from a leading
// YAML block. The variable names "include", "includes", "template"
// and "templates" are special
type HTempl struct {
	Vars     map[string]interface{}
	Template *template.Template
}

// DefaultTemplFuncs are the default function mapping for New.
var DefaultTemplFuncs template.FuncMap = template.FuncMap{
	"withDefault": func(m map[interface{}]interface{}, key string, value interface{}) map[interface{}]interface{} {
		if len(key) == 0 || value == nil {
			return m
		}
		nm := make(map[interface{}]interface{})
		nm[key] = value
		for k, v := range m {
			nm[k] = v
		}
		return nm
	},
	"md2html": func(value string) template.HTML {
		return template.HTML(markdown.ToHTML([]byte(value), nil, nil))
	},
	"safeattr": func(value string) template.HTMLAttr {
		return template.HTMLAttr(value)
	},
	"safehtml": func(value string) template.HTML {
		return template.HTML(value)
	},
	"safejs": func(value string) template.JS {
		return template.JS(value)
	},
	"safecss": func(value string) template.CSS {
		return template.CSS(value)
	},
	"safeurl": func(value string) template.URL {
		return template.URL(value)
	},
}

// New parses a YAML/template file combination using the default FuncMap
func New(fname string) (*HTempl, error) {
	return NewWithTemplFuncs(fname, DefaultTemplFuncs)
}

// NewWithTemplFuncs parses a YAML/template file combination using the given FuncMap
func NewWithTemplFuncs(fname string, funcMap template.FuncMap) (*HTempl, error) {
	in, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer in.Close()
	header, body, err := splitHeader(in)
	if err != nil {
		return nil, err
	}
	var templ = HTempl{}
	templ.Vars = make(map[string]interface{})
	templ.Template = template.New(fname).Funcs(funcMap)
	var includeFiles []string
	var templateFiles []string
	if header != nil {
		includeFiles, templateFiles, templ.Vars, err = processYamlVars(header)
		if err != nil {
			return nil, err
		}
	}
	for len(includeFiles) > 0 {
		var fn string
		fn, includeFiles = includeFiles[0], includeFiles[1:]
		f, err := os.Open(fn)
		if err != nil {
			return nil, err
		}
		nincs, ntempls, nvars, err := processYamlVars(f)
		f.Close()
		if err != nil {
			return nil, err
		}
		includeFiles = append(includeFiles, nincs...)
		templateFiles = append(templateFiles, ntempls...)
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
	if len(templateFiles) > 0 {
		templ.Template, err = templ.Template.ParseFiles(templateFiles...)
		if err != nil {
			return nil, err
		}
	}
	// Why is there no function to parse from an io.Reader?
	main, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	templ.Template, err = templ.Template.Parse(string(main))
	if err != nil {
		return nil, err
	}
	return &templ, nil
}

func processYamlVars(f io.Reader) (includes []string, templates []string, otherVars map[string]interface{}, err error) {
	var includeFiles []string
	var templateFiles []string
	vars := make(map[string]interface{})
	dec := yaml.NewDecoder(f)
	err = dec.Decode(&vars)
	if err != nil {
		return nil, nil, nil, err
	}
	fn, ok := vars["include"]
	if ok {
		includeFiles = append(includeFiles, fn.(string))
		delete(vars, "include")
	}
	fnArray, ok := vars["includes"]
	if ok {
		for _, fn := range fnArray.([]interface{}) {
			includeFiles = append(includeFiles, fn.(string))
		}
		delete(vars, "includes")
	}
	fn, ok = vars["template"]
	if ok {
		templateFiles = append(templateFiles, fn.(string))
		delete(vars, "template")
	}
	fnArray, ok = vars["templates"]
	if ok {
		for _, fn := range fnArray.([]interface{}) {
			templateFiles = append(templateFiles, fn.(string))
		}
		delete(vars, "templates")
	}
	return includeFiles, templateFiles, vars, nil
}

func splitHeader(in io.Reader) (header, body io.Reader, err error) {
	bf := bufio.NewReader(in)
	b, err := bf.Peek(len(yamlStart))
	if err != nil {
		return nil, nil, err
	}
	if string(b) == yamlStart {
		_, err := io.ReadFull(bf, b)
		if err != nil {
			return nil, nil, err
		}
		var header bytes.Buffer
		for {
			b, err := bf.ReadByte()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, nil, err
			}
			if b == '\n' {
				endMarker, err := bf.Peek(len(yamlEnd))
				if err != nil {
					return nil, nil, err
				}
				if string(endMarker) == yamlEnd {
					_, err := io.ReadFull(bf, endMarker)
					if err != nil {
						return nil, nil, err
					}
					header.WriteByte(b)
					break
				}
			}
			header.WriteByte(b)
		}
		return &header, bf, nil
	}
	return nil, bf, nil
}
