package brashtag

import (
	"fmt"
	"strings"
	"testing"
)

func walk(t Tree, tab string) {
	switch x := t.(type) {
	case Bag:
		fmt.Printf("%sBag(%s)\n", tab, x.Tag())
		for _, kid := range x.Kids() {
			walk(kid, tab+".")
		}
	case Code:
		fmt.Printf("%sCode(%s)\n", tab, x)
	case Blob:
		fmt.Printf("%sBlob(%s)\n", tab, x)
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		text string
	}{
		{
			"text",
			`This is good thing.`,
		},
		{
			"bag",
			`#foo{is bar}`,
		},
		{
			"code",
			`$print("Hello, World")$`,
		},
		{
			"card",
			`#q{
What is your name? 
#a{
	The name is Bond. James Bond.
}
}`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tree := MustParse(test.text)
			t.Logf(tree.String())
			walk(tree, "")
		})
	}

}

func TestDump(t *testing.T) {

	lipsum := `
"Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi
ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit
in voluptate velit esse cillum dolore eu fugiat nulla pariatur. 

Excepteur sint occaecat cupidatat non proident, sunt in culpa
qui officia deserunt mollit anim id est laborum."
`

	lipsumBag := `
#some tag here{"Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi
ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit
in voluptate velit esse cillum dolore eu fugiat nulla pariatur. 

Excepteur sint occaecat cupidatat non proident, sunt in culpa
qui officia deserunt mollit anim id est laborum."}
`

	lipsumCode := `
$#some tag here{"Lorem ipsum dolor sit amet, consectetur adipiscing elit,
sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.
Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi
ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit
in voluptate velit esse cillum dolore eu fugiat nulla pariatur. 

Excepteur sint occaecat cupidatat non proident, sunt in culpa
qui officia deserunt mollit anim id est laborum."}$
`

	tests := []struct {
		name string
		text string
		tree Tree
	}{
		{
			"plain text",
			lipsum,
			Bag{
				kids: []Tree{
					Blob{
						text: lipsum,
					},
				},
			},
		},
		{
			"bag",
			lipsumBag,
			Bag{
				kids: []Tree{
					Blob{
						text: "\n",
					},
					Bag{
						tag: "some tag here",
						kids: []Tree{
							Blob{
								text: strings.TrimSpace(lipsum),
							},
						},
					},
					Blob{
						text: "\n",
					},
				},
			},
		},
		{
			"code",
			lipsumCode,
			Bag{
				kids: []Tree{
					Blob{
						text: "\n",
					},
					Code{
						tag:  "$",
						text: strings.TrimSpace(lipsumBag),
					},
					Blob{
						text: "\n",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dump := test.tree.String()
			t.Log(len(dump), "bytes.")
			expected := fmt.Sprintf("#{%s}", test.text)
			if dump != expected {
				t.Errorf("Expected:\n %s \nGot:\n %s", expected, dump)
			}
		})
	}

}
