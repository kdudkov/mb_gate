package main

import (
	"encoding/binary"
	"sync"

	"github.com/kdudkov/mb_gate/modbus"
)

type Translator interface {
	Translate(pdu *modbus.ProtocolDataUnit) (dontSend bool)
}

type SimpleChineese struct {
	registers map[uint16]uint16
	mutex     sync.Mutex
}

func NewSimpleChinese() *SimpleChineese {
	return &SimpleChineese{registers: map[uint16]uint16{}, mutex: sync.Mutex{}}
}

func (s *SimpleChineese) Translate(pdu *modbus.ProtocolDataUnit) (dontSend bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	switch pdu.FunctionCode {
	case modbus.FuncCodeWriteSingleRegister:
		addr := binary.BigEndian.Uint16(pdu.Data)
		val := binary.BigEndian.Uint16(pdu.Data[2:])
		s.registers[addr] = val
		// translate 0 (off) -> 0x200, other (on) -> 0x100
		// this shit works this way, don't ask me why
		if val == 0 {
			binary.BigEndian.PutUint16(pdu.Data[2:], 0x200)
		} else {
			binary.BigEndian.PutUint16(pdu.Data[2:], 0x100)
		}

	case modbus.FuncCodeWriteMultipleRegisters:
		addr := binary.BigEndian.Uint16(pdu.Data)
		num := binary.BigEndian.Uint16(pdu.Data[2:])
		var i uint16
		for i = 0; i < num; i++ {
			val := binary.BigEndian.Uint16(pdu.Data[5+2*i:])
			s.registers[addr+i] = val
			if val == 0 {
				binary.BigEndian.PutUint16(pdu.Data[5+2*i:], 0x200)
			} else {
				binary.BigEndian.PutUint16(pdu.Data[5+2*i:], 0x100)
			}
		}

	case modbus.FuncCodeReadHoldingRegisters:
		addr := binary.BigEndian.Uint16(pdu.Data)
		num := binary.BigEndian.Uint16(pdu.Data[2:])
		// make new data for pdu
		pdu.Data = make([]byte, num*2+1)
		pdu.Data[0] = byte(num * 2)

		var i uint16
		for i = 0; i < num; i++ {
			binary.BigEndian.PutUint16(pdu.Data[2*i+1:], s.registers[addr+i])
		}
		dontSend = true
	}

	return
}

type FakeTranslator struct {
	registers []uint16
	coils     []bool
	mutex     sync.Mutex
}

func NewFakeTranslator() *FakeTranslator {
	f := &FakeTranslator{}
	f.registers = make([]uint16, 65535)
	f.coils = make([]bool, 65535)
	f.mutex = sync.Mutex{}
	return f
}

func (t *FakeTranslator) Translate(pdu *modbus.ProtocolDataUnit) bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	switch pdu.FunctionCode {
	case modbus.FuncCodeWriteSingleRegister:
		addr := binary.BigEndian.Uint16(pdu.Data)
		val := binary.BigEndian.Uint16(pdu.Data[2:])
		t.registers[addr] = val
		return true

	case modbus.FuncCodeWriteMultipleRegisters:
		addr := binary.BigEndian.Uint16(pdu.Data)
		num := binary.BigEndian.Uint16(pdu.Data[2:])
		var i uint16
		for i = 0; i < num; i++ {
			t.registers[addr+i] = binary.BigEndian.Uint16(pdu.Data[5+2*i:])
		}
		return true

	case modbus.FuncCodeReadHoldingRegisters:
		addr := binary.BigEndian.Uint16(pdu.Data)
		num := binary.BigEndian.Uint16(pdu.Data[2:])
		// make new data for pdu
		pdu.Data = make([]byte, num*2+1)
		pdu.Data[0] = byte(num * 2)

		var i uint16
		for i = 0; i < num; i++ {
			binary.BigEndian.PutUint16(pdu.Data[2*i+1:], t.registers[addr+i])
		}
		return true

	case modbus.FuncCodeWriteSingleCoil:
		addr := binary.BigEndian.Uint16(pdu.Data)
		val := binary.BigEndian.Uint16(pdu.Data[2:])
		if val == 0xff00 {
			t.coils[addr] = true
		} else {
			t.coils[addr] = false
		}
		return true

	case modbus.FuncCodeWriteMultipleCoils:
		addr := binary.BigEndian.Uint16(pdu.Data)
		num := binary.BigEndian.Uint16(pdu.Data[2:])

		var ii uint16
		for ii = 0; ii < num; ii++ {
			val := pdu.Data[5+ii>>3]&(1<<(ii&7)) > 0
			t.coils[addr+ii] = val
		}
		return true

	case modbus.FuncCodeReadCoils:
		addr := binary.BigEndian.Uint16(pdu.Data)
		num := binary.BigEndian.Uint16(pdu.Data[2:])
		retBytes := num >> 3
		if num&7 > 0 {
			retBytes++
		}
		pdu.Data = make([]byte, retBytes+1)
		pdu.Data[0] = byte(retBytes)

		var ii uint16
		for ii = 0; ii < num; ii++ {
			var v byte = 0
			if t.coils[addr+ii] {
				v = 1
			}
			pdu.Data[1+ii>>3] = pdu.Data[1+ii>>3] | (v << (ii & 7))
		}
		return true
	}

	return true
}
