package xmlparser

import "bytes"

type ElementTree = tree[XMLToken]

type XMLReader struct {
	in     []byte
	tree   ElementTree
	parser *XMLTokenizer
}

func NewXMLReader(path *xpath) *XMLReader {
	return &XMLReader{
		tree:   ElementTree{},
		parser: NewXMLTokenizer(path),
	}
}

func (xr *XMLReader) tokenHandler(name string, parent *Element, child Element) {
	xr.tree.insert(parent, child)
}

func (xr *XMLReader) Parse(in []byte) error {
	xr.tree.reset()
	xr.in = in
	return xr.parser.Parse(in, xr.tokenHandler)
}

func (xr *XMLReader) FindElement(parent *Element, path ...string) *Element {
	return xr.tree.get(parent, path[:], func(s string, t XMLToken) bool {
		return bytes.Equal(t.Name(xr.in), []byte(s))
	})
}

func (xr *XMLReader) FindElements(parent *Element, path ...string) (result []*Element) {
	return xr.tree.getAll(parent, path[:], func(s string, t XMLToken) bool {
		return bytes.Equal(t.Name(xr.in), []byte(s))
	})
}

func (xr *XMLReader) GetAttribute(node *Element, key string) (value string) {
	attr := node.data.ParseAttribute(xr.in)
	for _, at := range attr {
		if bytes.Equal(at.Key(xr.in), []byte(key)) {
			return string(at.Value(xr.in))
		}
	}
	return ""
}

func (xr *XMLReader) GetText(node *Element, removeCDATA bool) (value string) {
	return string(node.data.Text(xr.in, removeCDATA))
}

func (xr *XMLReader) getXML(in []byte) string {
	buf := bytes.Buffer{}
	start := 0
	for _, node := range xr.tree.nodes {
		buf.Write(in[start:node.data.end.ei])
		start = node.data.end.ei
	}
	return buf.String()
}
