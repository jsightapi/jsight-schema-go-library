package validator

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTreeLog_nodeToString(t *testing.T) {
	v := &objectValidator{}
	node1 := &objectValidator{}
	node2 := &objectValidator{}

	s := treeLog{
		tree: map[validator][]validator{
			v: {
				node1,
				node2,
			},
			node1: {
				&objectValidator{},
			},
		},
	}.nodeToString(v, 0)
	assert.Equal(t, strings.TrimLeft(`
object [%!p(<nil>)]
	object [%!p(<nil>)]
object [%!p(<nil>)]
`, "\n"), s)
}
