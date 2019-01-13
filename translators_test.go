package main

import (
	"encoding/binary"
	"fmt"
	"mb_gate/modbus"
	"testing"
)

func TestSimpleOn(t *testing.T) {
	tr := NewSimpleChinese()

	pdu := modbus.WriteSingleRegister(5, 1, 25)

	dontSend := tr.Translate(pdu)
	if dontSend {
		t.Fail()
	}
	addr := binary.BigEndian.Uint16(pdu.Data)
	val := binary.BigEndian.Uint16(pdu.Data[2:])

	if addr != 1 {
		t.Fail()
	}

	if val != 0x100 {
		t.Fail()
	}
	fmt.Println(pdu)
}

func TestSimpleOff(t *testing.T) {
	tr := NewSimpleChinese()

	pdu := modbus.WriteSingleRegister(5, 1, 0)

	dontSend := tr.Translate(pdu)
	if dontSend {
		t.Fail()
	}
	addr := binary.BigEndian.Uint16(pdu.Data)
	val := binary.BigEndian.Uint16(pdu.Data[2:])

	if addr != 1 {
		t.Fail()
	}

	if val != 0x200 {
		t.Fail()
	}
	fmt.Println(pdu)
}

func TestWrite(t *testing.T) {
	tr := NewSimpleChinese()

	pdu := modbus.WriteSingleRegister(5, 1, 25)
	tr.Translate(pdu)

	pdu = modbus.ReadHoldingRegisters(5, 1, 2)
	dontSend := tr.Translate(pdu)

	if !dontSend {
		t.Fail()
	}
	if pdu.Data[0] != 4 {
		t.Fatalf("len = %d", pdu.Data[0])
	}
	if binary.BigEndian.Uint16(pdu.Data[1:]) != 25 {
		t.Fail()
	}
	if binary.BigEndian.Uint16(pdu.Data[3:]) != 0 {
		t.Fail()
	}
}
