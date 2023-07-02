package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"

	bt "github.com/pratikdeoghare/brashtag"
)

var (
	cardTmpl = template.Must(template.ParseFiles("./card-template2.html"))
)

type card struct {
	front string
	back  string
	score int
}

type deck struct {
	cards []card
	curr  int
	perm  []int
}

func (d *deck) Build() {
	var scores []int
	for _, c := range d.cards {
		scores = append(scores, c.score)
	}
	w := newWeights(scores)
	sort.Sort(w)
	d.perm = w.idx
}

func (d *deck) Score(x int) {
	fmt.Println(d.curr, x)
	d.cards[d.curr].score += x
}

func (d *deck) Next() card {
	d.curr = (d.curr + 1) % len(d.perm)
	return d.cards[d.perm[d.curr]]
}

func (d *deck) handler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)
	if d.curr != -1 {
		switch r.URL.Path {
		case "/good":
			d.Score(1)
		case "/bad":
			d.Score(-1)
		}
	}

	card := d.Next()

	c := struct {
		Front template.HTML
		Back  template.HTML
	}{
		Front: template.HTML(card.front),
		Back:  template.HTML(card.back),
	}

	err := cardTmpl.Execute(w, c)
	if err != nil {
		log.Fatal(err)
	}

}

type weights struct {
	ws  []int
	idx []int
}

func newWeights(ws []int) weights {
	idx := make([]int, len(ws))
	for i := 0; i < len(ws); i++ {
		idx[i] = i
	}
	return weights{
		ws:  ws,
		idx: idx,
	}
}

func (w weights) Len() int {
	return len(w.ws)
}

func (w weights) Less(i, j int) bool {
	return w.ws[w.idx[i]] < w.ws[w.idx[j]]
}

func (w weights) Swap(i, j int) {
	w.idx[i], w.idx[j] = w.idx[j], w.idx[i]
}

var _ sort.Interface = &weights{}

func main() {
	var cards string
	flag.StringVar(&cards, "cards", "", "name of file with cards")
	flag.Parse()

	text, err := ioutil.ReadFile(cards)
	if err != nil {
		panic(err)
	}

	tree, err := bt.Parse(string(text))
	if err != nil {
		panic(err)
	}

	var d deck
	d.curr = -1
	for _, kid := range tree.(bt.Bag).Kids() {
		switch x := kid.(type) {
		case bt.Bag:
			d.cards = append(d.cards, makeCard(x))
		}
	}
	fmt.Println("total cards: ", len(d.cards))
	d.Build()

	dumpScores := func() {
		i := 0
		for _, kid := range tree.(bt.Bag).Kids() {
			switch x := kid.(type) {
			case bt.Bag:
				SetKid(x, "score", fmt.Sprint(d.cards[i].score))
				i++
			}
		}

		var text []string
		for _, kid := range tree.(bt.Bag).Kids() {
			text = append(text, kid.String())
		}

		err := ioutil.WriteFile(cards, []byte(strings.Join(text, "")), 0644)
		if err != nil {
			panic(err)
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			dumpScores()
			os.Exit(0)
		}
	}()

	http.HandleFunc("/", d.handler)
	http.HandleFunc("/favicon.ico", noop)
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func noop(w http.ResponseWriter, r *http.Request) {

}

func getBags(x bt.Bag, tags ...string) []int {
	var bags []int
	for i, k := range x.Kids() {
		switch a := k.(type) {
		case bt.Bag:
			for _, tag := range tags {
				if a.Tag() == tag {
					bags = append(bags, i)
					break
				}
			}
		}
	}
	return bags
}

func SetKid(x bt.Bag, path string, val string) {
	bags := getBags(x, "score")
	for _, bag := range bags {
		x.RemoveKid(bag)
	}

	x.AddKids(bt.NewBag(path, bt.NewBlob(val)))
}

func makeCard(t bt.Bag) card {
	var front []string
	var b bt.Bag
loop:
	for _, kid := range t.Kids() {
		switch x := kid.(type) {
		case bt.Blob:
			front = append(front, x.Text())
		case bt.Code:
			front = append(front, code(x))
		case bt.Bag:
			b = x
			break loop
		}
	}

	var back []string
	for _, kid := range b.Kids() {
		switch x := kid.(type) {
		case bt.Code:
			back = append(back, code(x))
		case bt.Blob:
			back = append(back, x.Text())
		}
	}

	score := 0
	bags := getBags(t, "score")
	if len(bags) != 0 {
		score = parseScore(t.Kids()[bags[0]].(bt.Bag))
	}

	return card{
		front: strings.Join(front, ""),
		back:  strings.Join(back, ""),
		score: score,
	}
}

func parseScore(t bt.Bag) int {
	s := ""
	for _, k := range t.Kids() {
		switch k.(type) {
		case bt.Blob:
			s += k.String()
		}
	}

	s = strings.TrimSpace(s)
	fmt.Println(s)
	v, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return v
}

func code(x bt.Code) string {
	// for math
	trimmed := strings.TrimSpace(x.Text())
	if trimmed != "" {
		if trimmed[0] == '$' {
			return x.Text()
		}
	}

	if len(x.Tag()) == 1 {
		return fmt.Sprintf("<code>%s</code>", x.Text())
	}
	return fmt.Sprintf("<pre>%s</pre>", x.Text())
}
