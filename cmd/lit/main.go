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

	tree, err := bt.Parse(fmt.Sprintf("#{%s}", text))
	if err != nil {
		panic(err)
	}

	a := &A{
		tree: tree,
		m:    make(map[string]string),
		deps: make(map[string][]dep),
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
		deps: make(map[string][]dep),
	}
	a.buildDeps(a.tree, "")
	fmt.Println(a.printProg("<<<main.go>>>"))
}

func (a A) printProg(root string) string {
	s := a.m[root]

	var lines []string
	for _, line := range strings.Split(s, "\n") {
		if spaceThunk.Match([]byte(line)) {
			idx := strings.Index(line, "<")
			prefix := line[:idx]
			for _, thunkLine := range strings.Split(a.printProg(strings.TrimSpace(line)), "\n") {
				lines = append(lines, prefix+thunkLine)
			}
		} else {
			lines = append(lines, line)
		}
	}

	return strings.Join(lines, "\n")
}

type A struct {
	tree bt.Node
	m    map[string]string
	deps map[string][]dep
}

type dep struct {
	id     string
	prefix string
}

func (a *A) buildDeps(tree bt.Node, path string) {
	path = fmt.Sprintf("<<<%s>>>", path)
	switch x := tree.(type) {
	case bt.Blob:

	case bt.Code:
		a.m[path] = stripMargin(x.Text())
		a.deps[path] = extractDeps(x.Text())

	case bt.Bag:
		for _, k := range x.Kids() {
			a.buildDeps(k, x.Tag())
		}
	}

}

func stripMargin(s string) string {
	lines := strings.Split(s, "\n")
	i := 0
	if len(lines) >= 2 {
		for i < len(lines[1]) && lines[1][i] == ' ' {
			i++
		}
	} else {
		return s
	}
	for j, line := range lines {
		if len(line) >= i && strings.TrimSpace(line[:i]) == "" {
			lines[j] = line[i:]
		} else {
			lines[j] = line
		}
	}
	return strings.Join(lines, "\n")
}

func stripMargin1(s string) string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		line2 := line
		line = strings.TrimSpace(line)
		if len(line) >= 2 && line[0] == '|' {
			line = line[2:]
		} else {
			line = line2
		}
		lines = append(lines, line)
	}
	x := strings.Join(lines, "\n")
	return x
}

var thunk = regexp.MustCompile(`<<<.*>>>`)
var spaceThunk = regexp.MustCompile(`^\s*<<<.*>>>`)

func extractDeps(text string) []dep {
	var deps []dep
	for _, line := range strings.Split(text, "\n") {
		if thunk.Match([]byte(line)) {
			idx := strings.Index(line, "<")
			deps = append(deps, dep{id: strings.TrimSpace(line[idx:]), prefix: line[:idx]})
		}
	}
	return deps
}
