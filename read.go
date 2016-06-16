// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// ini包提供了对ini的基本操作。
package ini

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// 表示ini的语法错误信息。
type SyntaxError struct {
	Line int
	Msg  string
}

func (s *SyntaxError) Error() string {
	return fmt.Sprintf("encoding/ini，在第%d行发生语法错误：%v", s.Line, s.Msg)
}

// ini节点元素类型
const (
	Undefined = iota // 未定义，初始状态
	Element
	Section
	Comment
	EOF // 已经读取完毕。
)

// Token用于描述每一个节点的类型信息及数据内容。
type Token struct {
	Type  int    // 类型，可以是上面的任意节点类型
	Key   string // 该节点的键名，仅在Type值为Element时才有效
	Value string // 该节点对应的值
}

func (t *Token) reset() {
	t.Type = Undefined
	t.Value = t.Value[:0]
	t.Key = t.Key[:0]
}

// 复制一个新的Token
func (t *Token) Copy() *Token {
	return &Token{
		Type:  t.Type,
		Value: t.Value,
		Key:   t.Key,
	}
}

// ini数据的读取操作类。
// 注释只支持以`#`,`;`开头的行，不支持行尾注释；
//
// 对于空格的处理:
// - section:去掉首尾空格。
// - comment:去掉尾部空格。
// - element:去掉key和value的首尾空格
type Reader struct {
	reader *bufio.Reader
	atEOF  bool // 已经读取完毕
	line   int  // 当前正在处理的行数。
	token  *Token
}

// 从一个io.Reader初始化Reader
func NewReader(r io.Reader) *Reader {
	return &Reader{reader: bufio.NewReader(r), token: &Token{}}
}

// 从一个[]byte初始化Reader
func NewReaderBytes(data []byte) *Reader {
	return NewReader(bytes.NewReader(data))
}

// 从一个字符串初始化Reader
func NewReaderString(str string) *Reader {
	return NewReader(strings.NewReader(str))
}

// 返回下一个Token，当内容读取完毕之后，将返回Type值为EOF的Token。
// 返回的Token.Value都将不包含尾部的空格（包括换行符）。
//
// 返回的Token变量，在下次调用Reader.Token()方法时，数据会被重置，
// 若需要保存Token的数据，可使用Token.Copy()函数复制一份。
func (r *Reader) Token() (*Token, error) {
	r.token.reset()

START:
	if r.atEOF {
		r.token.Type = EOF
		return r.token, nil
	}

	buffer, err := r.reader.ReadString('\n')
	r.line++
	if err != nil {
		if err != io.EOF { // 真的发生错误了
			return nil, err
		}

		// 读取完毕
		r.atEOF = true
		if len(buffer) == 0 { // 读取完毕，且当前这次也没有新内容
			r.token.Type = EOF
			return r.token, nil
		}
	}

	buffer = strings.TrimSpace(buffer)
	if len(buffer) == 0 { // 空行
		goto START
	}

	return r.parseLine(buffer)
}

// 将一行字符串转换成对应的Token实例。
// 返回的Token.Value都将不包含尾部的空格。
func (r *Reader) parseLine(line string) (*Token, error) {
	switch line[0] {
	case '[': // section
		if line[len(line)-1] != ']' {
			return nil, r.newSyntaxError("parseLine:section名称没有以]作为结尾")
		}

		r.token.Value = strings.TrimSpace(line[1 : len(line)-1])
		if len(r.token.Value) == 0 {
			return nil, r.newSyntaxError("parseLine:section名称不能为空字符串")
		}
		r.token.Type = Section
		return r.token, nil
	case '#', ';': // comment
		r.token.Type = Comment
		r.token.Value = line[1:]
		return r.token, nil
	default: // element
		pos := strings.IndexRune(line, '=')
		if pos < 0 {
			return nil, r.newSyntaxError("parseLine:表达式中未找到`=`符号")
		}
		if pos == 0 { // 键名不能为空，键值不能为空
			return nil, r.newSyntaxError("parseLine:键名不能为空")
		}

		r.token.Type = Element
		r.token.Key = strings.TrimRightFunc(line[:pos], unicode.IsSpace)
		r.token.Value = strings.TrimLeftFunc(line[pos+1:], unicode.IsSpace)
		return r.token, nil
	}

	r.token.Type = EOF
	return r.token, nil
}

// 构造一个SyntaxError实例。
func (r *Reader) newSyntaxError(msg string) error {
	return &SyntaxError{
		Msg:  msg,
		Line: r.line,
	}
}

// 将ini转换成map[string]map[string]string格式的数据。其内容表示如下：
//  map[string]map[string]string{
//      "" : map[string]string{"k1":"v1", "K2":"v2"},
//      "section1" : map[string]string{"k1":"v1", "k2":"v2"},
//  }
// 索引值为空的map表示的是非section下的键值对。
//
// 没有与之相对就的MarshalMap，因为map是无序的，若一个map带了section，
// 则转换结果未必是正确的。
func UnmarshalMap(data []byte) (map[string]map[string]string, error) {
	if len(data) == 0 {
		return nil, &SyntaxError{Msg: "UnmarshalMap:没有内容", Line: 0}
	}

	m := make(map[string]map[string]string)
	currSection := map[string]string{}
	sectionName := ""

	r := NewReaderBytes(data)
LOOP:
	for {
		token, err := r.Token()
		if err != nil {
			return nil, err
		}

		switch token.Type {
		case Comment:
			continue
		case EOF:
			break LOOP
		case Element:
			currSection[token.Key] = token.Value
		case Section:
			m[sectionName] = currSection

			currSection = map[string]string{}
			sectionName = token.Value
		default:
			return nil, errors.New("UnmarshalMap:未知的元素类型")
		}
	} // end for
	m[sectionName] = currSection

	return m, nil
}
