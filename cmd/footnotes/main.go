package main

import (
	"fmt"
	bt "github.com/pratikdeoghare/brashtag"
	"strings"
)

func main() {

	text := `There are only two families of _proper arbitrary_ markup languages: [TeX](https://tug.org/) and #footnote:sgml-link{[SGML](https://en.wikipedia.org/wiki/Standard_Generalized_Markup_Language)}. By _arbitrary_, I mean the grammar specifically, and how it can be used to _mark_ arbitrary plain text with information. And by _proper_, I mean the ability to have standalone nodes, user-definable nodes, nodes with attributes, and the wrapping of plain text. Everything else either lacks one of the these capabilities, or is a derivative or syntactic makeover of TeX or SGML.

		#footnote-content:sgml-link{ I would normally link to official thing as reference but it's behind the "wonderful" ISO paywall: [ISO 8879:1986](https://www.iso.org/standard/16387.html). }`

	tree, err := bt.Parse(fmt.Sprintf("#{%s}", text))
	if err != nil {
		panic(err)
	}

	footnotes := make(map[string]bt.Node)
	collectFootnotes(tree, footnotes)

	fmt.Println(toHtml(tree, footnotes))
}

func collectFootnotes(r bt.Node, footnotes map[string]bt.Node) {
	switch x := r.(type) {
	case bt.Bag:
		if strings.HasPrefix(x.Tag(), "footnote-content:") {
			footnotes[strings.TrimPrefix(x.Tag(), "footnote-content:")] = x
		}
		for _, k := range x.Kids() {
			collectFootnotes(k, footnotes)
		}
	default:
	}

}

func toHtml(r bt.Node, footnotes map[string]bt.Node) string {
	switch x := r.(type) {
	case bt.Bag:
		s := ""
		if strings.HasPrefix(x.Tag(), "footnote-content") {
			return ""
		}
		if strings.HasPrefix(x.Tag(), "footnote:") {
			f := footnotes[strings.TrimPrefix(x.Tag(), "footnote:")]

			s = getText(x) + fmt.Sprintf(`<div class="footnote">%s</div>`, getText(f.(bt.Bag)))

			return s
		}
		for _, k := range x.Kids() {
			s += toHtml(k, footnotes)
		}
		return s
	case bt.Code:
		return x.String()
	case bt.Blob:
		return x.String()

	default:
		return ""

	}
}

func getText(x bt.Bag) string {
	s := ""
	for _, k := range x.Kids() {
		s += k.String()
	}
	return s
}
