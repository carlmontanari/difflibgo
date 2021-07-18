difflibgo
=========

difflibgo is a partial port of the Python standard library [`difflib` module](https://github.com/python/cpython/blob/main/Lib/difflib.py)
, and builds upon the [`go-difflib` module](https://github.com/pmezard/go-difflib).

This implementation includes the `SequenceMatcher` class, as well as the `compare` function of 
the `Differ` class. The only goal of this project is to be able to produce a slice of "diff" strings
with appropriate tags to represent additions and deletions as the original Python module does.

## A simple Example

```go
package main

import (
	"github.com/carlmontanari/difflibgo/difflibgo"
	"fmt"
)

func main() {
	seqA := []string{"something", "differentthing", "more", "more", "more"}
	seqB := []string{"sometihng", "anotherthing", "more"}

	d := difflibgo.Differ{}

	dLines := d.Compare(seqA, seqB)
	for _, v := range dLines {
		fmt.Printf("%s\n", v)
	}
}
```

Which outputs:

```bash
- something
?       -

+ sometihng
?      +

- differentthing
+ anotherthing
  more
- more
- more
```
