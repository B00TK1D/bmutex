package bmutex

import (
	"sync"
)

var MAXQUEUE = 100000

type Bmutex struct {
	sync.Mutex
	op      sync.Mutex
	ch      chan func()
	locked  bool
	worker  bool
	waiting int
}

func (m *Bmutex) Lock() {
	m.waiting++
	m.Mutex.Lock()
	m.waiting--
	m.locked = true
}

func (m *Bmutex) Unlock() bool {
	m.op.Lock()
	l := m.Mutex.TryLock()
	m.Mutex.Unlock()
	m.locked = false
	m.op.Unlock()
	return !l
}

func (m *Bmutex) TryLock() bool {
	m.op.Lock()
	m.locked = true
	l := m.Mutex.TryLock()
	m.op.Unlock()
	return l
}

func (m *Bmutex) IsLocked() bool {
	m.op.Lock()
	l := m.locked
	m.op.Unlock()
	return l
}

func (m *Bmutex) Protect(f func()) {
	m.waiting++
	m.Mutex.Lock()
	m.waiting--
	m.locked = true
	f()
	m.op.Lock()
	m.Mutex.Unlock()
	m.locked = false
	m.op.Unlock()
}

func (m *Bmutex) startWorker() {
	go func() {
		var t func()
		for {
			select {
			case t = <-m.ch:
				m.Mutex.Lock()
				t()
				m.Mutex.Unlock()
			}
		}
	}()
}

func (m *Bmutex) Queue(f func()) {
	m.op.Lock()
	defer m.op.Unlock()
	if m.ch == nil {
		m.ch = make(chan func(), MAXQUEUE)
		m.startWorker()
	}
	select {
	case m.ch <- f:
	default:
		defer func() {
			m.ch <- f
		}()
	}
}

func (m *Bmutex) QueueMany(f func(int), n int) {
	m.op.Lock()
	defer m.op.Unlock()
	if m.ch == nil {
		if n > MAXQUEUE {
			m.ch = make(chan func(), n+1)
		} else {
			m.ch = make(chan func(), MAXQUEUE)
		}
		m.startWorker()
	}
	for i := 0; i < n; i++ {
		select {
		case m.ch <- func() { f(i) }:
		default:
			k := i
			defer func() {
				for j := k; j < n; j++ {
					m.ch <- func() { f(j) }
				}
			}()
			i = n
		}
	}
}

func (m *Bmutex) Waiting() int {
	return len(m.ch) + m.waiting
}

func (m *Bmutex) Wait() {
	done := make(chan struct{})
	m.Queue(func() {
		close(done)
	})
	<-done
}
