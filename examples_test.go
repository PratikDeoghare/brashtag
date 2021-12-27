package brashtag

import (
	"fmt"
	"strings"
)

func ExampleParse() {
	
	text := strings.Join([]string{
		"#div{",
		"	#b{A very short story}",
		"	#p{Lorem and impsum met for a coffee.}",
		"	` print 'Hello' `",
		"}",
	}, "\n")
	
	tree, _ := Parse(text)
	fmt.Println(toHTML(tree))
	
	// Input:
/*
	#div{
		#b{A very short story}
		#p{Lorem and impsum met for a coffee.}
		` print 'Hello' `
	}
*/

	// Output:
	//<div>
	//	<b>A very short story</b>
	//	<p>Lorem and impsum met for a coffee.</p>
	//	<code> print 'Hello' </code>
	//</div>
}

func toHTML(r Node) string {
	s := ""
	switch x := r.(type) {
	case Blob:
		return x.Text()
	case Code:
		return fmt.Sprintf("<code>%s</code>", x.Text())
	case Bag:
		s := ""
		for _, kid := range x.Kids() {
			s += toHTML(kid)
		}
		if x.Tag() != "" {
			return fmt.Sprintf("<%s>%s</%s>", x.Tag(), s, x.Tag())
		}
		return s
	}
	return s
}
