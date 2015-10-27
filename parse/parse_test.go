package parse

import (
	"bytes"
	"testing"
)

var itemBytesTests = []struct {
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
}

func TestItemBytes(t *testing.T) {
	for _, test := range itemBytesTests {
		out := test.in.bytes()
		if bytes.Compare(out, test.out) != 0 {
			t.Errorf("item %s got %q, expected %q", test.in, string(out), string(test.out))
		}
	}
}
