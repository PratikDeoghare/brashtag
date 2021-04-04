package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	bt "github.com/pratikdeoghare/brashtag"
)

var cardTmpl = template.Must(template.ParseFiles("./card-template.html"))

var cards []Card

var counter = 0

type Card struct {
	Front string
	Back  string
}

func handler(w http.ResponseWriter, _ *http.Request) {

	type htmlCard struct {
		Front template.HTML
		Back  template.HTML
	}

	n := len(cards)

	c := htmlCard{
		Front: template.HTML(cards[counter%n].Front),
		Back:  template.HTML(cards[counter%n].Back),
	}

	err := cardTmpl.Execute(w, c)
	if err != nil {
		log.Fatal(err)
	}

	counter++
	log.Printf("Studied %d cards.", counter)
}

func studylogHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Confidence score is: %s", r.PostFormValue("score"))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	text, err := ioutil.ReadFile("./cards.bt")
	if err != nil {
		panic(err)
	}

	tree, err := bt.Parse(string(text))
	if err != nil {
		panic(err)
	}

	for _, kid := range tree.(bt.Bag).Kids() {
		switch x := kid.(type) {
		case bt.Bag:
			cards = append(cards, makeCard(x))
		}
	}

	http.HandleFunc("/", handler)
	http.HandleFunc("/studylog", studylogHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func makeCard(t bt.Bag) Card {
	var front []string
	for _, kid := range t.Kids() {
		switch x := kid.(type) {
		case bt.Blob:
			front = append(front, x.Text())
		case bt.Code:
			front = append(front, code(x))
		case bt.Bag:
			t = x
		}
	}

	var back []string
	for _, kid := range t.Kids() {
		switch x := kid.(type) {
		case bt.Code:
			back = append(back, code(x))
		case bt.Blob:
			back = append(back, x.Text())
		}
	}

	return Card{
		Front: strings.Join(front, ""),
		Back:  strings.Join(back, ""),
	}
}

func code(x bt.Code) string {
	if len(x.Tag()) == 1 {
		return fmt.Sprintf("<code>%s</code>", x.Text())
	}
	return fmt.Sprintf("<pre>%s</pre>", x.Text())
}
