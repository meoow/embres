package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var inplace bool

func init() {
	flag.BoolVar(&inplace, "i", false, "Change file inplace")
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if inplace {
		for _, f := range flag.Args() {
			writer, err := ioutil.TempFile(filepath.Dir(f), "temp")
			if err != nil {
				log.Print(err)
				return
			}
			defer func() {
				if _, err := os.Stat(writer.Name()); err != nil {
					writer.Close()
					os.Remove(writer.Name())
				}
			}()
			embed_file(f, writer)
			writer.Close()
			err = os.Rename(writer.Name(), f)
			if err != nil {
				log.Print(err)
				return
			}
		}
	} else {
		for _, f := range flag.Args() {
			embed_file(f, os.Stdout)
		}
	}
}
