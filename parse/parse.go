package parse

import "fmt"

type item struct {
	typ itemType
	val []byte // value for inline, string, error and bulk
	i   int64  // value for integer and array
}

type itemType int

const (
	itemInline itemType = iota
	itemString
	itemError
	itemInteger
	itemBulk
	itemArray
)

// bytes returns a byte slice owned by caller (a copy of item contents)
func (i item) bytes() (b []byte) {
	switch i.typ {
	case itemInline:
		b = make([]byte, len(i.val)+2) // val + CRLF
		copy(b, i.val)
	case itemString:
		b = make([]byte, 1+len(i.val)+2) // '+' + val + CRLF
		b[0] = '+'
		copy(b[1:], i.val)
	case itemError:
		b = make([]byte, 1+len(i.val)+2) // '-' + val + CRLF
		b[0] = '-'
		copy(b[1:], i.val)
	case itemInteger:
		b = make([]byte, 1+20+2) // ':' + maxintlen + CRLF
		b[0] = ':'
		l := itoa(b[1:], i.i)
		b = b[:1+l+2] // ':' + intlen + CRLF
	case itemBulk:
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
	case itemArray:
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

func (i item) String() string {
	switch i.typ {
	case itemInline:
		return fmt.Sprintf("inline(%q)", string(i.val))
	case itemString:
		return fmt.Sprintf("string(%q)", string(i.val))
	case itemError:
		return fmt.Sprintf("error(%q)", string(i.val))
	case itemInteger:
		return fmt.Sprintf("integer(%d)", i.i)
	case itemBulk:
		if i.val == nil {
			return "bulk(nil)"
		}
		return fmt.Sprintf("bulk(%q)", string(i.val))
	case itemArray:
		return fmt.Sprintf("array(%d)", i.i)
	}
	if i.val == nil {
		return fmt.Sprintf("unknown(nil, %d)", i.i)
	}
	return fmt.Sprintf("unknown(%q, %d)", string(i.val), i.i)
}
