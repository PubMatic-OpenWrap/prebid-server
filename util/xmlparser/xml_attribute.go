package xmlparser

type xmlAttribute struct {
	key, value xmlTagIndex
}

func (a xmlAttribute) Key(in []byte) []byte {
	return in[a.key.si:a.key.ei]
}

func (a xmlAttribute) Value(in []byte) []byte {
	return in[a.value.si:a.value.ei]
}

// String function print key and value
func (a xmlAttribute) String(in []byte) string {
	return string(in[a.key.si : a.value.ei+1])
}

func parseAttributes(in []byte, si, ei int) (attributes []xmlAttribute) {
	found := true
	for found {
		var attr xmlAttribute

		//parsing key
		attr.key.si, attr.key.ei, found = _parseKey(in, si, ei)
		if found {
			//parsing = separator
			i := attr.key.ei
			for ; i < ei && whitespace[in[i]]; i = i + 1 {
			}
			if i > ei || in[i] != '=' {
				//invalid
				break
			}
			//parsing value
			attr.value.si, attr.value.ei, found = _parseValue(in, i+1, ei)
		}
		if found {
			attributes = append(attributes, attr)
			si = attr.value.ei + 1
		}
	}
	return
}

func _parseKey(in []byte, si, ei int) (int, int, bool) {
	len := ei
	for ; si < len && whitespace[in[si]]; si = si + 1 {
	}
	for ei = si; ei < len && name[in[ei]]; ei = ei + 1 {
	}
	if ei < len && (alpha[in[si]] || in[si] == '_') {
		return si, ei, true
	}
	return 0, 0, false
}

func _parseValue(in []byte, si, ei int) (int, int, bool) {
	len := ei
	for ; si < len && whitespace[in[si]]; si = si + 1 {
	}

	if si >= len || !(in[si] == '\'' || in[si] == '"') {
		return 0, 0, false
	}

	quote := in[si]
	for ei = si + 1; ei < len && in[ei] != quote; ei = ei + 1 {
	}

	if ei < len {
		return si + 1, ei, true
	}
	return 0, 0, false
}
