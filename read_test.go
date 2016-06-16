// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ini

import (
	"strings"
	"testing"
	"unicode"

	"github.com/issue9/assert"
)

func TestToken(t *testing.T) {
	a := assert.New(t)

	token := &Token{
		Type:  Element,
		Key:   "key",
		Value: "value",
	}

	t1 := token.Copy()
	a.Equal(t1, token)

	token.reset()
	a.Equal(token.Type, Undefined).Equal(t1.Type, Element)
	a.Equal(token.Key, "").Equal(t1.Key, "key")
	a.Equal(token.Value, "").Equal(t1.Value, "value")
}

func TestReader_ParseLine(t *testing.T) {
	a := assert.New(t)

	type test struct {
		value   string
		isError bool
		token   *Token
	}

	// parseLine只处理首尾没有空格的字符串。
	data := []*test{
		&test{value: "[section]", isError: false, token: &Token{Type: Section, Value: "section"}},
		&test{value: "[ section ]", isError: false, token: &Token{Type: Section, Value: "section"}},
		&test{value: "[ s ection ]", isError: false, token: &Token{Type: Section, Value: "s ection"}},
		&test{value: "key=val", isError: false, token: &Token{Type: Element, Key: "key", Value: "val"}},
		&test{value: "key = val", isError: false, token: &Token{Type: Element, Key: "key", Value: "val"}},
		&test{value: "k ey = val", isError: false, token: &Token{Type: Element, Key: "k ey", Value: "val"}},
		&test{value: "key = v al", isError: false, token: &Token{Type: Element, Key: "key", Value: "v al"}},
		&test{value: "key = v=al", isError: false, token: &Token{Type: Element, Key: "key", Value: "v=al"}},
		&test{value: "key =", isError: false, token: &Token{Type: Element, Key: "key", Value: ""}},

		// 各类错误格式
		&test{value: "[section", isError: true},
		&test{value: "key val", isError: true},
		&test{value: "[]", isError: true},
		&test{value: "=i", isError: true},
	}

	r := NewReader(nil)
	for index, item := range data {
		token, err := r.parseLine(item.value)
		if item.isError {
			a.Error(err)
			continue
		}

		a.NotError(err)
		a.Equal(item.token.Type, token.Type, "在测试%d行数据的Type时出错:v1=[%v],v2=[%v]", index, item.token.Type, token.Type).
			Equal(item.token.Key, token.Key, "在测试%d行数据的Key时出错:v1=[%v],v2=[%v]", index, item.token.Key, token.Key).
			Equal(item.token.Value, token.Value, "在测试%d行数据的Value时出错:v1=[%v],v2=[%v]", index, item.token.Value, token.Value)
	}
}

// 测试数据无法与writer同用，比如涉及到带换行符的comment等内容。
var readTestData = []*tester{
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
}

func TestReader(t *testing.T) {
	a := assert.New(t)

	for index, test := range readTestData {
		r := NewReaderString(test.value)
		count := 0 // 对应test.Tokens的索引值
		for {
			token, err := r.Token()
			a.NotError(err)
			if token.Type == EOF {
				break
			}

			t1 := test.tokens[count]
			a.Equal(token.Type, t1.Type, "第%d条测试数据，Type不相等。v1=[%v],v2=[%v]", index, token.Type, t1.Type).
				Equal(token.Key, t1.Key, "第%d条测试数据，Key不相等。v1=[%v],v2=[%v]", index, token.Key, t1.Key).
				Equal(token.Value, strings.TrimRightFunc(t1.Value, unicode.IsSpace), "第%d条测试数据，Value不相等。v1=[%v],v2=[%v]", index, token.Value, t1.Value)
			count++
		}
	}
}

func TestUnmarshalMap(t *testing.T) {
	a := assert.New(t)

	// 传递空的字符串，将返回错误信息。
	m, err := UnmarshalMap([]byte(""))
	a.Error(err).Nil(m)

	// 带section
	str := []byte(`
    nosectionkey=nosectionval
    [section]
    skey=sval
    [section1]
	;comment
	#comment
    key2=val2
    `)
	v1 := map[string]map[string]string{
		"": map[string]string{"nosectionkey": "nosectionval"},
		"section": map[string]string{
			"skey": "sval",
		},
		"section1": map[string]string{
			"key2": "val2",
		},
	}
	m, err = UnmarshalMap(str)
	a.NotError(err)
	a.Equal(m, v1)

	// 不带section
	str = []byte(`
    nosectionkey=nosectionval
    `)
	v1 = map[string]map[string]string{
		"": map[string]string{"nosectionkey": "nosectionval"},
	}
	m, err = UnmarshalMap(str)
	a.NotError(err)
	a.Equal(m, v1)

	// 只有section
	str = []byte(`
	[section]
    nosectionkey=nosectionval
    `)
	v1 = map[string]map[string]string{
		"":        map[string]string{},
		"section": map[string]string{"nosectionkey": "nosectionval"},
	}
	m, err = UnmarshalMap(str)
	a.NotError(err)
	a.Equal(m, v1)
}
