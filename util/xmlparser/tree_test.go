package xmlparser

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_insert(t *testing.T) {
	iTree := &tree[int]{}

	newiTreeNode := func(data, index, first, last, next int) treeNode[int] {
		return treeNode[int]{
			data:  data,
			index: index,
			first: first,
			last:  last,
			next:  next,
		}
	}

	assert.Equal(t, 0, len(iTree.nodes))

	//adding 1st level elements
	iTree.insert(nil, newiTreeNode(10, 0, -1, -1, -1))
	iTree.insert(nil, newiTreeNode(20, 0, -1, -1, -1))
	iTree.insert(nil, newiTreeNode(30, 0, -1, -1, -1))

	//adding 2nd level elements
	iTree.insert(&iTree.nodes[1], newiTreeNode(11, 0, -1, -1, -1))
	iTree.insert(&iTree.nodes[1], newiTreeNode(12, 0, -1, -1, -1))
	iTree.insert(&iTree.nodes[2], newiTreeNode(21, 0, -1, -1, -1))
	iTree.insert(&iTree.nodes[2], newiTreeNode(22, 0, -1, -1, -1))
	iTree.insert(&iTree.nodes[3], newiTreeNode(300, 0, -1, -1, -1))
	iTree.insert(&iTree.nodes[3], newiTreeNode(300, 0, -1, -1, -1))

	//assert tree
	assert.Equal(t, newiTreeNode(0, 0, 1, 3, -1), iTree.nodes[0])
	assert.Equal(t, newiTreeNode(10, 1, 4, 5, 2), iTree.nodes[1])
	assert.Equal(t, newiTreeNode(20, 2, 6, 7, 3), iTree.nodes[2])
	assert.Equal(t, newiTreeNode(30, 3, 8, 9, -1), iTree.nodes[3])
	assert.Equal(t, newiTreeNode(11, 4, -1, -1, 5), iTree.nodes[4])
	assert.Equal(t, newiTreeNode(12, 5, -1, -1, -1), iTree.nodes[5])
	assert.Equal(t, newiTreeNode(21, 6, -1, -1, 7), iTree.nodes[6])
	assert.Equal(t, newiTreeNode(22, 7, -1, -1, -1), iTree.nodes[7])
	assert.Equal(t, newiTreeNode(300, 8, -1, -1, 9), iTree.nodes[8])
	assert.Equal(t, newiTreeNode(300, 9, -1, -1, -1), iTree.nodes[9])

	assert.Equal(t, 10, len(iTree.nodes))

	var actual []int
	iTree.getChildrens(0, func(tn *treeNode[int]) {
		actual = append(actual, tn.data)
	})
	assert.Equal(t, []int{10, 20, 30}, actual)

	result := iTree.get(nil, []string{"20", "21"}, func(value string, data int) bool {
		return strconv.Itoa(data) == value
	})
	assert.Equal(t, result.data, 21)

	result = iTree.get(nil, []string{"30", "300"}, func(value string, data int) bool {
		return strconv.Itoa(data) == value
	})
	assert.Equal(t, iTree.nodes[9], *result)

	results := iTree.getAll(nil, []string{"30", "300"}, func(value string, data int) bool {
		return strconv.Itoa(data) == value
	})
	expected := []*treeNode[int]{&iTree.nodes[8], &iTree.nodes[9]}
	assert.Equal(t, expected, results)

	iTree.reset()
	assert.Equal(t, 0, len(iTree.nodes))
}
