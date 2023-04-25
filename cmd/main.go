package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/sunshineplan/shortcut"
)

var (
	m shortcut.Map

	menu         []shortcut.Key
	maxKeyLength int
	keyNumber    int
)

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
	m.Range(func(k shortcut.Key, s shortcut.Shortcut) bool {
		if l := len(k); l > maxKeyLength {
			maxKeyLength = l
		}
		menu = append(menu, k)
		return true
	})
	for n := len(menu); n != 0; keyNumber++ {
		n /= 10
	}
	sort.Slice(menu, func(i, j int) bool { return menu[i] < menu[j] })
}

func main() {
	if len(menu) == 0 {
		log.Print("no shortcut loaded")
		return
	}

	flag.Parse()
	if *list {
		print()
		return
	}
	if *id > 0 {
		if err := run(*id); err != nil {
			log.Fatal(err)
		}
		return
	}

	switch flag.NArg() {
	case 0:
		print()
		fmt.Print("\nPlease choose: ")
		var choice string
		fmt.Scan(&choice)
		if strings.ToLower(choice) == "q" {
			return
		}
		n, err := strconv.Atoi(choice)
		if err != nil {
			log.Fatalln("bad choice:", choice)
		}
		if err := run(n); err != nil {
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

func print() {
	for i, key := range menu {
		sc, _ := m.Load(key)
		fmt.Printf(fmt.Sprintf("%%%dd", keyNumber)+". %s  %s  %s\n", i+1, key, strings.Repeat(" ", maxKeyLength-len(key)), sc)
	}
}

func run(choice int) error {
	if choice < 1 || choice > len(menu) {
		return fmt.Errorf("bad choice: %d", choice)
	}
	key := menu[choice-1]
	sc, _ := m.Load(key)
	fmt.Println("Run", key)
	if args := sc.Args(); args == 0 {
		return sc.Run()
	} else {
		fmt.Println(sc)
		var a []any
		for n := 1; n <= args; n++ {
			fmt.Printf("Please input argument %d: ", n)
			var arg string
			fmt.Scan(&arg)
			a = append(a, arg)
		}
		return sc.Run(a...)
	}
}
