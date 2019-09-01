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

    "gopkg.in/yaml.v2"
    "github.com/russross/blackfriday"
)

const (
	yamlStart = "---\n"
	yamlEnd   = "...\n"
)

var (
	dest = flag.String("dest", ".", "Destination directory")
)

func main() {
	flag.Parse()
	for _, fname := range flag.Args() {
        //fmt.Printf("working on %s\n", fname)
        in, err := os.Open(fname)
        if err != nil {
            panic(err)
        }
		header, body, err := splitHeader(in)
		if err != nil {
			panic(err)
		}
		vars := make(map[string]interface{})
		templ := template.New("htempl").Funcs(template.FuncMap{
            "md2html": func(value string) template.HTML {
                return template.HTML(blackfriday.MarkdownCommon([]byte(value)))
            },
            "htmlattr": func(value string) template.HTMLAttr {
                return template.HTMLAttr(value)
            },
        })
		if header != nil {
			dec := yaml.NewDecoder(header)
			err = dec.Decode(&vars)
			if err != nil {
				panic(err)
			}
        }
        var includeFiles []string
		fn, ok := vars["include"]
		if ok {
			includeFiles = append(includeFiles, fn.(string))
		}
		fnArray, ok := vars["includes"]
		if ok {
			for _, fn := range fnArray.([]interface{}) {
				includeFiles = append(includeFiles, fn.(string))
			}
        }
        if len(includeFiles) > 0 {
            for _, fn := range includeFiles {
                f, err := os.Open(fn)
                if err != nil {
                    panic(err)
                }
                dec := yaml.NewDecoder(f)
                err = dec.Decode(&vars)
                if err != nil {
                    panic(err)
                }
                f.Close()
            }
        }
		//fmt.Printf("Header vars: %#v\n", vars)
        var templateFiles []string
		fn, ok = vars["template"]
		if ok {
			templateFiles = append(templateFiles, fn.(string))
		}
		fnArray, ok = vars["templates"]
		if ok {
			for _, fn := range fnArray.([]interface{}) {
				templateFiles = append(templateFiles, fn.(string))
			}
        }
        if len(templateFiles) > 0 {
		    templ, err = templ.ParseFiles(templateFiles...)
		    if err != nil {
			    panic(err)
            }
        }
		// Why is there no function to parse from an io.Reader?
		main, err := ioutil.ReadAll(body)
		if err != nil {
			panic(err)
		}
		templ, err = templ.Parse(string(main))
		if err != nil {
			panic(err)
        }
        in.Close()
		ext := filepath.Ext(fname)
		dname := filepath.Join(*dest, fname[0:len(fname)-len(ext)]+".html")
        out, err := os.Create(dname)
        if err != nil {
            panic(err)
        }
        err = templ.Execute(out, vars)
        if err != nil {
            panic(err)
        }
        out.Close()
	}
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
