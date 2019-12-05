package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jum/htempl"
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
	templ, err := htempl.New(fname)
	if err != nil {
		return err
	}
	ext := filepath.Ext(fname)
	dname := filepath.Join(*dest, fname[0:len(fname)-len(ext)]+"."+*suffix)
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
