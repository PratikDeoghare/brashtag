// Package brashtag provides parser for parsing
// files written in brashtag notation.
package brashtag

import (
	"fmt"
	"strings"
)

type Node interface {
	fmt.Stringer
	isNode()
}

type bag struct {
	tag  string
	kids []Node
}

// Bag has a tag and children.
type Bag struct {
	*bag
}

var _ Node = Bag{}

func (Bag) isNode() {}

func NewBag(tag string, kids ...Node) Bag {
	return Bag{
		&bag{
			tag:  tag,
			kids: kids,
		},
	}
}

func (b Bag) AddKids(kids ...Node) {
	b.kids = append(b.kids, kids...)
}

func (b Bag) RemoveKid(i int) {
	if i < len(b.kids) {
		b.kids = append(b.kids[:i], b.kids[i+1:]...)
	}
}

func (b Bag) SetTag(tag string) {
	b.tag = tag
}

func (b Bag) Kids() []Node {
	return b.kids
}

func (b Bag) Tag() string {
	return b.tag
}

func (b Bag) String() string {
	var text []string

	for _, kid := range b.Kids() {
		text = append(text, kid.String())
	}

	return fmt.Sprintf("#%s{%s}", b.tag, strings.Join(text, ""))
}

type code struct {
	tag  string
	text string
}

type Code struct {
	*code
}

var _ Node = Code{}

func (Code) isNode() {}

func NewCode(tag, text string) Code {
	return Code{
		&code{
			tag:  tag,
			text: text,
		},
	}
}

func (c Code) SetTag(tag string) {
	c.tag = tag
}

func (c Code) SetText(text string) {
	c.text = text
}

func (c Code) Text() string {
	return c.text
}

func (c Code) Tag() string {
	return c.tag
}

func (c Code) String() string {
	return fmt.Sprintf("%s%s%s", c.tag, c.text, c.tag)
}

type blob struct {
	text string
}

type Blob struct {
	*blob
}

var _ Node = Blob{}

func (Blob) isNode() {}

func NewBlob(text string) Blob {
	return Blob{
		&blob{text: text},
	}
}

func (b Blob) SetText(text string) {
	b.text = text
}

func (b Blob) Text() string {
	return b.text
}

func (b Blob) String() string {
	return fmt.Sprint(b.text)
}

// Parse parses the text and returns its brashtag tree.
// It always puts everything in a bag at the root.
func Parse(text string) (Node, error) {
	text = fmt.Sprintf("#{%s}", text)
	root, rem, err := parseBag([]byte(text))
	if len(rem) != 0 {
		fmt.Println(string(rem))
		return nil, fmt.Errorf("parsing incomplete: %s", string(rem))
	}
	if err != nil {
		return nil, err
	}
	return root, nil
}

func parseBag(text []byte) (Node, []byte, error) {
	root := NewBag("")

	c := 0
	for c < len(text) {
		if text[c] != '{' {
			c++
		} else {
			break
		}
	}

	root.SetTag(string(text[1:c]))
	text = text[c+1:]

	var k Node
	var err error

	for len(text) > 0 {
		switch {
		case text[0] == '#':
			k, text, err = parseBag(text)
			if err != nil {
				return nil, text, err
			}
			root.AddKids(k)

		case text[0] == '`':
			k, text, err = parseCode(text)
			if err != nil {
				return nil, text, err
			}
			root.AddKids(k)

		case text[0] == '}':
			return root, text[1:], nil

		default:
			k, text, err = parseBlob(text)
			if err != nil {
				return nil, text, err
			}
			root.AddKids(k)
		}
	}

	return nil, text, fmt.Errorf("bag not closed")
}

func parseCode(text []byte) (Node, []byte, error) {
	root := NewCode("", "")
	c := 0
	for c < len(text) {
		if text[c] == '`' {
			c++
		} else {
			break
		}
	}
	root.SetTag(string(text[0:c]))
	j := c
	for j < len(text)-c {
		if string(text[j:j+c]) != root.tag {
			j++
		} else {
			root.SetText(string(text[c:j]))
			return root, text[j+c:], nil
		}
	}

	return nil, text, fmt.Errorf("code not closed")
}

func parseBlob(text []byte) (Node, []byte, error) {
	root := NewBlob("")
	j := 0

loop:
	for j < len(text) {
		switch text[j] {
		case '`', '#', '}':
			break loop
		default:
			j++
		}
	}
	root.SetText(string(text[:j]))

	return root, text[j:], nil
}
