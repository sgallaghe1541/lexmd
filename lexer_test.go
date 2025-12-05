package lexer

import (
	"testing"
)

func TestLex(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantItems []item
	}{
		{
			name: "Test # Headings",
			input: `##### Valid Heading
# Another Valid Heading
###### Invalid Heading. this should be text.
#Another invalid heading. this should also be text...
`,
			wantItems: []item{
				{
					typ: itemHeading,
					val: "##### Valid Heading",
				},
				{
					typ: itemHeading,
					val: "# Another Valid Heading",
				},
				{
					typ: itemText,
					val: "###### Invalid Heading. this should be text.",
				},
				{
					typ: itemText,
					val: "#Another invalid heading. this should also be text...",
				},
				{
					typ: itemEOF,
				},
			},
		},
		{
			name: "Test NewLines",
			input: `##### Valid Heading

###### Invalid Heading. this should be text.



#Another invalid heading. this should also be text...

`,
			wantItems: []item{
				{
					typ: itemHeading,
					val: "##### Valid Heading",
				},
				{
					typ: itemBreak,
					val: "\n\n",
				},
				{
					typ: itemText,
					val: "###### Invalid Heading. this should be text.",
				},
				{
					typ: itemBreak,
					val: "\n\n\n\n",
				},
				{
					typ: itemText,
					val: "#Another invalid heading. this should also be text...",
				},
				{
					typ: itemBreak,
					val: "\n\n",
				},
				{
					typ: itemEOF,
				},
			},
		},
	}

	// start tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, tokens := lex(test.name, test.input)
			for i := 0; ; i++ {
				token, more := <-tokens
				if !more {
					break
				}
				expectedItem := test.wantItems[i]
				if token.typ != expectedItem.typ {
					t.Errorf("expected to receive itemType: %d value: %v. instead received itemType: %d value: %v",
						expectedItem.typ,
						expectedItem.val,
						token.typ,
						token.val,
					)
				}
			}
		})
	}
}
