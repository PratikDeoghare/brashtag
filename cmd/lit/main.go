package main

import (
	"flag"
	"fmt"
	bt "github.com/pratikdeoghare/brashtag"
	"os"
	"regexp"
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

	tree, err := bt.Parse(text)
	if err != nil {
		panic(err)
	}

	a := &A{
		tree: tree,
		m:    make(map[string]string),
		deps: make(map[string][]string),
	}
	a.buildDeps(a.tree, "")
	fmt.Println(a.printProg(fmt.Sprintf("<<<%s>>>", outFile)))
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

	a := &A{
		tree: tree,
		m:    make(map[string]string),
		deps: make(map[string][]string),
	}
	a.buildDeps(a.tree, "")
	fmt.Println(a.printProg("<<<main.go>>>"))
}

func (a A) printProg(root string) string {
	s := a.m[root]

	for _, dep := range a.deps[root] {
		s = strings.ReplaceAll(s, dep, a.printProg(dep))
	}
	return s
}

type A struct {
	tree bt.Node
	m    map[string]string
	deps map[string][]string
}

func (a *A) buildDeps(tree bt.Node, path string) {
	path = fmt.Sprintf("<<<%s>>>", path)
	switch x := tree.(type) {
	case bt.Blob:

	case bt.Code:
		a.m[path] = x.Text()
		a.deps[path] = extractDeps(x.Text())

	case bt.Bag:
		for _, k := range x.Kids() {
			a.buildDeps(k, x.Tag())
		}
	}

}

var thunk = regexp.MustCompile(`<<<.*>>>`)

func extractDeps(text string) []string {
	var deps []string
	for _, line := range strings.Split(text, "\n") {
		if thunk.Match([]byte(line)) {
			deps = append(deps, strings.TrimSpace(line))
		}
	}
	return deps
}
