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
  v = m.TryUnlock()
  if !v {
    t.Error("Expected true, got", v)
  }
  v = m.TryUnlock()
  if v {
    t.Error("Expected false, got", v)
  }
}


func TestProtect(t *testing.T) {
	a := 0
	m := &Bmutex{}
	m.Protect(func() {
		a++
	})
	if a != 1 {
		t.Error("Expected 1, got", a)
	}
}

func TestQueue(t *testing.T) {
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
	if a != iterations {
		t.Error("Expected", iterations, "got", a)
	}
}

func TestQueueOrder(t *testing.T) {
	const iterations = 1000
	a := 0
	m := Bmutex{}
	for i := 0; i < iterations; i++ {
		m.Queue(func() {
			a = iterations
		})
	}
	m.Wait()
	if a != iterations {
		t.Error("Expected", iterations, "got", a)
	}
}

func TestWait(t *testing.T) {
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
