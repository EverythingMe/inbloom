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

func TestMarshal(t *testing.T) {
	bf, err := NewFilter(20, 0.01)
	if err != nil {
		t.Fatal(err)
	}
	bf.Add("abc")

	serizliaed := fmt.Sprintf("%x", bf.Marshal())
	expected := "620d006400000014000000000020001000080000000000002000100008000400"
	if serizliaed != expected {
		t.Errorf("Expected %s, got %s", expected, serizliaed)
	}

	bfds, err := Unmarshal(bf.Marshal())
	if err != nil {
		t.Fatal(err)
	}

	serizliaed = fmt.Sprintf("%x", bfds.Marshal())
	if serizliaed != expected {
		t.Errorf("Expected %s, got %s", expected, serizliaed)
	}
	t.Logf("DESERIALIZED: %X\n", bfds.Marshal())

	// Test for bad checksum

	data := bfds.Marshal()
	data[0] = 0xff
	data[1] = 0xff

	if _, err = Unmarshal(data); err == nil {
		t.Error("Should have failed on bad checksum")
	} else {
		t.Log(err)
	}

	data[2] = 0xff
	if _, err = Unmarshal(data); err == nil {
		t.Error("Should have failed on bad size")
	} else {
		t.Log(err)
	}

	data = data[:4]
	if _, err = Unmarshal(data); err == nil {
		t.Error("Should have failed on bad data")
	} else {
		t.Log(err)
	}

}

func ExampleBloomFilter() {

	// create a blank filter - expecting 20 members and an error rate of 1/100
	f, err := NewFilter(20, 0.01)
	if err != nil {
		panic(err)
	}

	// the size of the filter
	fmt.Println(f.Len())

	// insert some values
	f.Add("foo")
	f.Add("bar")

	// test for existence of keys
	fmt.Println(f.Contains("foo"))
	fmt.Println(f.Contains("wat"))

	fmt.Println("marshaled data:", f.MarshalBase64())

	// Output:
	// 24
	// true
	// false
	// marshaled data: oU4AZAAAABQAAAAAAEIAABEAGAQAAgAgAAAwEAAJAAA=

}

func ExampleMarshalUnmarshal() {

	// a 20 cardinality 0.01 precision filter with "foo" and "bar" in it
	data := "oU4AZAAAABQAAAAAAEIAABEAGAQAAgAgAAAwEAAJAAA="

	// load it from base64
	f, err := UnmarshalBase64(data)
	if err != nil {
		panic(err)
	}

	// test it...
	fmt.Println(f.Contains("foo"))
	fmt.Println(f.Contains("wat"))
	fmt.Println(f.Len())

	// dump to pure binary
	fmt.Printf("%x\n", f.Marshal())
	// Output:
	// true
	// false
	// 24
	// a14e006400000014000000000042000011001804000200200000301000090000

}
