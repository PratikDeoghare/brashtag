package brashtag_test

import (
	"testing"

	bt "github.com/pratikdeoghare/brashtag"
)

func TestParser(t *testing.T) {
	r := bt.NewBag("hello")
	a := bt.NewCode("$$", "echo $PATH")
	b := bt.NewBlob("some text")
	r.AddKids(a, b)

	r2, err := bt.Parse(r.String())
	if err != nil {
		t.Fatal(err)
	}

	v := r2.(bt.Bag).Kids()[0]

	if v.String() != r.String() {
		t.Fail()
	}
}
