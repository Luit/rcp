package parse // import "luit.eu/rcp/parse"

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
)

var (
	errInvalid    = errors.New("rcp/parse: invalid RESP line")
	errIncomplete = errors.New("rcp/parse: incomplete RESP line")
)

type Item struct {
	typ ItemType
	val []byte // value for inline, string, error and bulk
	i   int64  // value for integer and array
}

type ItemType int

const (
	ItemInline ItemType = iota
	ItemString
	ItemError
	ItemInteger
	ItemBulk
	ItemArray
)

// bytes returns a byte slice owned by caller (a copy of Item contents)
func (i Item) bytes() (b []byte) {
	switch i.typ {
	case ItemInline:
		b = make([]byte, len(i.val)+2) // val + CRLF
		copy(b, i.val)
	case ItemString:
		b = make([]byte, 1+len(i.val)+2) // '+' + val + CRLF
		b[0] = '+'
		copy(b[1:], i.val)
	case ItemError:
		b = make([]byte, 1+len(i.val)+2) // '-' + val + CRLF
		b[0] = '-'
		copy(b[1:], i.val)
	case ItemInteger:
		b = make([]byte, 1+20+2) // ':' + maxintlen + CRLF
		b[0] = ':'
		l := itoa(b[1:], i.i)
		b = b[:1+l+2] // ':' + intlen + CRLF
	case ItemBulk:
		if i.val == nil {
			b = make([]byte, 1+2+2) // '$' + '-1' + CRLF
			b[0], b[1], b[2] = '$', '-', '1'
			break
		}
		b = make([]byte, 1+20+2+len(i.val)+2) // '$' + maxintlen + CRLF + val + CRLF
		b[0] = '$'
		l := itoa(b[1:], int64(len(i.val)))
		b[1+l], b[1+l+1] = '\r', '\n'
		b = b[:1+l+2+len(i.val)+2] // '$' + lenlen + CRLF + val + CRLF
		copy(b[1+l+2:], i.val)
	case ItemArray:
		b = make([]byte, 1+20+2) // '*' + maxintlen + CRLF
		b[0] = '*'
		l := itoa(b[1:], i.i)
		b = b[:1+l+2] // '*' + intlen + CRLF
	default:
		return nil
	}
	b[len(b)-2], b[len(b)-1] = '\r', '\n'
	return
}

func (i Item) String() string {
	switch i.typ {
	case ItemInline:
		return fmt.Sprintf("inline(%q)", string(i.val))
	case ItemString:
		return fmt.Sprintf("string(%q)", string(i.val))
	case ItemError:
		return fmt.Sprintf("error(%q)", string(i.val))
	case ItemInteger:
		return fmt.Sprintf("integer(%d)", i.i)
	case ItemBulk:
		if i.val == nil {
			return "bulk(nil)"
		}
		return fmt.Sprintf("bulk(%q)", string(i.val))
	case ItemArray:
		return fmt.Sprintf("array(%d)", i.i)
	}
	if i.val == nil {
		return fmt.Sprintf("unknown(nil, %d)", i.i)
	}
	return fmt.Sprintf("unknown(%q, %d)", string(i.val), i.i)
}

// split is a bufio.SplitFunc used to split and validate data from an
// io.Reader reading RESP lines.
func Split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	i := bytes.IndexByte(data, '\n')
	if i < 0 {
		if atEOF {
			return 0, nil, errIncomplete
		}
		// Request more data.
		return 0, nil, nil
	}
	j := i
	if j > 0 && data[j-1] == '\r' {
		j--
	}
	switch data[0] {
	case ':', '$', '*':
	default:
		// We have a full newline-terminated line.
		return i + 1, data[0:j], nil
	}
	var n int64
	n, err = atoi(data[1:j])
	if err != nil {
		return 0, nil, errInvalid
	}
	if data[0] != '$' {
		return i + 1, data[0:j], nil
	}
	// TODO: make sure n fits in plain int?
	m := int(n)
	switch {
	case m < -1:
		return 0, nil, errInvalid
	case m == -1:
		return i + 1, data[0 : i+1], nil
	case len(data) < i+1+m+1, len(data) < i+1+m+2 && data[i+1+m] == '\r':
		if atEOF {
			return 0, nil, errIncomplete
		}
		// Request more data
		return 0, nil, nil
	case data[i+1+m] == '\n':
		return i + 1 + m + 1, data[0 : i+1+m], nil
	case data[i+1+m] == '\r' && data[i+1+m+1] == '\n':
		return i + 1 + m + 2, data[0 : i+1+m], nil
	default:
		return 0, nil, errInvalid
	}
}

type Reader interface {
	Read() (Item, error)
}

type reader struct {
	s *bufio.Scanner
}

func NewReader(r io.Reader) Reader {
	rr := &reader{s: bufio.NewScanner(r)}
	rr.s.Split(Split)
	return rr
}

func (r *reader) Read() (i Item, err error) {
	if !r.s.Scan() {
		if r.s.Err() == nil {
			return i, errors.New("rcp/parse: unable to scan")
		}
		return i, r.s.Err()
	}
	b := r.s.Bytes()
	if len(b) < 1 {
		return i, errors.New("rcp/parse: empty scan")
	}
	switch b[0] {
	case '+':
		i.typ = ItemString
	case '-':
		i.typ = ItemError
	case ':':
		i.typ = ItemInteger
	case '$':
		i.typ = ItemBulk
	case '*':
		i.typ = ItemArray
	default:
		i.typ = ItemInline
	}
	if i.typ != ItemInline {
		b = b[1:]
	}
	end := len(b)
	switch i.typ {
	case ItemString, ItemError, ItemInline:
		i.val = make([]byte, end)
		copy(i.val, b)
		return
	case ItemBulk:
		end = bytes.IndexByte(b, '\n')
		if end < 0 {
			return i, errInvalid
		}
		if b[end-1] == '\r' {
			end--
		}
	}
	i.i, err = atoi(b[:end])
	if err != nil {
		return
	}
	if i.typ == ItemBulk {
		if i.i == -1 {
			return
		}
		end = bytes.IndexByte(b, '\n') + 1
		if int64(len(b)) < int64(end)+i.i {
			return i, errIncomplete
		}
		i.val = make([]byte, i.i)
		i.val = b[end : end+int(i.i)]
	}
	return
}
