package main

import "github.com/DrewRepaskyOlive/loop-sheets/loop"

func main() {
	if err := loop.Serve(); err != nil {
		panic(err)
	}
}
