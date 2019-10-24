package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/russross/blackfriday"
	"gopkg.in/yaml.v2"
)

const (
	yamlStart = "---\n"
	yamlEnd   = "...\n"
)

var (
	dest    = flag.String("dest", ".", "Destination directory")
	suffix  = flag.String("suffix", "html", "Default suffix for generated files")
	verbose = flag.Bool("verbose", false, "verbose debugging output")
)

func main() {
	flag.Parse()
	for _, fname := range flag.Args() {
		if *verbose {
			fmt.Printf("working on %s\n", fname)
		}
		err := processFile(fname)
		if err != nil {
			err = fmt.Errorf("%v: %w", fname, err)
			panic(err)
		}
	}
}

func processFile(fname string) error {
	in, err := os.Open(fname)
	if err != nil {
		return err
	}
	defer in.Close()
	header, body, err := splitHeader(in)
	if err != nil {
		return err
	}
	templ := template.New("htempl").Funcs(template.FuncMap{
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
			return template.HTML(blackfriday.MarkdownCommon([]byte(value)))
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
	})
	vars := make(map[string]interface{})
	var includeFiles []string
	var templateFiles []string
	if header != nil {
		includeFiles, templateFiles, vars, err = processYamlVars(header)
		if err != nil {
			return err
		}
	}
	if *verbose {
		fmt.Printf("initial vars %v\n", vars)
	}
	for len(includeFiles) > 0 {
		var fn string
		fn, includeFiles = includeFiles[0], includeFiles[1:]
		f, err := os.Open(fn)
		if err != nil {
			return err
		}
		nincs, ntempls, nvars, err := processYamlVars(f)
		f.Close()
		includeFiles = append(includeFiles, nincs...)
		templateFiles = append(ntempls, templateFiles...)
		for k, v := range nvars {
			vars[k] = v
		}
		if *verbose {
			fmt.Printf("vars after including %s: %v\n", fn, vars)
		}
	}
	if len(templateFiles) > 0 {
		if *verbose {
			fmt.Printf("parsing templates %v\n", templateFiles)
		}
		templ, err = templ.ParseFiles(templateFiles...)
		if err != nil {
			return err
		}
	}
	// Why is there no function to parse from an io.Reader?
	main, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	templ, err = templ.Parse(string(main))
	if err != nil {
		return err
	}
	ext := filepath.Ext(fname)
	dname := filepath.Join(*dest, fname[0:len(fname)-len(ext)]+"."+*suffix)
	out, err := os.Create(dname)
	if err != nil {
		return err
	}
	err = templ.Execute(out, vars)
	if err != nil {
		return err
	}
	return out.Close()
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
				fmt.Printf("header EOF?\n")
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
