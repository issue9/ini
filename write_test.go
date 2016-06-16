// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ini

import (
	"bytes"
	"testing"

	"github.com/issue9/assert"
)

// 用于Reader和Writer的测试用例
type tester struct {
	tokens []*Token
	value  string
}

var testData = []*tester{
	// 比较正规的写法
	&tester{
		tokens: []*Token{
			&Token{Type: Comment, Value: "comment"},
			&Token{Type: Element, Value: "Value1", Key: "Key1"},
			&Token{Type: Element, Value: "Value2", Key: "Key2"},
			&Token{Type: Section, Value: "section1"},
			&Token{Type: Element, Value: "Value1", Key: "Key1"},
		},
		value: `#comment
Key1=Value1
Key2=Value2
[section1]
Key1=Value1
`,
	},

	// 多行注释
	&tester{
		tokens: []*Token{
			&Token{Type: Comment, Value: "comment line 1"},
			&Token{Type: Comment, Value: "comment line 2"},
			&Token{Type: Comment, Value: ""}, // 空注释行
			&Token{Type: Element, Value: "value", Key: "key"},
		},
		value: `#comment line 1
#comment line 2
#
key=value
`,
	},

	// 各类特殊字符乱入
	&tester{
		tokens: []*Token{
			&Token{Type: Comment, Value: "comme#nt"},
			&Token{Type: Element, Value: "=Value1", Key: "Key1"},
			&Token{Type: Element, Value: "Value2=", Key: "Key2"},
			&Token{Type: Section, Value: "sect=ion1"},
			&Token{Type: Element, Value: "Val#ue1", Key: "Ke;y1"},
		},
		value: `#comme#nt
Key1==Value1
Key2=Value2=
[sect=ion1]
Ke;y1=Val#ue1
`,
	},

	// 带转义字符和注释行空格
	&tester{
		tokens: []*Token{
			&Token{Type: Comment, Value: "comment 1\n\n"},
			&Token{Type: Element, Value: "value", Key: "key"},
			&Token{Type: Comment, Value: "\n comment 3"},
			&Token{Type: Element, Value: "value", Key: "key"},
			&Token{Type: Comment, Value: "\n"},
		},
		value: `#comment 1
#
#
key=value
#
# comment 3
key=value
#
#
`,
	},
}

func TestWrite(t *testing.T) {
	a := assert.New(t)
	buf := new(bytes.Buffer)

	for index, test := range testData {
		buf.Reset()
		w, err := NewWriter(buf, '#')
		a.NotError(err).NotNil(w)
		for _, token := range test.tokens {
			switch token.Type {
			case Comment:
				w.AddComment(token.Value)
			case Element:
				w.AddElementf(token.Key, token.Value)
			case EOF:
				break
			case Section:
				w.AddSection(token.Value)
			case Undefined:
				t.Errorf("在第[%v]个测试数据中检测到Type值为Undefined的Token", index)
			default:
				t.Errorf("在第[%v]个测试数据中检测到Type值[%v]为未定义的Token", index, token.Type)
			}
		} // end for test.tokens
		w.Flush()
		a.Equal(buf.String(), test.value)
	}
}
