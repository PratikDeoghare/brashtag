package main

import (
	"bufio"
	"fmt"
	"html"
	"os"
	"strings"

	bt "github.com/pratikdeoghare/brashtag"
)

func main() {
	text := ""
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text += scanner.Text() + "\n"
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	text = fmt.Sprintf("#{%s}", text)
	tree, err := bt.Parse(text)
	if err != nil {
		panic(err)
	}

	g := new(graph)
	toDot(tree, g)

	fmt.Println(g)

}

type graph struct {
	nodes []string
	edges []string
}

func (g graph) String() string {
	return fmt.Sprintf(digraph, strings.Join(g.nodes, "\n")+"\n"+strings.Join(g.edges, "\n"))
}

const digraph = `digraph {
	graph [ rankdir = "LR" ];
	%s
}`

const code = `n%d [
	shape = "record"
	style = "rounded,filled"
	fillcolor = "lightblue"
	label = "code | { %s | %s}"
];`

const blob = `n%d [
	shape = "record"
	style = "rounded,filled"
	fillcolor = "lightyellow"
	label = "blob | {%s}"
];`

const bag = `n%d [
	shape = "record"
	style = "rounded,filled"
	fillcolor = "pink"
	label = "bag| {%s}"
];`

const edge = `n%d -> n%d`

var c = 0

func id() int {
	c++
	return c
}

func toDot(tree bt.Node, g *graph) int {
	nid := id()
	switch t := tree.(type) {
	case bt.Bag:
		g.nodes = append(g.nodes, fmt.Sprintf(bag, nid, ws(t.Tag())))
		for _, kid := range t.Kids() {
			knid := toDot(kid, g)
			g.edges = append(g.edges, fmt.Sprintf(edge, nid, knid))
		}
	case bt.Blob:
		g.nodes = append(g.nodes, fmt.Sprintf(blob, nid, ws(t.Text())))
	case bt.Code:
		g.nodes = append(g.nodes, fmt.Sprintf(code, nid, t.Tag(), ws(t.Text())))
	}
	return nid
}

func ws(x string) string {
	x = strings.ReplaceAll(x, "{", `\{`)
	x = strings.ReplaceAll(x, "}", `\}`)
	x = strings.ReplaceAll(x, "   ", "  _")
	x = strings.ReplaceAll(x, "\n", "~\\l")
	return html.EscapeString(x)
}
