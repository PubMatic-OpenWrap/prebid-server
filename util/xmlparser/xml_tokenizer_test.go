package xmlparser

import "testing"

func TestXMLTokenizer_Parse(t *testing.T) {

	tests := []struct {
		name    string
		in      string
		wantErr bool
	}{
		{
			name: `simple`,
			in: `<root>
					<element1>Value1</element1>
					<element2>Value2</element2>
				</root>`,
			wantErr: false,
		},
		{
			name: `nested`,
			in: `<root>
					<parent>
						<child>Value</child>
					</parent>
				</root>`,
			wantErr: false,
		},
		{
			name: `attributes`,
			in: `<root attribute="value">
					<element>Content</element>
				</root>`,
			wantErr: false,
		},
		{
			name: `empty_elements`,
			in: `<root>
					<emptyElement />
				</root>`,
			wantErr: false,
		},
		{
			name: `cdata`,
			in: `<root>
					<![CDATA[This is a CDATA section. It can contain <tags> and special characters &]]>
				</root>`,
			wantErr: false,
		},
		{
			name:    `entity_reference`,
			in:      `<root>This is an entity reference: &amp;</root>`,
			wantErr: false,
		},
		{
			name: `comments`,
			in: `<root>
					<!-- This is a comment -->
					<element>Content</element>
				</root>`,
			wantErr: false,
		},
		{
			name: `processing_instructions`,
			in: `<?xml version="1.0" encoding="UTF-8"?>
				<root>
					<?processing-instruction attribute="value"?>
					<element>Content</element>
				</root>`,
			wantErr: false,
		},
		{
			name: `namespace`,
			in: `<ns:root xmlns:ns="http://example.com">
					<ns:element>Content</ns:element>
				</ns:root>`,
			wantErr: false,
		},
		{
			name:    `mixed_content`,
			in:      `<root>This is <b>mixed</b> content with <i>HTML</i> tags.</root>`,
			wantErr: false,
		},
		{
			name: `complex_structure`,
			in: `<books>
					<book>
						<title>Book 1</title>
						<author>Author 1</author>
					</book>
					<book>
						<title>Book 2</title>
						<author>Author 2</author>
					</book>
				</books>`,
			wantErr: false,
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sp := &XMLTokenizer{}
			if err := sp.Parse([]byte(tt.in), nil); (err != nil) != tt.wantErr {
				t.Errorf("XMLTokenizer.Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
