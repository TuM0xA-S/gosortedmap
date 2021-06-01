# sorted map for go

* idiomatic api(?)
* O(logn) read-write operations
* sorted iteration for free

### example
```go
package main

import (
	"fmt"
	"strings"

	ds "github.com/TuM0xA-S/gosortedmap"
)

func main() {
	// initialize map with sorting callback
	wordcount := ds.NewSortedMap(func(a, b interface{}) int {
		as, bs := a.(string), b.(string)
		return strings.Compare(as, bs)
	})

	// data
	words := []string{
		"hello", "world", "abcd", "bcde", "abcd", "hello", "abcd",
	}

	// count words
	for _, word := range words {
		if cnt, ok := wordcount.Get(word); ok {
			wordcount.Set(word, cnt.(int)+1)
		} else {
			wordcount.Set(word, 1)
		}
	}

	for e := range wordcount.AsChan() {
		fmt.Printf("%v: %v\n", e.Key, e.Value)
	}
	// output:
	// abcd: 3
	// bcde: 1
	// hello: 2
	// world: 1
}


```
