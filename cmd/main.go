package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/sunshineplan/shortcut"
)

var m shortcut.Map

func init() {
	var path string
	if path = os.Getenv("SHORTCUT"); path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		path = filepath.Join(home, "shortcut.json")
		if _, err = os.Stat(path); err != nil {
			pwd, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path = filepath.Join(pwd, "shortcut.json")
			if _, err = os.Stat(path); err != nil {
				log.Fatal("no shortcut file found")
			}
		}
	}
	if err := m.FromFile(path); err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.Parse()

	switch flag.NArg() {
	case 0:
		flag.PrintDefaults()
	default:
		if cmd, ok := m.Load(shortcut.Key(flag.Arg(0))); ok {
			var a []any
			for _, arg := range flag.Args()[1:] {
				a = append(a, arg)
			}
			if err := cmd.Run(a...); err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatalf("Shortcut %s not found", flag.Arg(0))
		}
	}
}
