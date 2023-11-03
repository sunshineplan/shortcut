package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sunshineplan/shortcut"
	"github.com/sunshineplan/utils/choice"
)

var m shortcut.Map

var (
	list = flag.Bool("list", false, "list shortcuts")
	id   = flag.Int("run", 0, "run shortcut")
)

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
	if m.Count() == 0 {
		log.Print("no shortcut loaded")
		return
	}

	flag.Parse()
	if *list {
		fmt.Print(m.Menu(false))
		return
	}
	if *id > 0 {
		key, sc, err := m.Index(*id)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Run", key)
		if err := run(sc); err != nil {
			log.Fatal(err)
		}
		return
	}

	switch flag.NArg() {
	case 0:
		fmt.Print(m.Menu(true))
		ok, key, sc, err := m.Choose()
		for i := 0; errors.Is(err, choice.ErrBadChoice); i++ {
			fmt.Println(err)
			if i != 0 && i%5 == 0 {
				fmt.Print(m.Menu(true))
			}
			ok, key, sc, err = m.Choose()
		}
		if !ok {
			return
		}
		fmt.Println("Run", key)
		if err := run(sc); err != nil {
			log.Fatal(err)
		}
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

func run(sc shortcut.Shortcut) error {
	if args := sc.Args(); args == 0 {
		return sc.Run()
	} else {
		fmt.Println(sc)
		var a []any
		for n := 1; n <= args; n++ {
			fmt.Printf("Please input argument %d: ", n)
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			arg := scanner.Text()
			a = append(a, arg)
		}
		return sc.Run(a...)
	}
}
