package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBufferPool(t *testing.T) {
	const capacity = 1024
	p := NewBufferPool(capacity)
	require.NotNil(t, p)

	require.NotNil(t, p.pool.New)

	b := p.Get()
	assert.Equal(t, 0, b.Len())
	assert.Equal(t, capacity, b.Cap())
}
