# bmutex
A Better Mutex for Golang

## Features

* Prevents panics when unlocking an unlocked mutex (returns result as bool instead)
* Provides clean `Protect()` wrapper for better readability and safety
* Implements `Queue()` and `QueueMany()` functions for cleanly maintaining order of background operations
* Provides `Wait()` function to block until all currently-queued operations are complete
* Exposes `Waiting()` function to report the number of operations waiting for the mutex, and `IsLocked()` to report if the mutex is currently locked without blocking

## Example Usage

```golang
package main

import (
    "fmt"

    "github.com/B00TK1D/bmutex"
)

func main() {
	m := Bmutex{}

    m.Protect(func() {
        fmt.Println("Protected operation")
    })

	for i := 0; i < 20; i++ {
		m.Queue(func() {
			fmt.Println("Parallel 1 queued operation", i)
		})
	}
    for i := 0; i < 20; i++ {
		m.Queue(func() {
			fmt.Println("Parallel 1 queued operation", i)
		})
	}
	m.Wait()
	fmt.Println("Expected", iterations, "got", a)
}
```

## Performance

Based on basic benchmarks, this library provides nearly identical performance for uncontested `Lock()` and `Unlock()` operations (with no performance loss for `Protect()`),
and roughly half the performance for contested `Unlock()` operations (due to the need to call `TryLock()` before `Unlock()` to prevent panics).
However, it performs up to 4.75x faster than basic standard library alternatives for `Queue()` and `QueueMany()` operations, due to its use of a single goroutine to run serial operations.


```
# go test -bench=. -benchmem -cover
goos: darwin
goarch: arm64
pkg: github.com/B00TK1D/bmutex
cpu: Apple M1 Max
BenchmarkLock
BenchmarkLock-10                        70716147                16.99 ns/op            0 B/op          0 allocs/op
BenchmarkLockBlocked
BenchmarkLockBlocked-10                 10203492               127.7 ns/op             0 B/op          0 allocs/op
BenchmarkTryLockSucceed
BenchmarkTryLockSucceed-10              35324506                32.30 ns/op            0 B/op          0 allocs/op
BenchmarkTryLockFail
BenchmarkTryLockFail-10                 74538021                16.13 ns/op            0 B/op          0 allocs/op
BenchmarkProtect
BenchmarkProtect-10                     67977595                17.79 ns/op            0 B/op          0 allocs/op
BenchmarkQueue
BenchmarkQueue-10                        5386003               232.3 ns/op            16 B/op          1 allocs/op
BenchmarkQueueMany
BenchmarkQueueMany-10                    7566622               185.5 ns/op            40 B/op          2 allocs/op
BenchmarkCompareLock
BenchmarkCompareLock-10                 74170032                16.12 ns/op            0 B/op          0 allocs/op
BenchmarkCompareLockBlocked
BenchmarkCompareLockBlocked-10          13502494                87.42 ns/op            0 B/op          0 allocs/op
BenchmarkCompareQueue
BenchmarkCompareQueue-10                 1652316               733.7 ns/op            42 B/op          1 allocs/op
PASS
coverage: 100.0% of statements
ok      github.com/B00TK1D/bmutex       18.740s
```
