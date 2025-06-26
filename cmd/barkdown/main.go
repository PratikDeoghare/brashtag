package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	bt "github.com/pratikdeoghare/brashtag"
)

const doc = `
	<html>
	<head>

			<style>
		    body
    {
        width:80%;
        margin-left:auto;
        margin-right:auto;
    }

		    .highlight {
				  animation: highlightFade 2s ease-out;
				}

				@keyframes highlightFade {
				  0% {
				    background-color: pink;
				  }
				  100% {
				    background-color: transparent;
				  }
				}
			</style> 


			<script>
			  document.addEventListener('DOMContentLoaded', () => {
			    document.querySelectorAll('a[href^="#"]').forEach(link => {
			      link.addEventListener('click', function (e) {
			        const id = this.getAttribute('href').substring(1);
			        const target = document.getElementById(id);
			        if (target) {
			          // Reset scroll behavior if needed
			          target.classList.remove('highlight');
			          void target.offsetWidth; // Force reflow
			          target.classList.add('highlight');
			        }
			      });
			    });
			  });
			</script>


	</head>
	<body>
`

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

	fmt.Println(doc + toHTML(tree, "") + `
	</body>
	</html>
`)

	fmt.Println()
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
		default:
			if x.Tag() != "" {
				html = fmt.Sprintf("<%s>%s</%s>", x.Tag(), html, x.Tag())
			}
		}

		return html
	}

	return ""
}

var re = regexp.MustCompile(`<<<(.*)>>>`)

func insertLinks(s string) string {
	return re.ReplaceAllString(s, `<a href="#$1">[$1]</a>`)
}
