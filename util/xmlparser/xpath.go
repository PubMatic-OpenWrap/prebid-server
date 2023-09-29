package xmlparser

import (
	"bytes"
	"strings"
)

type xpath struct {
	data   string
	childs map[string]*xpath
}

type XPath = xpath

func (n *xpath) add(path []string) {
	tempNode := n
	for _, key := range path {
		childNode, ok := tempNode.childs[key]
		if !ok {
			childNode = &xpath{
				data:   key,
				childs: make(map[string]*xpath),
			}
			tempNode.childs[key] = childNode
		}
		tempNode = childNode
	}
}

func (n *xpath) get(key string) (val *xpath) {
	/*
		if len(n.childs) <= 5 {
			for data, p := range n.childs {
				if data == key {
					return p
				}
			}
		}
	*/
	return n.childs[key]
}

func (n *xpath) _print(buf *bytes.Buffer, indent int) {
	buf.WriteByte('\n')
	buf.WriteString(strings.Repeat("\t", indent))
	buf.WriteByte('|')
	buf.WriteString(n.data)
	indent++
	for _, child := range n.childs {
		child._print(buf, indent)
	}
}

func (n *xpath) print() string {
	buf := bytes.Buffer{}
	n._print(&buf, 0)
	return buf.String()
}

func GetXPath(path [][]string) *xpath {
	rule := &xpath{data: "/", childs: make(map[string]*xpath)}
	for _, p := range path {
		rule.add(p)
	}
	return rule
}
