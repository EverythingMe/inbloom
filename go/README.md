# inbloom
--
    import "github.com/EverythingMe/inbloom/go"

Package inbloom implements a portable bloom filter that can export and import
data to and from implementations of the same library in different languages.

## Usage

#### type BloomFilter

```go
type BloomFilter struct {
}
```

BloomFilter is our implementation of a simple dynamically sized bloom filter.

This code was adapted to Go from the libbloom C library -
https://github.com/jvirkki/libbloom

#### func  NewFilter

```go
func NewFilter(entries int, errorRate float64) (*BloomFilter, error)
```
NewFilter creates an empty bloom filter, with the given expected number of
entries, and desired error rate. The number of hash functions and size of the
filter are calculated from these 2 parameters

#### func  Unmarshal

```go
func Unmarshal(data []byte) (*BloomFilter, error)
```
Unmarshal reads a binary dump of an inbloom filter with its header, and returns
the resulting filter. Since this is a dump containing size and precisin
metadata, you do not need to specify them.

If the data is corrupt or the buffer is not complete, we return an error

#### func (*BloomFilter) Add

```go
func (f *BloomFilter) Add(key string) bool
```
Add adds a key to the filter

#### func (*BloomFilter) Contains

```go
func (f *BloomFilter) Contains(key string) bool
```
Contains returns true if a key exists in the filter

#### func (*BloomFilter) Len

```go
func (f *BloomFilter) Len() int
```
Len returns the number of BYTES in the filter

#### func (*BloomFilter) Marshal

```go
func (f *BloomFilter) Marshal() []byte
```
Marshal dumps the filter to a byte array, with a header containing the error
rate, cardinality and a checksum. This data can be passed to another inbloom
filter over the network, and thus the other end can open the data without the
user having to pass the filter size explicitly. See Unmarshal for reading these
dumpss
