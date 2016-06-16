// Copyright 2014 by caixw, All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package ini

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// 用于输出ini内容到指定的io.Writer。
//
// 内容并不是实时写入io.Writer的，
// 需要调用Writer.Flush()才会真正地写入到io.Writer流中。
// 对于重复的键名和section名称并不会报错，若需要唯一值，
// 需要用户自行解决。
type Writer struct {
	buf    *bufio.Writer
	symbol byte
}

// 声明一个新的Writer实例。
//
// w写入的io.Writer接口；
// commentSymbol注释符号。只能是'#'或';'，传递其它参数将返回错误信息。
func NewWriter(w io.Writer, commentSymbol byte) (*Writer, error) {
	if commentSymbol != '#' && commentSymbol != ';' {
		return nil, errors.New("NewWriter:注释符号只能是`;`或`#`")
	}

	return &Writer{
		buf:    bufio.NewWriter(w),
		symbol: commentSymbol,
	}, nil
}

// 添加一个新的空行。
func (w *Writer) NewLine() error {
	return w.buf.WriteByte('\n')
}

// 添加section，section没有嵌套功能，添加一个新的Section，意味着前一个section的结束。
// section名称只能在同一行，若section值中包含换行符，则会返回错误信息。
func (w *Writer) AddSection(section string) (err error) {
	if len(section) == 0 {
		return errors.New("AddSection:section名称不能为空值")
	}

	if strings.IndexByte(section, '\n') > -1 {
		return errors.New("AddSection:section名称中不能包含换行符")
	}

	if err = w.buf.WriteByte('['); err != nil {
		return err
	}

	if _, err = w.buf.WriteString(section); err != nil {
		return err
	}

	_, err = w.buf.WriteString("]\n")
	return err
}

// 添加一个键值对。
func (w *Writer) AddElement(key, val string) (err error) {
	if len(key) == 0 { // val可以为空，key不能为空
		return errors.New("AddElement:参数key不能为空")
	}

	if strings.IndexByte(key, '\n') > -1 || strings.IndexByte(val, '\n') > -1 {
		return errors.New("AddElement:参数key和val都不能包含换行符")
	}

	if _, err = w.buf.WriteString(key); err != nil {
		return err
	}

	if err = w.buf.WriteByte('='); err != nil {
		return err
	}

	if _, err = w.buf.WriteString(val); err != nil {
		return err
	}

	return w.NewLine()
}

// 添加一个键值对。val使用fmt.Sprint格式化成字符串。
func (w *Writer) AddElementf(key string, val interface{}) error {
	return w.AddElement(key, fmt.Sprint(val))
}

// 添加注释。能正确识别换行符。
// 若需要添加一个不带注释符号的空，请使用NewLine()方法。
//
// 对于换行符的处理：简单地将\n符号去你的成\n#，
// 所以若传递一个仅有\n的字符串，最终将输出2行空注释。
func (w *Writer) AddComment(comment string) (err error) {
	if strings.IndexByte(comment, '\n') > -1 { // 存在换行符
		comment = strings.Replace(comment, "\n", "\n"+string(w.symbol), -1)
	}

	if err = w.buf.WriteByte(w.symbol); err != nil {
		return err
	}

	if _, err = w.buf.WriteString(comment); err != nil {
		return err
	}

	return w.NewLine()
}

// 将内容输出到io.Writer中
func (w *Writer) Flush() {
	w.buf.Flush()
}
