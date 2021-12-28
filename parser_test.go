package brashtag_test

import (
	"os"
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

func TestUnexpectedL(t *testing.T) {
	data, err := os.ReadFile("./testdata/fail1.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = bt.Parse(string(data))
	if err == nil {
		t.Fatal("expected error")
	}
	t.Log(err)
}

func TestFail2(t *testing.T) {
	// TODO
	t.Fail()
}
