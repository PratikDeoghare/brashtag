package main

import (
	"bufio"
	"fmt"
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

	tree, err := bt.Parse(text)
	if err != nil {
		panic(err)
	}

	fmt.Println(toHTML(tree))
}

func toHTML(tree bt.Node) string {
	switch x := tree.(type) {
	case bt.Blob:
		return x.Text()

	case bt.Code:
		return "<pre>" + x.Text() + "</pre>"

	case bt.Bag:
		html := ""
		for _, kid := range x.Kids() {
			html += toHTML(kid)
		}

		switch x.Tag() {
		case "list":
			items := strings.Split(html, ",")
			temp := "<ul>"
			for _, item := range items {

				item = strings.TrimSpace(item)
				if item == "" {
					continue
				}

				temp += "\n\t" + "<li>" + item + "</li>" + "\n"
			}
			html = temp + "</ul>"

		case "post":
			html = "<div>" + html + "</div>"

		case "title":
			html = "<h1>" + html + "</h1>"

		case "subtitle":
			html = "<h3>" + html + "</h3>"
		}

		return html
	}

	return ""
}
