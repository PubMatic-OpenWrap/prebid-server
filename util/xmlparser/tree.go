package xmlparser

import (
	"bytes"
	"fmt"
	"strings"
)

type comparable[T any] func(string, T) bool

type treeNode[T any] struct {
	data                     T
	index, first, last, next int
}

func (n treeNode[T]) Data() T {
	return n.data
}

func (n treeNode[T]) Index() int {
	return n.index
}

type tree[T any] struct {
	nodes []treeNode[T] //first node will be always last node
}

/*
insert function to insert node n in parent node
NOTE: always re-fetch parent object everytime when inserting new object
*/
func (t *tree[T]) insert(parent *treeNode[T], n treeNode[T]) {
	if len(t.nodes) == 0 {
		t.nodes = append(t.nodes, treeNode[T]{index: 0, first: -1, last: -1, next: -1})
	}

	n.index = len(t.nodes)
	if parent == nil {
		parent = &t.nodes[0]
	}

	if parent.first == -1 {
		//first node
		parent.first = n.index
	} else {
		//subsequent node
		t.nodes[parent.last].next = n.index
	}
	parent.last = n.index

	t.nodes = append(t.nodes, n)
}

func (t *tree[T]) reset() {
	t.nodes = t.nodes[:0]
}

func (t *tree[T]) getMatchedChildrens(parent int, match func(T) bool, cb func(*treeNode[T])) {
	if parent < 0 || parent >= len(t.nodes) {
		return
	}
	for i := t.nodes[parent].first; i != -1; i = t.nodes[i].next {
		if match == nil || match(t.nodes[i].data) {
			cb(&t.nodes[i])
		}
	}
}

func (t *tree[T]) getChildrens(parent int, cb func(*treeNode[T])) {
	t.getMatchedChildrens(parent, nil, cb)
}

func (t *tree[T]) _get(parent int,
	path []string,
	match comparable[T],
	cb func(*treeNode[T])) {

	t.getMatchedChildrens(parent,
		func(node T) bool {
			return match(path[0], node)
		},
		func(node *treeNode[T]) {
			if len(path) == 1 { // last element
				cb(node)
			} else {
				t._get(node.index, path[1:], match, cb)
			}
		},
	)
}

/*
get function returns always last element
TODO: write changes for return first element
*/
func (t *tree[T]) get(parent *treeNode[T], path []string, match comparable[T]) (result *treeNode[T]) {
	parentIndex := 0
	if parent != nil {
		parentIndex = parent.index
	}

	t._get(parentIndex, path[:], match, func(node *treeNode[T]) {
		result = node
	})
	return
}

func (t *tree[T]) getAll(parent *treeNode[T], path []string, match comparable[T]) (result []*treeNode[T]) {
	parentIndex := 0
	if parent != nil {
		parentIndex = parent.index
	}

	t._get(parentIndex, path[:], match, func(node *treeNode[T]) {
		result = append(result, node)
	})
	return
}

/* Printing Function */
func (t *tree[T]) _print(buf *bytes.Buffer, index, indent int, f func(T) string) {
	buf.WriteByte('\n')
	buf.WriteString(strings.Repeat("\t", indent))
	buf.WriteByte('|')
	buf.WriteString(f(t.nodes[index].data))
	indent++
	for i := t.nodes[index].first; i != -1; i = t.nodes[i].next {
		t._print(buf, i, indent, f)
	}
}

func (t *tree[T]) print(f func(T) string) string {
	root := len(t.nodes) - 1
	buf := bytes.Buffer{}
	t._print(&buf, root, 0, f)
	return buf.String()
}

func (t *tree[T]) printRaw(f func(T) string) string {
	buf := bytes.Buffer{}
	for i, node := range t.nodes {
		buf.WriteString(fmt.Sprintf("\n%d:<%d,%d,%d>:%s", i, node.first, node.last, node.next, f(node.data)))
	}
	return buf.String()
}
