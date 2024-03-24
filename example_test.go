package annot_test

import (
	"fmt"

	"github.com/meyermarcel/annot"
)

func Example() {
	fmt.Println("The quick brown fox jumps over the lazy dog")
	fmt.Println(annot.String(
		&annot.Annot{Col: 6, Lines: []string{"adjective"}},
		&annot.Annot{Col: 12, Lines: []string{"adjective"}},
		&annot.Annot{Col: 36, Lines: []string{"adjective"}},
	))
	// Output:
	// The quick brown fox jumps over the lazy dog
	//       ↑     ↑                       ↑
	//       │     └─ adjective            └─ adjective
	//       │
	//       └─ adjective
}
