package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
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

	tree, err := bt.Parse(fmt.Sprintf("#{%s}", strings.TrimSpace(text)))
	if err != nil {
		panic(err)
	}

	fmt.Println(toHTML(tree, ""))
}

func toHTML(tree bt.Node, parent string) string {
	switch x := tree.(type) {
	case bt.Blob:
		return x.Text()

	case bt.Code:
		if x.Tag() == "`" {
			return fmt.Sprintf("<code>%s</code>", x.Text())
		}
		s := "<pre>" + insertLinks(x.Text()) + "</pre>"
		return fmt.Sprintf(`<div id="%s" class="code3"><code>[%s]</code>%s</div>`, parent, parent, s)

	case bt.Bag:
		html := ""
		for _, kid := range x.Kids() {
			html += toHTML(kid, x.Tag())
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
		case "p":
			html = "<p>" + html + "</p>"

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

var re = regexp.MustCompile(`<<<(.*)>>>`)

func insertLinks(s string) string {
	return re.ReplaceAllString(s, `<a href="#$1">[$1]</a>`)
}
