package xmlparser

type xmlTokenType int

const (
	UnknownXMLToken    xmlTokenType = iota //unknown token
	StartXMLToken                          //<ns:xmltag k1="v1" k2='v2'>
	InlineXMLToken                         //<ns:xmltag/>
	EndXMLToken                            //</ns:xmltag>
	ProcessingXMLToken                     //<? text ?>
	CommentsXMLToken                       //<!-- text -->
	CDATAXMLToken                          //<![CDATA[ text ]]>
	DOCTYPEXMLToken                        //<!DOCTYPE [ text ]>
	TextToken                              //text
)

const (
	cdataStart = "<![CDATA["
	cdataEnd   = "]]>"
)

func (t xmlTokenType) String() string {
	return []string{
		"UnknownTokenType",
		"StartXMLToken",
		"InlineXMLToken",
		"EndXMLToken",
		"ProcessingXMLToken",
		"CommentsXMLToken",
		"CDATAToken",
		"DOCTYPEToken",
		"TextToken",
	}[t]
}

type xmlTagIndex struct {
	si, ei int
}

type XMLToken struct {
	start, end xmlTagIndex
	name, text xmlTagIndex
}

func NewXMLToken(ssi, sei, esi, eei int) XMLToken {
	return XMLToken{
		start: xmlTagIndex{si: ssi, ei: sei},
		end:   xmlTagIndex{si: esi, ei: eei},
	}
}

func (t *XMLToken) Text(in []byte, removeCDATA bool) []byte {
	if t.text.si == 0 {
		t.text.si, t.text.ei = t.start.ei, t.end.si
		if removeCDATA {
			t.text.si, t.text.ei = _trimCDATA(in, t.text.si, t.text.ei)
		}
	}
	return in[t.text.si:t.text.ei]
}

func (t XMLToken) Name(in []byte) []byte {
	if t.name.si == 0 {
		t.name.si, t.name.ei = getTokenNameIndex(in, t.start.si, 1)
	}
	return in[t.name.si:t.name.ei]
}

func (t XMLToken) StartTagOffset() (start, end int) {
	return t.start.si, t.start.ei
}

func (t XMLToken) EndTagOffset() (start, end int) {
	return t.end.si, t.end.ei
}

func (t XMLToken) ParseAttribute(in []byte) []xmlAttribute {
	//check for inline token
	return parseAttributes(in[:], t.start.si, t.start.ei)
}

func (t XMLToken) IsInline() bool {
	return (t.start.si == t.end.si)
}

func getTokenType(in []byte, index int) xmlTokenType {
	if index >= len(in) {
		return UnknownXMLToken
	}
	//remove whitespace
	ch := in[index]
	switch ch {
	case '/':
		return EndXMLToken
	case '!':
		//remove whitespace
		if index+1 >= len(in) {
			return UnknownXMLToken
		}
		ch1 := in[index+1]
		switch ch1 {
		case '-':
			return CommentsXMLToken
		case '[':
			return CDATAXMLToken
		case 'D':
			return DOCTYPEXMLToken
		}
	case '?':
		return ProcessingXMLToken
	default:
		if alpha[ch] {
			return StartXMLToken
		}
	}
	return UnknownXMLToken
}

func getTokenNameIndex(in []byte, startIndex int, offset int) (si, ei int) {
	si = startIndex + offset
	for i := si; i < len(in); i++ {
		if in[i] == '>' || whitespace[in[i]] || in[i] == '/' {
			return si, i
		} else if in[i] == ':' {
			si = i + 1
		}
	}
	return si, si //not found
}

func getTokenEndIndex(in []byte, startIndex int, ttype xmlTokenType) (int, bool) {
	index := -1
	inline := false
	//TODO: loops can be avoided
	switch ttype {
	case StartXMLToken:
		// read until >, no need for read until ?> for processing tokens
		for i := startIndex + 1; i < len(in); i++ {
			if in[i] == '>' {
				if in[i-1] == '/' {
					inline = true
				}
				//found end tag
				index = i
				break
			}
		}
	case EndXMLToken:
		// read until >, no need for read until ?> for processing tokens
		for i := startIndex + 1; i < len(in); i++ {
			if in[i] == '>' {
				//found end tag
				index = i
				break
			}
		}
	case ProcessingXMLToken:
		// read until >, no need for read until ?> for processing tokens
		for i := startIndex + 1; i < len(in); i++ {
			if in[i] == '>' && in[i-1] == '?' {
				//found end tag
				index = i
				break
			}
		}
	case CommentsXMLToken:
		// read until found -->
		for i := startIndex + 1; i < len(in); i++ {
			if in[i] == '>' && in[i-1] == '-' && in[i-2] == '-' {
				//found end tag
				index = i
				break
			}
		}
	case CDATAXMLToken:
		// read until ]]> /*<![CDATA[ 25.00 ]]>*/
		for i := startIndex + 1; i < len(in); i++ {
			if in[i] == '>' && in[i-1] == ']' && in[i-2] == ']' {
				/*
					TODO: Special handling (https://en.wikipedia.org/wiki/CDATA#Nesting)
					input: <![CDATA[ data ]]> data ]]>
					replace ]]> with ]]]]><![CDATA[>
					output: <![CDATA[ data ]]]]><![CDATA[> data ]]>
					action: ignore if found ']]]]><![CDATA[>'
				*/
				//found end tag
				index = i
				break
			}
		}
	case DOCTYPEXMLToken:
		//read until ]>
		for i := startIndex + 1; i < len(in); i++ {
			if in[i] == '>' && in[i-1] == ']' {
				//found end tag
				index = i
				break
			}
		}
	default:
		//read token based on tokentype
		for i := startIndex + 1; i < len(in); i++ {
			if in[i] == '>' {
				index = i
				break
			}
		}
	}
	return index + 1, inline
}
