## inbloom

_inbloom_ - a cross language Bloom filter implementation (https://en.wikipedia.org/wiki/Bloom_filter).

![inbloom](https://raw.githubusercontent.com/EverythingMe/inbloom/master/inbloom.png)

## What's a Bloom Filter?
A Bloom filter is a probabalistic data structure which provides an extremely space-efficient method of representing large sets.
It can have false positives but never false negatives which means a query returns either "possibly in set" or "definitely not in set".

You can tune a Bloom filter to the desired error rate, it's basically a tradeoff between size and accuracy (See example: http://hur.st/bloomfilter). For example, a filter for about 100 keys with 1% error rate can be compressed to just 120 bytes. With 5% error rate it can be compressed to 78 bytes.

## Why Cross Language?
At EverythingMe we have an Android client written in Java, and servers written mostly in Python and Go. When we wanted to pass filters from the client to the server to avoid saving some state on the server side, we needed an efficient implementation that can read and write Bloom Filters in all three languages at least, and found none.

Having such a library allows us to send filters between clients and any server component easily.

So we decided to build on top of an existing simple implementation in C called libbloom (https://github.com/jvirkki/libbloom) and expand it to all 3 langauges.
We chose to use the original C implementation for the Python version only, and **translated the code to pure Java and Go, without calling any C code**.
We chose this approach because the original C code is fairly short and straightforward, so porting it to other languages was a simple task;
and voiding calling C from Java and Go simplifies the build process and executable size in both cases.

## Filter headers

InBloom provides utilities for serializing / deserializing Bloom filters so they can be sent over network.
Since when you create a Bloom filter, you need to initialize it with parameters of expected cardinality and false positive rates,
they are also needed to read a filter written by another party. Instead of choosing fixed parameters in our configurations, we opted for encoding
those parameters as a header when serizlizing the filter. We've added a 16 bit checksum for good measure as part of the header.

### Serialized filter structure:

| Field        | Type            | bits |
| ------------- |:-------------:| -----:|
| checksum      | ushort | 16 |
| errorRate (1/N)| ushort | 16 |
| cardinality   | int     |   21 |
| data          | byte[]  | ? |


### Example Usage

#### Python
```python
import inbloom
import base64
import requests

# Basic usage
bf = inbloom.Filter(entries=100, error=0.01)
bf.add("abc")
bf.add("def")

assert bf.contains("abc")
assert bf.contains("def")
assert not bf.contains("ghi")

bf2 = inbloom.Filter(entries=100, error=0.01, data=bf.buffer())
assert bf2.contains("abc")
assert bf2.contains("def")
assert not bf2.contains("ghi")


# Serialization
payload = 'Yg0AZAAAABQAAAAAACAAEAAIAAAAAAAAIAAQAAgABAA='
assert base64.b64encode(inbloom.dump(inbloom.load(base64.b64decode(payload)))) == payload

# Sending it over HTTP
serialized = base64.b64encode(inbloom.dump(bf))
requests.get('http://api.endpoint.me', params={'filter': serialized})
```

#### Go
```go
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
```

```go
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
```

#### Java
```java
// The basics
BloomFilter bf = new BloomFilter(20, 0.01);
bf.add("foo");
bf.add("bar");

assertTrue(bf.contains("foo"));
assertTrue(bf.contains("bar"));
assertFalse(bf.contains("baz"));


BloomFilter bf2 = new BloomFilter(bf.bf, bf.entries, bf.error);
assertTrue(bf2.contains("foo"));
assertTrue(bf2.contains("bar"));
assertFalse(bf2.contains("baz"));

// Serialization
String serialized = BinAscii.hexlify(BloomFilter.dump(bf));
System.out.printf("Serialized: %s\n", serialized);

String hexPayload = "620d006400000014000000000020001000080000000000002000100008000400";
BloomFilter deserialized = BloomFilter.load(BinAscii.unhexlify(hexPayload));
String dump = BinAscii.hexlify(BloomFilter.dump(deserialized));
System.out.printf("Re-Serialized: %s\n", dump);
assertEquals(dump.toLowerCase(), hexPayload);

assertEquals(deserialized.entries, 20);
assertEquals(deserialized.error, 0.01);
assertTrue(deserialized.contains("abc"));
```
