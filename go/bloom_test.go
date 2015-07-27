package inbloom

import (
	"fmt"
	"testing"
)

func TestBloom(t *testing.T) {

	bf, err := NewFilter(20, 0.01)
	if err != nil {
		t.Fatal(err)
	}

	keys := []string{"foo", "bar", "foosdfsdfs", "fossdfsdfo", "foasdfasdfasdfasdfo", "foasdfasdfasdasdfasdfasdfasdfasdfo"}

	faux := []string{"goo", "gar", "gaz"}

	for _, k := range keys {
		if bf.Add(k) == true {
			t.Errorf("adding %s returned true", k)
		}
	}

	t.Logf("Bloom filter params: %X", bf.bf)
	for _, k := range keys {
		if !bf.Contains(k) {
			t.Error("not containig ", k)
		}

	}

	for _, k := range faux {
		if bf.Contains(k) {
			t.Error("containig faux key", k)
		}
	}

	expected := "02000C0300C2246913049E040002002000017614002B0002"
	actual := fmt.Sprintf("%X", bf.bf)
	if actual != expected {
		t.Errorf("expected\n%s\nactual\n%s", expected, actual)
	}
}
