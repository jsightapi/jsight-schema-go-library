package scanner

import (
	"fmt"
	"j/schema/bytes"
	"j/schema/fs"
	"j/schema/internal/lexeme"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newScanner(t *testing.T) {
	const content = "bar"
	contentLen := len(content)

	f := fs.NewFile("foo", bytes.Bytes(content))

	s := newScanner(f)

	assert.NotNil(t, s.step)
	assert.Equal(t, f, s.file)
	assert.Equal(t, content, string(s.data))
	assert.Equal(t, contentLen, int(s.dataSize))
	assert.NotNil(t, s.returnToStep)
	assert.NotNil(t, s.stack)
	assert.NotNil(t, s.finds)
	assert.Equal(t, context{Type: contextTypeInitial}, s.context)
}

func BenchmarkScanner_Length(b *testing.B) {
	file := fs.NewFile("", bytes.Bytes(`
/*
{}
*/

{
	"k": 123
}

/*
{}
*/

some text
`))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		NewSchemaScanner(file, true).Length()
	}
}

func TestScanner_Length(t *testing.T) {
	tests := map[string]uint{
		"1 // {}":                             7,
		"1 // {}\n":                           7,
		"1 // {}\n\n":                         7,
		"1 // {} \n\n text":                   7,
		"/* {} */ 123 /* {} */ text":          21,
		"/* {} */ 123 /* {} */ /* {} */ text": 30,
	}

	for given, expected := range tests {
		t.Run(given, func(t *testing.T) {
			assert.NotPanics(t, func() {
				actual := NewSchemaScanner(fs.NewFile("", bytes.Bytes(given)), true).
					Length()
				assert.Equal(t, expected, actual)
			})
		})
	}
}

func TestScanner_Next(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := []validResults{
			{`12.34`, []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd}},
			{` 12.34 `, []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd}},
			{"12.34\n", []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.NewLine}},
			{"12.34\r\n", []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.NewLine, lexeme.NewLine}},
			{"12.34\r\n", []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.NewLine, lexeme.NewLine}},
			{"12.34 ", []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd}},
			{"12.34 \r\n", []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.NewLine, lexeme.NewLine}},
			{`"str"`, []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd}},
			{`"str" `, []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd}},
			{`"\u0000"`, []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd}},
			{`"\\" `, []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd}},
			{`true`, []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd}},
			{`false`, []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd}},
			{`null`, []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd}},
			{`-1`, []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd}},
			{`0.123`, []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd}},
			{`-0.123`, []lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd}},
			{`[]`, []lexeme.LexEventType{lexeme.ArrayBegin, lexeme.ArrayEnd}},
			{`[ ]`, []lexeme.LexEventType{lexeme.ArrayBegin, lexeme.ArrayEnd}},
			{`[1 ]`, []lexeme.LexEventType{lexeme.ArrayBegin, lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd, lexeme.ArrayEnd}},
			{`[{}]`, []lexeme.LexEventType{lexeme.ArrayBegin, lexeme.ArrayItemBegin, lexeme.ObjectBegin, lexeme.ObjectEnd, lexeme.ArrayItemEnd, lexeme.ArrayEnd}},
			{`[[]]`, []lexeme.LexEventType{lexeme.ArrayBegin, lexeme.ArrayItemBegin, lexeme.ArrayBegin, lexeme.ArrayEnd, lexeme.ArrayItemEnd, lexeme.ArrayEnd}},
			{`{}`, []lexeme.LexEventType{lexeme.ObjectBegin, lexeme.ObjectEnd}},
			{`{} `, []lexeme.LexEventType{lexeme.ObjectBegin, lexeme.ObjectEnd}},
			{` {} `, []lexeme.LexEventType{lexeme.ObjectBegin, lexeme.ObjectEnd}},
			{`{"foo":"bar"}`, []lexeme.LexEventType{lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd, lexeme.ObjectEnd}},
			{` { "foo" : "bar" } `, []lexeme.LexEventType{lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd, lexeme.ObjectEnd}},
			{
				`["",[]]`,
				[]lexeme.LexEventType{
					lexeme.ArrayBegin,
					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.ArrayItemBegin, lexeme.ArrayBegin, lexeme.ArrayEnd, lexeme.ArrayItemEnd,
					lexeme.ArrayEnd,
				},
			},
			{
				`{"foo": "bar", "key": 1}`,
				[]lexeme.LexEventType{
					lexeme.ObjectBegin,
					lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd,
					lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd,
					lexeme.ObjectEnd,
				},
			},
			{
				`[1,"str",false]`,
				[]lexeme.LexEventType{
					lexeme.ArrayBegin,
					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.ArrayEnd,
				},
			},
			{
				`{"foo": [1,"str",false]}`,
				[]lexeme.LexEventType{
					lexeme.ObjectBegin,
					lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd,
					lexeme.ObjectValueBegin,
					lexeme.ArrayBegin,
					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.ArrayEnd,
					lexeme.ObjectValueEnd,
					lexeme.ObjectEnd,
				},
			},
			{`{
	
		"foo"
	
		:
	
		123
	
		}`,
				[]lexeme.LexEventType{
					lexeme.ObjectBegin,
					lexeme.NewLine,
					lexeme.NewLine,
					lexeme.ObjectKeyBegin,
					lexeme.ObjectKeyEnd,
					lexeme.NewLine,
					lexeme.NewLine,
					lexeme.NewLine,
					lexeme.NewLine,
					lexeme.ObjectValueBegin,
					lexeme.LiteralBegin,
					lexeme.LiteralEnd,
					lexeme.ObjectValueEnd,
					lexeme.NewLine,
					lexeme.NewLine,
					lexeme.ObjectEnd},
			},
			{`
		{
			"a": 1,
			"b": [2,3,4],
			"c": 5
		}`,
				[]lexeme.LexEventType{
					lexeme.NewLine,
					lexeme.ObjectBegin,
					lexeme.NewLine,
					lexeme.ObjectKeyBegin,
					lexeme.ObjectKeyEnd,
					lexeme.ObjectValueBegin,
					lexeme.LiteralBegin,
					lexeme.LiteralEnd,
					lexeme.ObjectValueEnd,
					lexeme.NewLine,
					lexeme.ObjectKeyBegin,
					lexeme.ObjectKeyEnd,
					lexeme.ObjectValueBegin,
					lexeme.ArrayBegin,
					lexeme.ArrayItemBegin,
					lexeme.LiteralBegin,
					lexeme.LiteralEnd,
					lexeme.ArrayItemEnd,
					lexeme.ArrayItemBegin,
					lexeme.LiteralBegin,
					lexeme.LiteralEnd,
					lexeme.ArrayItemEnd,
					lexeme.ArrayItemBegin,
					lexeme.LiteralBegin,
					lexeme.LiteralEnd,
					lexeme.ArrayItemEnd,
					lexeme.ArrayEnd,
					lexeme.ObjectValueEnd,
					lexeme.NewLine,
					lexeme.ObjectKeyBegin,
					lexeme.ObjectKeyEnd,
					lexeme.ObjectValueBegin,
					lexeme.LiteralBegin,
					lexeme.LiteralEnd,
					lexeme.ObjectValueEnd,
					lexeme.NewLine,
					lexeme.ObjectEnd,
				},
			},
			{`
		[
			1,
			{"k": 2},
			3
		]`,
				[]lexeme.LexEventType{
					lexeme.NewLine,
					lexeme.ArrayBegin,
					lexeme.NewLine,
					lexeme.ArrayItemBegin,
					lexeme.LiteralBegin,
					lexeme.LiteralEnd,
					lexeme.ArrayItemEnd,
					lexeme.NewLine,
					lexeme.ArrayItemBegin,
					lexeme.ObjectBegin,
					lexeme.ObjectKeyBegin,
					lexeme.ObjectKeyEnd,
					lexeme.ObjectValueBegin,
					lexeme.LiteralBegin,
					lexeme.LiteralEnd,
					lexeme.ObjectValueEnd,
					lexeme.ObjectEnd,
					lexeme.ArrayItemEnd,
					lexeme.NewLine,
					lexeme.ArrayItemBegin,
					lexeme.LiteralBegin,
					lexeme.LiteralEnd,
					lexeme.ArrayItemEnd,
					lexeme.NewLine,
					lexeme.ArrayEnd,
				},
			},
			{`"str" // comment`, []lexeme.LexEventType{
				lexeme.LiteralBegin,
				lexeme.LiteralEnd,
				lexeme.InlineAnnotationBegin,
				lexeme.InlineAnnotationTextBegin,
				lexeme.InlineAnnotationTextEnd,
				lexeme.InlineAnnotationEnd,
			}},
			{
				`12.34`,
				[]lexeme.LexEventType{lexeme.LiteralBegin, lexeme.LiteralEnd},
			},
			{
				`12.34 // comment`,
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,
					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
				},
			},
			{
				`"str" // comment`,
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,
					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
				},
			},
			{
				`123// comment`,
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,
					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
				},
			},
			{
				"12.34 // comment\n",
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,

					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,
				},
			},
			{
				`"str" // comment` + "\n",
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,

					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,
				},
			},
			{
				"12.34 // comment\n\n",
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,

					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,
					lexeme.NewLine,
				},
			},
			{
				"12.34 // comment\r\n",
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,

					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,
					lexeme.NewLine,
				},
			},
			{
				"12.34\n// comment",
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,
					lexeme.NewLine,

					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
				},
			},
			{
				"[1,2,3]",
				[]lexeme.LexEventType{
					lexeme.ArrayBegin,
					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.ArrayEnd,
				},
			},
			{
				`{
			"q": 1, // comment
			"w": 2
		}`,
				[]lexeme.LexEventType{
					lexeme.ObjectBegin,
					lexeme.NewLine,

					lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd,
					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,

					lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd,
					lexeme.NewLine,

					lexeme.ObjectEnd,
				},
			},
			{`[
			1, // comment
			2
		]`,
				[]lexeme.LexEventType{
					lexeme.ArrayBegin,
					lexeme.NewLine,

					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,

					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.NewLine,

					lexeme.ArrayEnd,
				},
			},
			{`// AAA some comment

		[ // the beginning of the array
			1, // comment
			2 // comment
		]
`,
				[]lexeme.LexEventType{
					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,
					lexeme.NewLine,

					lexeme.ArrayBegin,
					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,

					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,

					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,

					lexeme.ArrayEnd,
					lexeme.NewLine,
				},
			},
			{`// BBB some comment
	
		[// the beginning of the array
			1,// comment
			2 // comment
		]

		`,
				[]lexeme.LexEventType{
					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,
					lexeme.NewLine,

					lexeme.ArrayBegin,
					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,

					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,

					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,

					lexeme.ArrayEnd,
					lexeme.NewLine,
					lexeme.NewLine,
				},
			},
			{`	// 111
	
{ // 222
	"k" // 444
	: // 555
	[ // 666
		"val" // 777
	]
}
		`,
				[]lexeme.LexEventType{
					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,
					lexeme.NewLine,

					lexeme.ObjectBegin,

					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,

					lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd,

					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,
					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,

					lexeme.ObjectValueBegin,
					lexeme.ArrayBegin,

					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,

					lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
					lexeme.InlineAnnotationBegin, lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd, lexeme.InlineAnnotationEnd,
					lexeme.NewLine,

					lexeme.ArrayEnd,
					lexeme.ObjectValueEnd,

					lexeme.NewLine,

					lexeme.ObjectEnd,

					lexeme.NewLine,
				},
			},
			{
				`123 // {"min": 1}`,
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,

					lexeme.InlineAnnotationBegin,
					lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd, lexeme.ObjectEnd,
					lexeme.InlineAnnotationEnd,
				},
			},
			{
				`123 // {min: 1}`,
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,

					lexeme.InlineAnnotationBegin,
					lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd, lexeme.ObjectEnd,
					lexeme.InlineAnnotationEnd,
				},
			},
			{
				`123 // {min: 1,}`,
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,

					lexeme.InlineAnnotationBegin,
					lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd, lexeme.ObjectEnd,
					lexeme.InlineAnnotationEnd,
				},
			},
			{
				"123 // {min: 1,} \n",
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,

					lexeme.InlineAnnotationBegin,
					lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd, lexeme.ObjectEnd,
					lexeme.InlineAnnotationEnd,
					lexeme.NewLine,
				},
			},
			{
				`123 // {min: 1} - text`,
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,

					lexeme.InlineAnnotationBegin,
					lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd, lexeme.ObjectEnd,
					lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd,
					lexeme.InlineAnnotationEnd,
				},
			},
			{
				`123 // {min: 1}-text`,
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,

					lexeme.InlineAnnotationBegin,
					lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd, lexeme.ObjectEnd,
					lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd,
					lexeme.InlineAnnotationEnd,
				},
			},
			{
				`123 // {"min": 1, max: 999,} - text`,
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,

					lexeme.InlineAnnotationBegin,
					lexeme.ObjectBegin,
					lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd,
					lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd,
					lexeme.ObjectEnd,
					lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd,
					lexeme.InlineAnnotationEnd,
				},
			},
			{
				`/*
{
  k: 2
}
*/

"a few multi-line comments in the schema"

/* {} */`,
				[]lexeme.LexEventType{
					lexeme.MultiLineAnnotationBegin, lexeme.NewLine,
					lexeme.ObjectBegin, lexeme.NewLine,
					lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd, lexeme.NewLine,
					lexeme.ObjectEnd, lexeme.NewLine,
					lexeme.MultiLineAnnotationEnd,

					lexeme.NewLine,
					lexeme.NewLine,

					lexeme.LiteralBegin, lexeme.LiteralEnd,

					lexeme.NewLine,
					lexeme.NewLine,

					lexeme.MultiLineAnnotationBegin,
					lexeme.ObjectBegin,
					lexeme.ObjectEnd,
					lexeme.MultiLineAnnotationEnd,
				},
			},
			{
				`111 // {mixed: [{type: "string"}, {type: "integer"}]}`,
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,

					lexeme.InlineAnnotationBegin,
					lexeme.ObjectBegin,

					lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd,

					lexeme.ObjectValueBegin,
					lexeme.ArrayBegin,
					lexeme.ArrayItemBegin,
					lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd, lexeme.ObjectEnd,
					lexeme.ArrayItemEnd,
					lexeme.ArrayItemBegin,
					lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd, lexeme.ObjectEnd,
					lexeme.ArrayItemEnd,
					lexeme.ArrayEnd,
					lexeme.ObjectValueEnd,

					lexeme.ObjectEnd,
					lexeme.InlineAnnotationEnd,
				},
			},
			{
				`111 // {mixed: [{type: "string"}, {type: "integer"}],}`,
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,

					lexeme.InlineAnnotationBegin,
					lexeme.ObjectBegin,

					lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd,

					lexeme.ObjectValueBegin,
					lexeme.ArrayBegin,
					lexeme.ArrayItemBegin,
					lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd, lexeme.ObjectEnd,
					lexeme.ArrayItemEnd,
					lexeme.ArrayItemBegin,
					lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd, lexeme.ObjectEnd,
					lexeme.ArrayItemEnd,
					lexeme.ArrayEnd,
					lexeme.ObjectValueEnd,

					lexeme.ObjectEnd,
					lexeme.InlineAnnotationEnd,
				},
			},
			{
				`/*
{
	k1:1 // {k2:2} - txt
}
*/

"inline comment within a multi-line comment"`,
				[]lexeme.LexEventType{
					lexeme.MultiLineAnnotationBegin, lexeme.NewLine,
					lexeme.ObjectBegin, lexeme.NewLine,

					lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd,

					lexeme.InlineAnnotationBegin,
					lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd, lexeme.ObjectEnd,
					lexeme.InlineAnnotationTextBegin, lexeme.InlineAnnotationTextEnd,
					lexeme.InlineAnnotationEnd,

					lexeme.NewLine,

					lexeme.ObjectEnd, lexeme.NewLine,
					lexeme.MultiLineAnnotationEnd,

					lexeme.NewLine,
					lexeme.NewLine,

					lexeme.LiteralBegin, lexeme.LiteralEnd,
				},
			},
			{
				`"multi-line comments in a single string after the literal" /* {} */`,
				[]lexeme.LexEventType{
					lexeme.LiteralBegin, lexeme.LiteralEnd,
					lexeme.MultiLineAnnotationBegin, lexeme.ObjectBegin, lexeme.ObjectEnd, lexeme.MultiLineAnnotationEnd,
				},
			},
			{
				`/* {} */ "multi-line comments in a single string before the literal"`,
				[]lexeme.LexEventType{
					lexeme.MultiLineAnnotationBegin, lexeme.ObjectBegin, lexeme.ObjectEnd, lexeme.MultiLineAnnotationEnd,
					lexeme.LiteralBegin, lexeme.LiteralEnd,
				},
			},
			{
				`/* {} */ "a few multi-line comments in a single string" /* {} */`,
				[]lexeme.LexEventType{
					lexeme.MultiLineAnnotationBegin, lexeme.ObjectBegin, lexeme.ObjectEnd, lexeme.MultiLineAnnotationEnd,
					lexeme.LiteralBegin, lexeme.LiteralEnd,
					lexeme.MultiLineAnnotationBegin, lexeme.ObjectBegin, lexeme.ObjectEnd, lexeme.MultiLineAnnotationEnd,
				},
			},
		}

		for _, tst := range cc {
			t.Run(tst.content, func(t *testing.T) {
				file := new(fs.File)
				file.SetContent(bytes.Bytes(tst.content))

				s := newScanner(file)
				processingValid(t, s, tst)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := []string{
			`+1`,
			`zzz`,
			`tRue`,
			`trUe`,
			`truE`,
			`tru`,
			`fAlse`,
			`faLse`,
			`falSe`,
			`falsE`,
			`fal`,
			`nUll`,
			`nuLl`,
			`nulL`,
			`nul`,
			`"	"`,
			`"\x"`,
			`"\uZ"`,
			`"\u1Z"`,
			`"\u22Z"`,
			`"\u33Z"`,
			`"\u444Z"`,
			`-z`,
			`5.1.2`,
			`2"`,
			`2'`,
			`0.z`,
			`1.23e+Z`,
			`[}`,
			`[1,]`,
			`[1:]`,
			`{}x`,
			`{"key"}`,
			`{"key":1:}`,
			`{"key": 1,:`,
			`{`,
			`[`,
			`"string without closing quotation mark`,
			`string without opening quotation mark"`,
			"123\n2",
			"{}-",
			`"str" / comment`,
			`123 // {min: 1} text`,
			`123 // {min: 1} {}`,
			`123 // {`,
			`1 // {key: "aaa
bbb"}`,
			`1 // {
		key: 123
	}`,
			`/* /* {} */ */ 123`,
		}

		for _, content := range cc {
			t.Run(content, func(t *testing.T) {
				assert.Panics(t, func() {
					s := newScanner(fs.NewFile("", bytes.Bytes(content)))
					for {
						if _, ok := s.Next(); ok == false {
							break
						}
					}
				})
			})
		}
	})
}

func TestScanner_isSpace(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := []byte{
			' ',
			'\t',
			'\n',
			'\r',
		}

		for _, c := range cc {
			t.Run(string(c), func(t *testing.T) {
				assert.True(t, (&Scanner{}).isSpace(c))
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := make([]byte, 0, 255)

		for i := 0; i <= 255; i++ {
			if i == ' ' || i == '\t' || i == '\n' || i == '\r' {
				continue
			}
			cc = append(cc, byte(i))
		}

		for _, c := range cc {
			t.Run(string(c), func(t *testing.T) {
				assert.False(t, (&Scanner{}).isSpace(c))
			})
		}
	})
}

func TestScanner_isAnnotationStart(t *testing.T) {
	cc := map[byte]bool{
		'/': true,
	}
	for i := 0; i <= 255; i++ {
		if i != '/' {
			cc[byte(i)] = false
		}
	}

	s := &Scanner{}
	for c, expected := range cc {
		t.Run(string(c), func(t *testing.T) {
			assert.Equal(t, expected, s.isAnnotationStart(c))
		})
	}
}

func Test_stateFoundRootValue(t *testing.T) {
	cc := map[byte]struct {
		expectedState state
		expectedFinds []lexeme.LexEventType
	}{
		'/': {scanContinue, []lexeme.LexEventType{}},
		'#': {scanContinue, []lexeme.LexEventType{}},
		'{': {scanBeginObject, []lexeme.LexEventType{lexeme.ObjectBegin}},
		'[': {scanBeginArray, []lexeme.LexEventType{lexeme.ArrayBegin}},
		'@': {scanBeginTypesShortcut, []lexeme.LexEventType{
			lexeme.MixedValueBegin,
			lexeme.TypesShortcutBegin,
		}},
	}

	for b, c := range cc {
		t.Run(string(b), func(t *testing.T) {
			f := &fs.File{}
			s := newScanner(f)

			st := stateFoundRootValue(s, b)

			assert.Equal(t, c.expectedState, st)
			assert.Equal(t, c.expectedFinds, s.finds)
		})
	}
}

func Test_stateFoundObjectValueBegin(t *testing.T) {
	cc := map[string]struct {
		char              byte
		expectedState     state
		expectedLexEvents []lexeme.LexEventType
	}{
		"literal": {
			char:              '1',
			expectedState:     scanBeginLiteral,
			expectedLexEvents: []lexeme.LexEventType{lexeme.ObjectValueBegin, lexeme.LiteralBegin},
		},
		"object": {
			char:              '{',
			expectedState:     scanBeginObject,
			expectedLexEvents: []lexeme.LexEventType{lexeme.ObjectValueBegin, lexeme.ObjectBegin},
		},
		"array": {
			char:              '[',
			expectedState:     scanBeginArray,
			expectedLexEvents: []lexeme.LexEventType{lexeme.ObjectValueBegin, lexeme.ArrayBegin},
		},
		"types shortcut": {
			char:          '@',
			expectedState: scanBeginTypesShortcut,
			expectedLexEvents: []lexeme.LexEventType{
				lexeme.ObjectValueBegin,
				lexeme.MixedValueBegin,
				lexeme.TypesShortcutBegin,
			},
		},
	}

	for n, c := range cc {
		t.Run(n, func(t *testing.T) {
			s := newScanner(&fs.File{})
			st := stateFoundObjectValueBegin(s, c.char)
			assert.Equal(t, c.expectedState, st)
			assert.Len(t, s.finds, len(c.expectedLexEvents))
			assert.Equal(t, c.expectedLexEvents, s.finds)
		})
	}
}

func Test_stateFoundArrayItemBeginOrEmpty(t *testing.T) {
	cc := map[string]struct {
		char              byte
		expectedState     state
		expectedLexEvents []lexeme.LexEventType
	}{
		"literal": {
			char:              '1',
			expectedState:     scanBeginLiteral,
			expectedLexEvents: []lexeme.LexEventType{lexeme.ArrayItemBegin, lexeme.LiteralBegin},
		},
		"object": {
			char:              '{',
			expectedState:     scanBeginObject,
			expectedLexEvents: []lexeme.LexEventType{lexeme.ArrayItemBegin, lexeme.ObjectBegin},
		},
		"array": {
			char:              '[',
			expectedState:     scanBeginArray,
			expectedLexEvents: []lexeme.LexEventType{lexeme.ArrayItemBegin, lexeme.ArrayBegin},
		},
		"types shortcut": {
			char:          '@',
			expectedState: scanBeginTypesShortcut,
			expectedLexEvents: []lexeme.LexEventType{
				lexeme.ArrayItemBegin,
				lexeme.MixedValueBegin,
				lexeme.TypesShortcutBegin,
			},
		},
	}

	for n, c := range cc {
		t.Run(n, func(t *testing.T) {
			s := newScanner(&fs.File{})
			st := stateFoundArrayItemBeginOrEmpty(s, c.char)
			assert.Equal(t, c.expectedState, st)
			assert.Len(t, s.finds, len(c.expectedLexEvents))
			assert.Equal(t, c.expectedLexEvents, s.finds)
		})
	}
}

type validResults struct {
	content string
	results []lexeme.LexEventType
}

func processingValid(t *testing.T, s *Scanner, tst validResults) {
	defer func() {
		if r := recover(); r != nil {
			str := fmt.Sprint(r)
			t.Errorf("Panic at:\n%s\n\n%s", tst.content, str)
		}
	}()

	var results []lexeme.LexEventType

	for {
		if lex, ok := s.Next(); ok {
			results = append(results, lex.Type())
		} else {
			break
		}
	}

	assert.Equal(
		t,
		lexSliceToStringSlice(tst.results),
		lexSliceToStringSlice(results),
	)
}

func lexSliceToStringSlice(ll []lexeme.LexEventType) []string {
	ss := make([]string, 0, len(ll))
	for _, l := range ll {
		ss = append(ss, l.String())
	}
	return ss
}
