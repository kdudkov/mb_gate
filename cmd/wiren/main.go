package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/kdudkov/mb_gate/modbus"
	"net"
	"os"
)

type Sender struct {
	conn net.Conn
	trId uint16
}

func NewSender(addr string) (s *Sender) {
	s = &Sender{}

	conn, err := net.Dial("tcp", addr)
	//conn, err := net.Dial("tcp", "127.0.0.1:1502")
	if err != nil {
		fmt.Printf("can't write %v\n", err)
		os.Exit(1)
	}

	s.conn = conn
	return
}

func (s *Sender) Send(pdu *modbus.ProtocolDataUnit) *modbus.ProtocolDataUnit {
	data := pdu.MakeTCP(s.trId)
	s.trId++

	if _, err := s.conn.Write(data); err != nil {
		fmt.Printf("conn write error %s\n", err.Error())
		return nil
	}

	res := make([]byte, 255)

	n, err := s.conn.Read(res)
	if err != nil {
		fmt.Printf("conn read error %s\n", err.Error())
		return nil
	}

	_, ans, _ := modbus.FromTCP(res[:n])
	return ans
}

func main() {
	var host = flag.String("host", "192.168.1.1:1502", "host:port")

	flag.Parse()
	s := NewSender(*host)
	defer s.conn.Close()

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
	ans := s.Send(pdu)

	pdu = modbus.ReadCoils(42, 0, 4)
	ans = s.Send(pdu)

	fmt.Println(modbus.DecodeCoils(ans))
	//pdu = modbus.WriteSingleRegister(42, 110,192)
	//ans = s.Send(pdu)

	fmt.Println(ans)

	CheckDev(s, 42)
}

func CheckDev(s *Sender, addr uint16) {
	pdu := modbus.ReadHoldingRegisters(byte(addr), 200, 6)
	resp := s.Send(pdu)

	if resp != nil && resp.FunctionCode&0x80 == 0 {
		fmt.Printf("version: %s\n", GetString(resp))
	}

	pdu = modbus.ReadInputRegisters(byte(addr), 250, 16)
	resp = s.Send(pdu)

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
		s += string(val)
	}
	return s
}
