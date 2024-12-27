package bmutex

import (
	"sync"
)

type Bmutex struct {
	sync.Mutex
	op     sync.Mutex
	wg     sync.WaitGroup
	locked bool
}

func (m *Bmutex) Lock() {
	m.op.Lock()
	m.Mutex.Lock()
	m.locked = true
	m.op.Unlock()
}

func (m *Bmutex) Unlock() {
	m.op.Lock()
	m.Mutex.Unlock()
	m.locked = false
	m.op.Unlock()
}

func (m *Bmutex) TryLock() bool {
	m.op.Lock()
	m.locked = true
  l := m.Mutex.TryLock()
	m.op.Unlock()
  return l
}

func (m *Bmutex) TryUnlock() bool {
	m.op.Lock()
	l := m.Mutex.TryLock()
	m.Mutex.Unlock()
	m.locked = false
	m.op.Unlock()
	return !l
}

func (m *Bmutex) IsLocked() bool {
	m.op.Lock()
  l := m.locked
	m.op.Unlock()
  return l
}

func (m *Bmutex) Protect(f func()) {
	m.op.Lock()
	m.Mutex.Lock()
	m.op.Unlock()
	m.wg.Add(1)
	f()
	m.wg.Done()
	m.Mutex.Unlock()
}

func (m *Bmutex) Queue(f func()) {
	m.op.Lock()
	m.Mutex.Lock()
	m.op.Unlock()
	m.wg.Add(1)
	go func() {
		f()
		m.wg.Done()
		m.Mutex.Unlock()
	}()
}

func (m *Bmutex) Wait() {
	m.wg.Wait()
}
