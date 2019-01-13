package main

import (
	"encoding/binary"
	"mb_gate/modbus"
)

type Translator interface {
	Translate(pdu *modbus.ProtocolDataUnit) (dontSend bool)
}

type SimpleChineese struct {
	registers map[uint16]uint16
}

func NewSimpleChinese() (t *SimpleChineese) {
	t = &SimpleChineese{}
	t.registers = make(map[uint16]uint16)
	return
}

func (s *SimpleChineese) Translate(pdu *modbus.ProtocolDataUnit) (dontSend bool) {
	if pdu.FunctionCode == modbus.FuncCodeWriteSingleRegister {
		addr := binary.BigEndian.Uint16(pdu.Data)
		val := binary.BigEndian.Uint16(pdu.Data[2:])
		s.registers[addr] = val
		if val == 0 {
			binary.BigEndian.PutUint16(pdu.Data[2:], 0x200)
		} else {
			binary.BigEndian.PutUint16(pdu.Data[2:], 0x100)
		}
		return
	}

	if pdu.FunctionCode == modbus.FuncCodeReadHoldingRegisters {
		addr := binary.BigEndian.Uint16(pdu.Data)
		num := binary.BigEndian.Uint16(pdu.Data[2:])
		pdu2 := &modbus.ProtocolDataUnit{}
		pdu2.SlaveId = pdu.SlaveId
		pdu2.FunctionCode = pdu.FunctionCode
		pdu.Data = make([]byte, num*2+1)
		pdu.Data[0] = byte(num * 2)
		var i uint16
		for i = 0; i < num; i++ {
			binary.BigEndian.PutUint16(pdu.Data[2*i+1:], s.registers[addr+i])
		}
		dontSend = true
		return
	}
	return
}
