package parse

import (
	"bufio"
	"bytes"
	"testing"
)

var itemTests = []struct {
	in  item
	out []byte
}{
	{
		item{typ: itemString, val: []byte("OK")},
		[]byte("+OK\r\n"),
	},
	{
		item{typ: itemError, val: []byte("Error message")},
		[]byte("-Error message\r\n"),
	},
	{
		item{typ: itemInteger},
		[]byte(":0\r\n"),
	},
	{
		item{typ: itemInteger, i: 1000},
		[]byte(":1000\r\n"),
	},
	{
		item{typ: itemBulk, val: []byte("foobar")},
		[]byte("$6\r\nfoobar\r\n"),
	},
	{
		item{typ: itemBulk, val: []byte{}},
		[]byte("$0\r\n\r\n"),
	},
	{
		item{typ: itemBulk},
		[]byte("$-1\r\n"),
	},
	{
		item{typ: itemArray},
		[]byte("*0\r\n"),
	},
	{
		item{typ: itemArray, i: 2},
		[]byte("*2\r\n"),
	},
	{
		item{typ: itemArray, i: -1},
		[]byte("*-1\r\n"),
	},
	{
		item{typ: itemInteger, i: -9223372036854775807},
		[]byte(":-9223372036854775807\r\n"),
	},
	{
		item{typ: itemInteger, i: 9223372036854775807},
		[]byte(":9223372036854775807\r\n"),
	},
	{
		item{typ: itemInline, val: []byte("EXISTS somekey")},
		[]byte("EXISTS somekey\r\n"),
	},
	{
		item{typ: -999, val: []byte("junk")},
		nil,
	},
	{
		item{typ: -998, i: 10},
		nil,
	},
}

func TestItemBytes(t *testing.T) {
	for _, test := range itemTests {
		if test.in.String() == "" {
			t.Errorf("item.String() failed for %#v", test.in)
		}
		out := test.in.bytes()
		if !bytes.Equal(out, test.out) {
			t.Errorf("item %s got %q, expected %q", test.in, string(out), string(test.out))
		}
	}
}

func errorOrNil(err error) string {
	if err == nil {
		return "nil"
	}
	return err.Error()
}

var invalidSplit = [][]byte{
	{':', '\n'},
	{':', 0, '\n'},
	{'$', '-', '3', '\n'},
	{'$', '1', '\n', 0, 0, 0, 0},
}

func TestSplit(t *testing.T) {
	advance, token, err := split([]byte{}, false)
	if advance != 0 || token != nil || err != nil {
		t.Errorf("unexpected return from empty non-EOF split: %d, %q, %s", advance, string(token), errorOrNil(err))
	}
	advance, token, err = split([]byte{}, true)
	if advance != 0 || token != nil || err != nil {
		t.Errorf("unexpected return from empty at-EOF split: %d, %q, %s", advance, string(token), errorOrNil(err))
	}
	for _, data := range invalidSplit {
		advance, token, err = split(data, true)
		if advance != 0 || token != nil || err != errInvalid {
			t.Errorf("unexpected return from invalid split %q: %d, %q, %s", string(data), advance, string(token), errorOrNil(err))
		}
	}
	advance, token, err = split([]byte{'$', '0', '\n', '\n'}, true)
	if err != nil {
		t.Errorf("unexpected error for split zero bulk: %s", errorOrNil(err))
	}
	if advance != 4 || !bytes.Equal(token, []byte{'$', '0', '\n'}) || err != nil {
		t.Errorf("unexpected return from zero bulk: %d, %v, %s", advance, token, errorOrNil(err))
	}
	for _, test := range itemTests {
		if test.out == nil {
			continue
		}
		for n := range test.out[:len(test.out)-1] {
			advance, token, err = split(test.out[:n+1], true)
			if err == nil {
				t.Errorf("nil error for incomplete split %q", string(test.out[:n+1]))
			}
			if advance != 0 || token != nil {
				t.Errorf("unexpected return from incomplete split %q: %d %q", string(test.out[:n+1]), advance, string(token))
			}
			advance, token, err = split(test.out[:n+1], false)
			if err != nil {
				t.Error("error returned from split: ", err)
			}
			if advance != 0 || token != nil {
				t.Errorf("unexpected return from incomplete split %q: %d %q", string(test.out[:n+1]), advance, string(token))
			}
		}
	}
}

func TestScanner(t *testing.T) {
	var b []byte
	var added, scanned int
	for x := 0; x < 1000; x++ {
		for _, test := range itemTests {
			if test.out == nil {
				continue
			}
			b = append(b, test.out...)
			added++
		}
	}
	r := bytes.NewReader(b)
	s := bufio.NewScanner(r)
	s.Split(split)
	for x := 0; x < 1000; x++ {
		for _, test := range itemTests {
			if test.out == nil {
				continue
			}
			if !s.Scan() {
				break
			}
			token := append(s.Bytes(), '\r', '\n')
			if !bytes.Equal(token, test.out) {
				t.Errorf("expected %q, got %q", string(test.out), string(token))
			}
			scanned++
		}
	}
	if s.Err() != nil {
		t.Fatalf("error scanning: %s", s.Err().Error())
	}
	if added != scanned {
		t.Fatalf("scanned %d, expected %d", scanned, added)
	}
}
