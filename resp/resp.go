package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

const (
	STRING  = '+'
	ERROR   = '-'
	INTEGER = ':'
	BULK    = '$'
	ARRAY   = '*'
)

type Value struct {
	Type  string
	Str   string
	Num   int64
	Bulk  []byte
	Array []Value
	Null  bool
}

type Reader struct {
	reader *bufio.Reader
}

func NewReader(rd *bufio.Reader) *Reader {
	return &Reader{reader: rd}
}

type Writer struct {
	writer *bufio.Writer
}

func NewWriter(writer *bufio.Writer) *Writer {
	return &Writer{writer: writer}
}

func (r *Reader) Read() (Value, error) {
	t, err := r.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch t {
	case STRING:
		return r.readString()
	case ERROR:
		return r.readError()
	case INTEGER:
		return r.readInteger()
	case BULK:
		return r.readBulk()
	case ARRAY:
		return r.readArray()
	default:
		return Value{}, fmt.Errorf("unknown RESP type: %q", t)
	}
}

func (r *Reader) readLine() (string, error) {
	line, err := r.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(line) < 2 || line[len(line)-2] != '\r' {
		return "", fmt.Errorf("invalid RESP line ending")
	}
	return line[:len(line)-2], nil
}

func (r *Reader) readString() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{Type: "string", Str: line}, nil
}

func (r *Reader) readError() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{Type: "error", Str: line}, nil
}

func (r *Reader) readInteger() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}

	num, err := strconv.ParseInt(line, 10, 64)
	if err != nil {
		return Value{}, fmt.Errorf("invalid integer %q: %w", line, err)
	}

	return Value{Type: "integer", Num: num}, nil
}

func (r *Reader) readBulk() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}

	n, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, fmt.Errorf("invalid bulk length %q: %w", line, err)
	}
	if n < -1 {
		return Value{}, fmt.Errorf("invalid bulk length: %d", n)
	}
	if n == -1 {
		return Value{Type: "bulk", Null: true}, nil
	}

	data := make([]byte, n)
	if _, err := io.ReadFull(r.reader, data); err != nil {
		return Value{}, err
	}

	if err := r.expectCRLF(); err != nil {
		return Value{}, err
	}

	return Value{Type: "bulk", Bulk: data}, nil
}

func (r *Reader) readArray() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}

	n, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, fmt.Errorf("invalid array length %q: %w", line, err)
	}
	if n < -1 {
		return Value{}, fmt.Errorf("invalid array length: %d", n)
	}
	if n == -1 {
		return Value{Type: "array", Null: true}, nil
	}

	array := make([]Value, n)
	for i := range n {
		val, err := r.Read()
		if err != nil {
			return Value{}, err
		}
		array[i] = val
	}

	return Value{Type: "array", Array: array}, nil
}

// CRLF is the carriage return line feed sequence of \r\n
func (r *Reader) expectCRLF() error {
	b1, err := r.reader.ReadByte()
	if err != nil {
		return err
	}
	b2, err := r.reader.ReadByte()
	if err != nil {
		return err
	}
	if b1 != '\r' || b2 != '\n' {
		return fmt.Errorf("expected CRLF")
	}
	return nil
}

func (w *Writer) Write(v Value) error {
	bytes := v.Marshal()
	_, err := w.writer.Write(bytes)
	return err
}

func (v Value) Marshal() []byte {
	switch v.Type {
	case "array":
		return v.marshalArray()
	case "bulk":
		return v.marshalBulk()
	case "string":
		return v.marshalString()
	case "null":
		return v.marshallNull()
	case "error":
		return v.marshallError()
	default:
		return []byte{}
	}
}

func (v Value) marshalString() []byte {
	var bytes []byte
	bytes = append(bytes, STRING)
	bytes = append(bytes, v.Str...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalBulk() []byte {
	var bytes []byte
	bytes = append(bytes, BULK)
	bytes = append(bytes, strconv.Itoa(len(v.Bulk))...)
	bytes = append(bytes, '\r', '\n')
	bytes = append(bytes, v.Bulk...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshalArray() []byte {
	len := len(v.Array)
	var bytes []byte
	bytes = append(bytes, ARRAY)
	bytes = append(bytes, strconv.Itoa(len)...)
	bytes = append(bytes, '\r', '\n')

	for i := 0; i < len; i++ {
		bytes = append(bytes, v.Array[i].Marshal()...)
	}

	return bytes
}

func (v Value) marshallError() []byte {
	var bytes []byte
	bytes = append(bytes, ERROR)
	bytes = append(bytes, []byte(v.Str)...)
	bytes = append(bytes, '\r', '\n')

	return bytes
}

func (v Value) marshallNull() []byte {
	return []byte("$-1\r\n")
}
