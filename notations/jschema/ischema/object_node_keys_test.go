package ischema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObjectNodeKeys_Set(t *testing.T) {
	cc := map[string]struct {
		keys         *ObjectNodeKeys
		given        ObjectNodeKey
		expectedErr  string
		expectedData []ObjectNodeKey
	}{
		"new isn't shortcut, without duplication": {
			keys: fakeObjectNodeKeys(
				ObjectNodeKey{Key: "bar"},
				ObjectNodeKey{Key: "fizz", IsShortcut: true},
			),
			given: ObjectNodeKey{Key: "foo", IsShortcut: false},
			expectedData: []ObjectNodeKey{
				{Key: "bar"},
				{Key: "fizz", IsShortcut: true},
				{Key: "foo"},
			},
		},
		"new is shortcut, without duplication": {
			keys: fakeObjectNodeKeys(
				ObjectNodeKey{Key: "bar"},
				ObjectNodeKey{Key: "fizz", IsShortcut: true},
			),
			given: ObjectNodeKey{Key: "foo", IsShortcut: true},
			expectedData: []ObjectNodeKey{
				{Key: "bar"},
				{Key: "fizz", IsShortcut: true},
				{Key: "foo", IsShortcut: true},
			},
		},

		"new isn't shortcut, exists isn't shortcut": {
			keys:        fakeObjectNodeKeys(ObjectNodeKey{Key: "foo"}),
			given:       ObjectNodeKey{Key: "foo"},
			expectedErr: "Duplicate keys (foo) in the schema",
		},

		"new isn't shortcut, exists is shortcut": {
			keys:  fakeObjectNodeKeys(ObjectNodeKey{Key: "foo", IsShortcut: true}),
			given: ObjectNodeKey{Key: "foo"},
			expectedData: []ObjectNodeKey{
				{Key: "foo", IsShortcut: true},
				{Key: "foo"},
			},
		},

		"new is shortcut, exists isn't shortcut": {
			keys:  fakeObjectNodeKeys(ObjectNodeKey{Key: "foo"}),
			given: ObjectNodeKey{Key: "foo", IsShortcut: true},
			expectedData: []ObjectNodeKey{
				{Key: "foo"},
				{Key: "foo", IsShortcut: true},
			},
		},

		"new is shortcut, exists is shortcut": {
			keys:        fakeObjectNodeKeys(ObjectNodeKey{Key: "foo", IsShortcut: true}),
			given:       ObjectNodeKey{Key: "foo", IsShortcut: true},
			expectedErr: "Duplicate keys (foo) in the schema",
		},
	}

	for n, c := range cc {
		t.Run(n, func(t *testing.T) {
			if c.expectedErr != "" {
				assert.PanicsWithError(t, c.expectedErr, func() {
					c.keys.Set(c.given)
				})
			} else {
				c.keys.Set(c.given)

				assert.Equal(t, c.keys.Data, c.expectedData)
			}
		})
	}
}

func fakeObjectNodeKeys(kk ...ObjectNodeKey) *ObjectNodeKeys {
	keys := newObjectNodeKeys()
	for _, k := range kk {
		keys.Set(k)
	}
	return keys
}
