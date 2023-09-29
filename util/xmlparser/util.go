package xmlparser

import "bytes"

var whitespace [256]bool //<space>, \r, \n, \t
var alnum [256]bool      //a-z, A-Z, 0-9
var alpha [256]bool      //a-z, A-Z
var num [256]bool        //0-9
var name [256]bool       //a-z, A-Z, 0-9, _, -

func init() {
	whitespace[' '] = true
	whitespace['\r'] = true
	whitespace['\n'] = true
	whitespace['\t'] = true

	//name
	name['_'] = true
	name['-'] = true

	//alnum
	for ch := 'a'; ch <= 'z'; ch++ {
		alnum[ch] = true
		alpha[ch] = true
		name[ch] = true
	}
	for ch := 'A'; ch <= 'Z'; ch++ {
		alnum[ch] = true
		alpha[ch] = true
		name[ch] = true
	}
	for ch := '0'; ch <= '9'; ch++ {
		alnum[ch] = true
		num[ch] = true
		name[ch] = true
	}
}

func _trimCDATA(in []byte, start, end int) (si, ei int) {
	//`#whitespaces#<![CDATA[ data ]]>#whitespaces#`
	si, ei = _trim(in, start, end)
	//search for <![CDATA[
	found := bytes.HasPrefix(in[si:ei], []byte(cdataStart))
	if found {
		si = si + len(cdataStart)
		ei = ei - len(cdataEnd)
		//if si+len(cdataStart) > ei-len(cdataEnd) {}
		//si, ei = _trim(in, si, ei)
		return si, ei
	}
	return start, end
}

func _trim(in []byte, start, end int) (int, int) {
	//remove heading whitespaces
	for ; start < end && whitespace[in[start]]; start++ {
	}
	//remove trailing whitespaces
	for ; end > start && whitespace[in[end-1]]; end-- {
	}
	return start, end
}
