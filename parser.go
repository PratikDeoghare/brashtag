package brashtag

import (
	"fmt"
	"strings"
)

type Tree interface {
	isTree()
	String() string
	Kids() []Tree
}

type Bag struct {
	kids []Tree
	tag  string
}

func (Bag) isTree() {}

func (b Bag) Kids() []Tree {
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

type Code struct {
	tag  string
	text string
}

func (Code) isTree() {}

func (Code) Kids() []Tree {
	return nil
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

type Blob struct {
	text string
}

func (Blob) isTree() {}

func (Blob) Kids() []Tree {
	return nil
}

func (b Blob) Text() string {
	return b.text
}

func (b Blob) String() string {
	return b.text
}

func Parse(text string) (Tree, error) {
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

func MustParse(text string) Tree {
	t, err := Parse(text)
	if err != nil {
		panic(err)
	}
	return t
}

func parseBag(text []byte) (Tree, []byte, error) {
	root := Bag{}

	c := 0
	for c < len(text) {
		if text[c] != '{' {
			c++
		} else {
			break
		}
	}

	root.tag = string(text[1:c])
	text = text[c+1:]

	var k Tree
	var err error

	for len(text) > 0 {
		switch {
		case text[0] == '#':
			k, text, err = parseBag(text)
			if err != nil {
				return nil, text, err
			}
			root.kids = append(root.kids, k)

		case text[0] == '$':
			k, text, err = parseCode(text)
			if err != nil {
				return nil, text, err
			}
			root.kids = append(root.kids, k)

		case text[0] == '}':
			return root, text[1:], nil

		default:
			k, text, err = parseBlob(text)
			if err != nil {
				return nil, text, err
			}
			root.kids = append(root.kids, k)
		}
	}

	return nil, text, fmt.Errorf("bag not closed")
}

func parseCode(text []byte) (Tree, []byte, error) {
	root := Code{}
	c := 0
	for c < len(text) {
		if text[c] == '$' {
			c++
		} else {
			break
		}
	}
	root.tag = string(text[0:c])
	j := c
	for j < len(text)-c {
		if string(text[j:j+c]) != root.tag {
			j++
		} else {
			root.text = string(text[c:j])
			return root, text[j+c:], nil
		}
	}

	return nil, text, fmt.Errorf("code not closed")
}

func parseBlob(text []byte) (Tree, []byte, error) {
	root := Blob{}
	j := 0

loop:
	for j < len(text) {
		switch text[j] {
		case '$', '#', '}':
			break loop
		default:
			j++
		}
	}
	root.text = string(text[:j])

	return root, text[j:], nil
}
