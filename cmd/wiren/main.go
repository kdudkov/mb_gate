package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/kdudkov/mb_gate/modbus"
	"os"
)

func main() {
	var host = flag.String("host", "192.168.1.1:1502", "host:port")

	flag.Parse()
	s := modbus.NewModbusSender(*host)

	err := s.Connect()
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		os.Exit(1)
	}

	defer s.Close()

	//for addr := 1; addr<90;addr++ {
	//	pdu := modbus.ReadHoldingRegisters(byte(addr), 200, 6)
	//	resp := s.Send(pdu)
	//
	//	if resp == nil || resp.FunctionCode & 0x80 != 0 {
	//		continue
	//	}
	//
	//	fmt.Printf("%d: %v\n", addr, resp)
	//}

	pdu := modbus.WriteSingleCoil(42, 1, false)
	ans, _ := s.Send(pdu)

	pdu = modbus.ReadCoils(42, 0, 4)
	ans, _ = s.Send(pdu)

	fmt.Println(modbus.DecodeCoils(ans))
	//pdu = modbus.WriteSingleRegister(42, 110,192)
	//ans = s.Send(pdu)

	fmt.Println(ans)

	CheckDev(s, 42)
}

func CheckDev(s *modbus.ModbusSender, addr uint16) {
	pdu := modbus.ReadHoldingRegisters(byte(addr), 200, 6)
	resp, _ := s.Send(pdu)

	if resp != nil && resp.FunctionCode&0x80 == 0 {
		fmt.Printf("version: %s\n", GetString(resp))
	}

	pdu = modbus.ReadInputRegisters(byte(addr), 250, 16)
	resp, _ = s.Send(pdu)

	if resp != nil && resp.FunctionCode&0x80 == 0 {
		fmt.Printf("version: %s\n", GetString(resp))
	} else {
		fmt.Println("Can't get firmware version")
		fmt.Println(resp)
	}
}

func GetString(pdu *modbus.ProtocolDataUnit) string {
	var i byte
	var s string
	for i = 0; i < pdu.Data[0]; i++ {
		val := binary.BigEndian.Uint16(pdu.Data[1+i*2:])
		if val == 0 {
			return s
		}
		s += fmt.Sprint(val)
	}
	return s
}
