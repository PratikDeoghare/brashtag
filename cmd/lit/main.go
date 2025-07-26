package main

import (
	"flag"
	"fmt"
	bt "github.com/pratikdeoghare/brashtag"
	"github.com/pratikdeoghare/brashtag/cmd/lit/extractor"
	"os"
	"strings"
)

func main() {
	filename := flag.String("f", "", "specify filename")
	outFile := flag.String("o", "", "output filename")
	flag.Parse()

	// tangle()
	weave(*filename, *outFile)
}

func weave(filename, outFile string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	text := strings.TrimSpace(string(data))

	tree, err := bt.Parse(fmt.Sprintf("#{%s}", text))
	if err != nil {
		panic(err)
	}

	a := extractor.NewExtractor(tree)
	fmt.Println(a.PrintProg(fmt.Sprintf("<<<%s>>>", outFile)))
}

func tangle(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	text := string(data)

	tree, err := bt.Parse(text)
	if err != nil {
		panic(err)
	}

	a := extractor.NewExtractor(tree)
	fmt.Println(a.PrintProg("<<<main.go>>>"))
}
