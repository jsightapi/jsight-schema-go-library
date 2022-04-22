// Autogenerated code!
// DO NOT EDIT!
//
// Generated by OrderedMap generator from the internal/cmd/generator command.

package schema

import (
	"bytes"
	"context"
	"encoding/json"
)

// Set sets a value with specified key.
func (m *objectNodeKeys) Set(k string, v objectNodeKey) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if m.data == nil {
		m.data = map[string]objectNodeKey{}
	}
	if !m.has(k) {
		m.order = append(m.order, k)
	}
	m.data[k] = v
}

// Update updates a value with specified key.
func (m *objectNodeKeys) Update(k string, fn func(v objectNodeKey) objectNodeKey) {
	m.mx.Lock()
	defer m.mx.Unlock()

	if !m.has(k) {
		// Prevent from possible nil pointer dereference if map value type is a
		// pointer.
		return
	}

	m.data[k] = fn(m.data[k])
}

// GetValue gets a value by key.
func (m *objectNodeKeys) GetValue(k string) objectNodeKey {
	m.mx.RLock()
	defer m.mx.RUnlock()

	return m.data[k]
}

// Get gets a value by key.
func (m *objectNodeKeys) Get(k string) (objectNodeKey, bool) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	v, ok := m.data[k]
	return v, ok
}

// Has checks that specified key is set.
func (m *objectNodeKeys) Has(k string) bool {
	m.mx.RLock()
	defer m.mx.RUnlock()

	return m.has(k)
}

func (m *objectNodeKeys) has(k string) bool {
	_, ok := m.data[k]
	return ok
}

// Len returns count of values.
func (m *objectNodeKeys) Len() int {
	m.mx.RLock()
	defer m.mx.RUnlock()

	return len(m.data)
}

func (m *objectNodeKeys) Delete(k string) {
	m.mx.Lock()
	defer m.mx.Unlock()

	m.delete(k)
}

func (m *objectNodeKeys) delete(k string) {
	var kk string
	i := -1

	for i, kk = range m.order {
		if kk == k {
			break
		}
	}

	delete(m.data, k)
	if i != -1 {
		m.order = append(m.order[:i], m.order[i+1:]...)
	}
}

// Filter iterates and changes values in the map.
func (m *objectNodeKeys) Filter(fn filterobjectNodeKeysFunc) {
	m.mx.Lock()
	defer m.mx.Unlock()

	for _, k := range m.order {
		if !fn(k, m.data[k]) {
			m.delete(k)
		}
	}
}

type filterobjectNodeKeysFunc = func(k string, v objectNodeKey) bool

// Iterate acts the same as IterateContext but with background context.
func (m *objectNodeKeys) Iterate() <-chan objectNodeKeysItem {
	return m.IterateContext(context.Background())
}

// IterateContext iterates over map key/values.
// Will block in case of slow consumer.
// Context should be canceled in order to avoid infinity lock.
// Use objectNodeKeys.Map when you have to update value.
func (m *objectNodeKeys) IterateContext(ctx context.Context) <-chan objectNodeKeysItem {
	ch := make(chan objectNodeKeysItem)
	go func() {
		m.mx.RLock()
		defer m.mx.RUnlock()

		for _, k := range m.order {
			select {
			case <-ctx.Done():
				break
			case ch <- objectNodeKeysItem{
				Key:   k,
				Value: m.data[k],
			}:
			}
		}
		close(ch)
	}()
	return ch
}

// objectNodeKeysItem represent single data from the objectNodeKeys.
type objectNodeKeysItem struct {
	Key   string
	Value objectNodeKey
}

var _ json.Marshaler = &objectNodeKeys{}

func (m *objectNodeKeys) MarshalJSON() ([]byte, error) {
	m.mx.RLock()
	defer m.mx.RUnlock()

	var buf bytes.Buffer
	buf.WriteRune('{')

	for i, k := range m.order {
		if i != 0 {
			buf.WriteRune(',')
		}

		// marshal key
		key, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteRune(':')

		// marshal value
		val, err := json.Marshal(m.data[k])
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	buf.WriteRune('}')
	return buf.Bytes(), nil
}
