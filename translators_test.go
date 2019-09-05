package main

import (
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/kdudkov/mb_gate/modbus"
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

func TestSimpleOnMulti(t *testing.T) {
	tr := NewSimpleChinese()

	pdu := modbus.WriteMultipleRegisters(5, 1, 2, []uint16{10, 0})

	dontSend := tr.Translate(pdu)
	if dontSend {
		t.Error("should send")
	}
	addr := binary.BigEndian.Uint16(pdu.Data)
	val1 := binary.BigEndian.Uint16(pdu.Data[5:])
	val2 := binary.BigEndian.Uint16(pdu.Data[7:])

	if addr != 1 {
		t.Error("wrong address")
	}

	if val1 != 0x100 {
		t.Error("wrong val1")
	}
	if val2 != 0x200 {
		t.Error("wrong val2")
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

func TestWriteFake(t *testing.T) {
	tr := NewFakeTranslator()

	pdu := modbus.WriteSingleRegister(5, 2, 10)

	tr.Translate(pdu)

	if tr.registers[2] != 10 {
		t.Errorf("wrong value in reg %d, has %d", 2, tr.registers[2])
	}
}

func TestWriteFakeMany(t *testing.T) {
	tr := NewFakeTranslator()

	pdu := modbus.WriteMultipleRegisters(5, 2, 2, []uint16{10, 20})

	tr.Translate(pdu)

	for i, v := range []uint16{0, 0, 10, 20, 0, 0} {
		if tr.registers[uint16(i)] != v {
			t.Errorf("wrong value in reg %d, has %d", i, tr.registers[uint16(i)])
		}

	}
}

func TestWriteFakeCoil(t *testing.T) {
	tr := NewFakeTranslator()

	pdu := modbus.WriteSingleCoil(5, 2, true)

	tr.Translate(pdu)

	if tr.coils[2] != true {
		t.Error("invalid value")
	}
}

func TestWriteFakeWriteCoils(t *testing.T) {
	tr := NewFakeTranslator()

	pdu := modbus.WriteMultipleCoilsRaw(5, 2, 3, []byte{6})
	tr.Translate(pdu)

	for i, v := range []bool{false, false, false, true, true, false, false} {
		if tr.coils[uint16(i)] != v {
			t.Errorf("wrong value in coil %d", i)
		}

	}

	pdu = modbus.WriteMultipleCoilsRaw(5, 1, 10, []byte{6, 1})
	tr.Translate(pdu)

	for i, v := range []bool{false, false, true, true, false, false, false, false, false, true, false} {
		if tr.coils[uint16(i)] != v {
			t.Errorf("wrong value in coil %d", i)
		}

	}

	pdu = modbus.ReadCoils(5, 1, 10)
	tr.Translate(pdu)

	for i, v := range []byte{2, 6, 1} {
		if pdu.Data[i] != v {
			t.Errorf("wrong data: %v", pdu.Data)
		}
	}

	pdu = modbus.ReadCoils(5, 1, 16)
	tr.Translate(pdu)

	for i, v := range []byte{2, 6, 1} {
		if pdu.Data[i] != v {
			t.Errorf("wrong data: %v", pdu.Data)
		}
	}

	pdu = modbus.ReadCoils(5, 1, 17)
	tr.Translate(pdu)

	for i, v := range []byte{3, 6, 1, 0} {
		if pdu.Data[i] != v {
			t.Errorf("wrong data: %v", pdu.Data)
		}
	}
}
