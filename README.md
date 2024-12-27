# bmutex
A Better Mutex

Example usage:

```golang
package main

import (
    "fmt"

    "github.com/B00TK1D/bmutex"
)

func main() {
    const iterations = 1000
	a := 0
	m := Bmutex{}
	for i := 0; i < iterations; i++ {
		m.Queue(func() {
			b := a
			b++
			a = b
		})
	}
	m.Wait()
	fmt.Println("Expected", iterations, "got", a)
}
```
