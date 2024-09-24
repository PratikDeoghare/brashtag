package main

import (
	"bufio"
	"fmt"
	"github.com/goccy/go-graphviz"
	bt "github.com/pratikdeoghare/brashtag"
	"os"

	_ "github.com/goccy/go-graphviz"
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

	txt, err := mermaidIt(tree)
	if err != nil {
		panic(err)
	}
	fmt.Println(txt)
}

func mermaidIt(tree bt.Node) (string, error) {
	switch x := tree.(type) {
	case bt.Blob:
		return x.Text(), nil
	case bt.Code:
		return x.Text(), nil

	case bt.Bag:
		if x.Tag() == "mermaid" {
			for _, k := range x.Kids() {
				if kk, ok := k.(bt.Code); ok {
					loc, err := processDot(kk)
					if err != nil {
						return "", err
					}
					return fmt.Sprintf(`<img src="%s" />`, loc), nil
				}
			}
		}
		txt := ""
		for _, k := range x.Kids() {
			t, err := mermaidIt(k)
			if err != nil {
				return "", err
			}
			txt += t
		}
		return txt, nil
	}

	return "", nil
}

func processDot(c bt.Code) (string, error) {
	graph, err := graphviz.ParseBytes([]byte(c.Text()))
	if err != nil {
		return "", err
	}

	f, err := os.CreateTemp("", "mermaid-xxxxxx.png")
	if err != nil {
		return "", err
	}
	defer f.Close()

	g := graphviz.New()
	if err := g.RenderFilename(graph, graphviz.PNG, f.Name()); err != nil {
		return "", err
	}

	return f.Name(), nil
}
