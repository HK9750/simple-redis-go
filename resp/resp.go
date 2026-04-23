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

type Resp struct {
	reader *bufio.Reader
}

func NewResp(rd *bufio.Reader) *Resp {
	return &Resp{reader: rd}
}

func (r *Resp) Read() (Value, error) {
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

func (r *Resp) readLine() (string, error) {
	line, err := r.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(line) < 2 || line[len(line)-2] != '\r' {
		return "", fmt.Errorf("invalid RESP line ending")
	}
	return line[:len(line)-2], nil
}

func (r *Resp) readString() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{Type: "string", Str: line}, nil
}

func (r *Resp) readError() (Value, error) {
	line, err := r.readLine()
	if err != nil {
		return Value{}, err
	}
	return Value{Type: "error", Str: line}, nil
}

func (r *Resp) readInteger() (Value, error) {
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

func (r *Resp) readBulk() (Value, error) {
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

func (r *Resp) readArray() (Value, error) {
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
func (r *Resp) expectCRLF() error {
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
