package main

import (
	"fmt"

	gd "github.com/amterp/go-delta"
)

const old = `{
  "name": "Alice",
  "age": 30,
  "email": "alice@example.com",
  "hobbies": ["reading", "hiking"],
  "address": {
    "street": "123 Main St",
    "city": "Springfield",
    "state": "IL"
  }
}`

const new_ = `{
  "name": "Bob",
  "age": 31,
  "email": "bob@example.com",
  "hobbies": ["reading", "cycling", "cooking"],
  "address": {
    "street": "456 Oak Ave",
    "city": "Springfield",
    "zip": "62704"
  }
}`

func main() {
	fmt.Println("=== Inline Mode ===")
	fmt.Println()
	fmt.Println(gd.DiffWith(old, new_, gd.WithColor(true)))

	fmt.Println("=== Side-by-Side Mode ===")
	fmt.Println()
	fmt.Println(gd.DiffWith(old, new_, gd.WithColor(true), gd.WithLayout(gd.LayoutSideBySide), gd.WithWidth(100)))

	fmt.Println("=== Side-by-Side Mode (full width) ===")
	fmt.Println()
	fmt.Println(gd.DiffWith(old, new_, gd.WithColor(true), gd.WithLayout(gd.LayoutSideBySide)))

	fmt.Println("=== Prefer Side-by-Side Mode ===")
	fmt.Println()
	fmt.Println(gd.DiffWith(old, new_, gd.WithColor(true), gd.WithLayout(gd.LayoutPreferSideBySide)))
}
