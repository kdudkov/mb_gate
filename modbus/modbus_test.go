package modbus

import (
	"fmt"
	"testing"
)

func TestFromRtu(t *testing.T) {
	pdu, err := FromRtu([]byte{0x1, 0x5, 0, 0x1, 0x1, 0, 0x9d, 0x9a})

	if err != nil {
		t.Fatalf("error %v", err)
	}

	fmt.Printf("pdu: %v\n", pdu)

	if pdu.FunctionCode != 5 {
		t.Fatalf("got function %d, expected %d", pdu.FunctionCode, 5)
	}

	pdu, err = FromRtu([]byte{0x1, 0xf, 0, 0x13, 0, 0xa, 0x2, 0xcd, 0x1, 0x72, 0xcb})

	if err != nil {
		t.Fatalf("error %v", err)
	}

	fmt.Printf("pdu: %v\n", pdu)

	if pdu.FunctionCode != 0xf {
		t.Fatalf("got function %d, expected %d", pdu.FunctionCode, 0xf)
	}

	pdu, err = FromRtu([]byte{0x2, 0xf, 0, 0x13, 0, 0xa, 0x2, 0xcd, 0x1, 0x72, 0xcb})

	if err == nil {
		t.Fatalf("invalid crc passed")
	}
}
