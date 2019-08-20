package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"mb_gate/modbus"
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
	var host = flag.String("host", "127.0.0.1:1502", "host:port")
	var fn = flag.Int("fn", 0, "function number")
	var dev = flag.Int("dev", 0, "device id")
	var addr = flag.Int("addr", 0, "address")
	var num = flag.Int("num", 1, "number of values")
	//var data = flag.String("data", "", "data to send")

	flag.Parse()
	s := NewSender(*host)
	defer s.conn.Close()

	switch *fn {
	case 1:
		pdu := modbus.ReadCoils(byte(*dev), uint16(*addr), uint16(*num))
		fmt.Printf("request: %s\n", pdu.ReqString())
		fmt.Printf("request raw: %s\n", pdu)
		resp := s.Send(pdu)
		if resp != nil {
			fmt.Printf("responce: %v\n", resp)
		}

		var i uint16
		for i = 0; i < uint16(*num); i++ {
			val := resp.Data[1+i>>3]&(1<<(i&7)) > 0
			fmt.Printf("  %d: %v\n", uint16(*addr)+i, val)
		}

	case 3:
		pdu := modbus.ReadHoldingRegisters(byte(*dev), uint16(*addr), uint16(*num))
		fmt.Printf("request: %s\n", pdu.ReqString())
		fmt.Printf("request raw: %s\n", pdu)
		resp := s.Send(pdu)
		if resp != nil {
			fmt.Printf("responce: %v\n", resp)
		}

		var i uint16
		for i = 0; i < uint16(*num); i++ {
			val := binary.BigEndian.Uint16(resp.Data[1+i*2:])
			fmt.Printf("  %d: %#x\n", uint16(*addr)+i, val)
		}

	default:
		fmt.Println("functions:")
		fmt.Println("  1 (0x01) — Read Coils.")
		fmt.Println("  2 (0x02) — Read Discrete Inputs.")
		fmt.Println("  3 (0x03) — Read Holding Registers.")
		fmt.Println("  4 (0x04) — Read Input Registers.")
		fmt.Println("  5 (0x05) — Write Single Coil.")
		fmt.Println("  6 (0x06) — Write Single Register.")
		os.Exit(1)
	}

}
