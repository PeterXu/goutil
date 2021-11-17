package util

import (
	"sync"
)

// If a <= b, then a is left of b.
type fnMapLess func(interface{}, interface{}) bool
type fnMapEqual func(interface{}, interface{}) bool

type Map struct {
	sync.Mutex
	maps map[interface{}]interface{}
	keys []interface{}
	less fnMapLess
}

// @param less: ordered if not null, else random
func NewMap(less fnMapLess) *Map {
	return &Map{
		maps: make(map[interface{}]interface{}),
		keys: make([]interface{}, 0, 0),
		less: less,
	}
}

func (m *Map) Put(key, value interface{}) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.maps[key]; !ok {
		m.keys = append(m.keys, key)
		if m.less != nil {
			max_idx := len(m.keys) - 1
			for idx := max_idx - 1; idx >= 0; idx -= 1 {
				item := m.keys[idx]
				if m.less(item, key) { // stable sort
					break
				} else {
					m.keys[idx+1] = item
					m.keys[idx] = key
				}
			}
		}
	}

	m.maps[key] = value
}

func (m *Map) Get(key interface{}) (value interface{}, ok bool) {
	m.Lock()
	defer m.Unlock()

	value, ok = m.maps[key]
	return
}

func (m *Map) Remove(key interface{}) {
	m.Lock()
	defer m.Unlock()

	if _, ok := m.maps[key]; !ok {
		return
	}

	delete(m.maps, key)
	for i := range m.keys {
		if m.keys[i] == key {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			break
		}
	}
}

func (m *Map) Empty() bool {
	m.Lock()
	defer m.Unlock()

	return len(m.maps) == 0
}

func (m *Map) Keys() []interface{} {
	m.Lock()
	defer m.Unlock()

	return m.keys
}

func (m *Map) Values() []interface{} {
	m.Lock()
	defer m.Unlock()

	values := make([]interface{}, len(m.maps))
	for i, key := range m.keys {
		values[i] = m.maps[key]
	}
	return values
}

func (m *Map) Size() int {
	m.Lock()
	defer m.Unlock()

	return len(m.maps)
}

func (m *Map) Find(key interface{}, equal fnMapEqual) interface{} {
	m.Lock()
	defer m.Unlock()

	if key != nil && equal != nil {
		for k, v := range m.maps {
			if equal(k, key) {
				return v
			}
		}
	}
	return nil
}
