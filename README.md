[![build](https://github.com/meyermarcel/annot/actions/workflows/build.yml/badge.svg)](https://github.com/meyermarcel/annot/actions/workflows/build.yml) [![Go Report Card](https://goreportcard.com/badge/github.com/meyermarcel/annot)](https://goreportcard.com/report/github.com/meyermarcel/annot)

# Annot

Annotate a string line with arrows leading to a description.

```
The quick brown fox jumps over the lazy dog
      ↑     ↑                       ↑
      │     └─ adjective            └─ adjective
      │
      └─ adjective
```

## Installation

```
go get github.com/meyermarcel/annot
```

## Example

```go
package main

import (
	"fmt"

	"github.com/meyermarcel/annot"
)

func main() {
	fmt.Println("The quick brown fox jumps over the lazy dog")
	fmt.Println(annot.String(
		&annot.Annot{Col: 6, Lines: []string{"adjective"}},
		&annot.Annot{Col: 12, Lines: []string{"adjective"}},
		&annot.Annot{Col: 36, Lines: []string{"adjective"}},
	))
}
```

Output:

```
The quick brown fox jumps over the lazy dog
      ↑     ↑                       ↑
      │     └─ adjective            └─ adjective
      │
      └─ adjective
```