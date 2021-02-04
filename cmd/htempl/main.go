package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jum/htempl"
)

var (
	dest    = flag.String("dest", ".", "Destination directory")
	src     = flag.String("src", "", "source directory to walk directory")
	ssuffix = flag.String("srcsuffix", "htempl", "Default suffix for source files")
	dsuffix = flag.String("destsuffix", "html", "Default suffix for generated files")
	verbose = flag.Bool("verbose", false, "verbose debugging output")
)

func main() {
	flag.Parse()
	// walk the source directory if given
	if len(*src) > 0 {
		err := filepath.Walk(*src, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if *verbose {
				fmt.Printf("considering %s\n", path)
			}
			if filepath.Ext(path) == "."+*ssuffix {
				if *verbose {
					fmt.Printf("working on %s\n", path)
				}
				err := processFile(path)
				if err != nil {
					err = fmt.Errorf("%v: %w", path, err)
					fmt.Fprintf(os.Stderr, "htempl: %v\n", err)
					os.Exit(1)
				}
			}
			return nil
		})
		if err != nil {
			err = fmt.Errorf("%v: %w", *src, err)
			fmt.Fprintf(os.Stderr, "htempl: %v\n", err)
			os.Exit(1)
		}
	}
	// process the explicitely givenvi command line args
	for _, fname := range flag.Args() {
		if *verbose {
			fmt.Printf("working on %s\n", fname)
		}
		err := processFile(fname)
		if err != nil {
			err = fmt.Errorf("%v: %w", fname, err)
			fmt.Fprintf(os.Stderr, "htempl: %v\n", err)
			os.Exit(1)
		}
	}
}

func processFile(fname string) error {
	templ, err := htempl.New(fname)
	if err != nil {
		return err
	}
	ext := filepath.Ext(fname)
	if len(*src) > 0 {
		if strings.HasPrefix(fname, *src) {
			fname = fname[len(*src):]
		}
	}
	dname := filepath.Join(*dest, fname[0:len(fname)-len(ext)]+"."+*dsuffix)
	out, err := os.Create(dname)
	if err != nil {
		return err
	}
	err = templ.Template.Execute(out, templ.Vars)
	if err != nil {
		return err
	}
	return out.Close()
}
