package bmutex

import (
	"sync"
	"testing"
)

func TestLock(t *testing.T) {
	m := Bmutex{}
	m.Lock()
	if !m.locked {
		t.Error("Expected true, got", m.locked)
	}
	if !m.IsLocked() {
		t.Error("Expected true, got", m.IsLocked())
	}
	a := 0
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		m.Lock()
		a++
		m.Unlock()
		wg.Done()
	}()
	m.Unlock()
	wg.Wait()
	if a != 1 {
		t.Error("Expected 1, got", a)
	}
}

func TestTry(t *testing.T) {
	m := Bmutex{}
	v := m.TryLock()
	if !v {
		t.Error("Expected true, got", v)
	}
	v = m.TryLock()
	if v {
		t.Error("Expected false, got", v)
	}
	v = m.Unlock()
	if !v {
		t.Error("Expected true, got", v)
	}
	v = m.Unlock()
	if v {
		t.Error("Expected false, got", v)
	}
}

func TestProtect(t *testing.T) {
	const iterations = 100
	a := 0
	m := &Bmutex{}
	go func() {
		for i := 0; i < iterations; i++ {
			m.Protect(func() {
				b := a
				b++
				a = b
			})
		}
	}()
	for i := 0; i < iterations; i++ {
		m.Protect(func() {
			b := a
			b++
			a = b
		})
	}
	if a != iterations {
		t.Error("Expected", iterations*2, ", got", a)
	}
}

func TestQueue(t *testing.T) {
	const iterations = 10000
	ch := make(chan int, iterations)
	m := Bmutex{}
	for i := 0; i < iterations; i++ {
		m.Queue(func() {
			ch <- i
		})
	}
	m.Wait()
	for i := 0; i < iterations; i++ {
		v := <-ch
		if v != i {
			t.Error("Expected", i, "got", v)
		}
	}
}

func TestQueueLarge(t *testing.T) {
	const iterations = 1000000
	m := Bmutex{}
	ch := make(chan int, iterations)
	for i := 0; i < iterations; i++ {
		m.Queue(func() {
			ch <- i
		})
	}
	m.Wait()
	for i := 0; i < iterations; i++ {
		v := <-ch
		if v != i {
			t.Error("Expected", i, "got", v)
		}
	}
}

func TestQueueMany(t *testing.T) {
	for j := 0; j < 10; j++ {
		const iterations = 10000
		ch := make(chan int, iterations)
		m := Bmutex{}
		m.QueueMany(func(i int) {
			ch <- i
		}, iterations)
		m.Wait()
		for i := 0; i < iterations; i++ {
			v := <-ch
			if v != i {
				t.Error("Expected", i, "got", v)
			}
		}
	}
}

func TestQueueManyLarge(t *testing.T) {
	for j := 0; j < 10; j++ {
		const iterations = 1000000
		m := Bmutex{}
		ch := make(chan int, iterations)
		m.QueueMany(func(i int) {}, 0)
		go m.QueueMany(func(i int) {
			ch <- i
		}, iterations)
		m.Wait()
		for i := 0; i < iterations; i++ {
			v := <-ch
			if v != i {
				t.Error("Expected", i, "got", v)
			}
		}
	}
}

func TestQueueLargerThanMax(t *testing.T) {
	for j := 0; j < 10; j++ {
		const iterations = 1000000
		m := Bmutex{}
		ch := make(chan int)
		go m.QueueMany(func(i int) {
			ch <- i
		}, iterations)
		m.Wait()
		for i := 0; i < iterations; i++ {
			v := <-ch
			if v != i {
				t.Error("Expected", i, "got", v)
			}
		}
	}
}

func TestWait(t *testing.T) {
	const iterations = 100
	a := 0
	m := Bmutex{}
	for i := 0; i < iterations; i++ {
		m.Queue(func() {
			a = i
		})
	}
	m.Wait()
	a = 0
	for i := 0; i < iterations; i++ {
		m.Queue(func() {
			b := a
			b++
		})
	}
	if a != 0 {
		t.Error("Expected 0, got", a)
	}
}

func TestWaiting(t *testing.T) {
	const iterations = 100
	m := Bmutex{}
	waitChan := make(chan struct{})
	for i := 0; i < iterations; i++ {
		m.Queue(func() {
			<-waitChan
		})
	}
	for i := iterations - 1; i >= 0; i-- {
		waitChan <- struct{}{}
		w := m.Waiting()
		if w != i && w != i-1 {
			t.Error("Expected", i, "got", w)
		}
	}
}

func TestBlocked(t *testing.T) {
	const iterations = 100
	m := Bmutex{}
	a := 0
	done := make(chan struct{})
	go func() {
		for i := 0; i < iterations; i++ {
			m.Lock()
			b := a
			b++
			a = b
			m.Unlock()
		}
		close(done)
	}()
	for i := 0; i < iterations; i++ {
		m.Lock()
		b := a
		b++
		a = b
		m.Unlock()
	}
	<-done
	if a != iterations*2 {
		t.Error("Expected", iterations*2, "got", a)
	}
}

func TestQueueThenMany(t *testing.T) {
	const iterations = 10000
	ch := make(chan int, iterations*12)
	m := Bmutex{}
	for i := 0; i < iterations; i++ {
		m.Queue(func() {
			ch <- 0
		})
		m.QueueMany(func(j int) {
			ch <- j
		}, 10)
	}
	for i := 0; i < iterations; i++ {
		<-ch
		for j := 0; j < 10; j++ {
			v := <-ch
			if v != j {
				t.Error("Expected", j, "got", v)
			}
		}
	}
}

func BenchmarkLock(b *testing.B) {
	m := Bmutex{}
	a := 0
	for i := 0; i < b.N; i++ {
		m.Lock()
		a += i
		m.Unlock()
	}
}

func BenchmarkLockBlocked(b *testing.B) {
	m := Bmutex{}
	a := 0
	go func() {
		for i := 0; i < b.N; i++ {
			m.Lock()
			a++
			m.Unlock()
		}
	}()
	for i := 0; i < b.N; i++ {
		m.Lock()
		a++
		m.Unlock()
	}
}

func BenchmarkTryLockSucceed(b *testing.B) {
	m := Bmutex{}
	a := 0
	for i := 0; i < b.N; i++ {
		m.TryLock()
		a += i
		m.Unlock()
	}
}

func BenchmarkTryLockFail(b *testing.B) {
	m := Bmutex{}
	m.Lock()
	for i := 0; i < b.N; i++ {
		m.TryLock()
	}
}

func BenchmarkProtect(b *testing.B) {
	m := Bmutex{}
	a := 0
	for i := 0; i < b.N; i++ {
		m.Protect(func() {
			a += i
		})
	}
}

func BenchmarkQueue(b *testing.B) {
	m := Bmutex{}
	ch := make(chan struct{}, b.N)
	go func() {
		for i := 0; i < b.N; i++ {
			ch <- struct{}{}
		}
	}()
	for i := 0; i < b.N; i++ {
		m.Queue(func() {
			<-ch
		})
	}
	m.Wait()
}

func BenchmarkQueueMany(b *testing.B) {
	m := Bmutex{}
	ch := make(chan struct{}, b.N)
	go func() {
		for i := 0; i < b.N; i++ {
			ch <- struct{}{}
		}
	}()
	m.QueueMany(func(i int) {
		<-ch
	}, b.N)
	m.Wait()
}

func BenchmarkCompareLock(b *testing.B) {
	m := sync.Mutex{}
	a := 0
	for i := 0; i < b.N; i++ {
		m.Lock()
		a++
		m.Unlock()
	}
}

func BenchmarkCompareLockBlocked(b *testing.B) {
	m := sync.Mutex{}
	a := 0
	go func() {
		for i := 0; i < b.N; i++ {
			m.Lock()
			a++
			m.Unlock()
		}
	}()
	for i := 0; i < b.N; i++ {
		m.Lock()
		a++
		m.Unlock()
	}
}

func BenchmarkCompareQueue(b *testing.B) {
	m := sync.Mutex{}
	wg := sync.WaitGroup{}
	ch := make(chan struct{}, b.N)
	go func() {
		for i := 0; i < b.N; i++ {
			ch <- struct{}{}
		}
	}()
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			m.Lock()
			<-ch
			m.Unlock()
			wg.Done()
		}()
	}
	wg.Wait()
}
