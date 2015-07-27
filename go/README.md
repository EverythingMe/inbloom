# inbloom
--
    import "github.com/EverythingMe/inbloom-go"

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

#### func  NewFilterFromData

```go
func NewFilterFromData(data []byte, entries int, errorRate float64) (*BloomFilter, error)
```
NewFilterFromData creates a bloom filter from an existing data buffer, created
by another instance of this library (probably in another language).

If the length of the data does not fit the number of entries and error rate, we
return an error. If data is nil we allocate a new filter

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
