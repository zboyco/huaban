package model

import "sync"

type Message struct {
	arr []string
	sync.Mutex
}

func (m *Message) Reset() {
	m.Lock()
	defer m.Unlock()
	m.arr = make([]string, 0)
}

func (m *Message) Add(v string) {
	m.Lock()
	defer m.Unlock()
	m.arr = append(m.arr, v)
}

func (m *Message) Pick() []string {
	m.Lock()
	defer m.Unlock()
	arr := m.arr
	m.arr = make([]string, 0)
	return arr
}
