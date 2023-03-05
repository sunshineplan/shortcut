# shortcut

The `shortcut` package provides a Go library for defining and executing shortcuts.

## Usage

First, create a new `Map` with `shortcut.NewMap()`. Then, use the `Store` method to define new shortcuts, and the `Load` method to retrieve them.

## Example
```go
package main

import (
	"fmt"

	"github.com/sunshineplan/shortcut"
)

func main() {
	m := shortcut.NewMap()
	m.Store("g", shortcut.Command("git", "%s"))
	m.Store("gi", shortcut.Command("git", "init"))
	m.Store("gs", shortcut.Command("git", "status"))

	if cmd, ok := m.Load("g"); ok {
		if err := cmd.Run("help"); err != nil {
			fmt.Println(err)
		}
	}

	if cmd, ok := m.Load("gs"); ok {
		if err := cmd.Run(); err != nil {
			fmt.Println(err)
		}
	}
}
```
This example creates a new shortcut.Map, adds three shortcuts to it, and executes two of them.

## License
This project is licensed under the MIT License.
