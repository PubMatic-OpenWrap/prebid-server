package xmlparser

import (
	"bytes"
	"sort"
)

type xmlOperation struct {
	startIndex int
	endIndex   int
	parameters []any
}

/*
TODO: makesure not 2 operations overlaps
*/
type XMLUpdater struct {
	ops []xmlOperation
}

func (xu *XMLUpdater) Insert(index int, parameters ...any) {
	xu.ops = append(xu.ops, xmlOperation{startIndex: index, endIndex: index, parameters: parameters})
}

func (xu *XMLUpdater) Replace(start int, end int, parameters ...any) {
	xu.ops = append(xu.ops, xmlOperation{startIndex: start, endIndex: end, parameters: parameters})
}

func (xu *XMLUpdater) Build(buf *bytes.Buffer, in []byte, cb func(*bytes.Buffer, ...any)) {
	//sort operations based on index
	sort.SliceStable(xu.ops[:], func(i, j int) bool {
		return (xu.ops[i].startIndex < xu.ops[j].startIndex ||
			(xu.ops[i].startIndex == xu.ops[j].startIndex && xu.ops[i].endIndex < xu.ops[j].endIndex))
	})

	offset := 0
	for _, op := range xu.ops {
		if offset < op.startIndex {
			buf.Write(in[offset:op.startIndex])
			offset = op.endIndex
		}
		cb(buf, op.parameters...)
	}
	buf.Write(in[offset:])
}
