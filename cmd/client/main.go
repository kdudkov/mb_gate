package main

import (
	"flag"
	"fmt"
	"github.com/kdudkov/mb_gate/modbus"
	"os"
)

func main() {
	var host = flag.String("host", "127.0.0.1:1502", "host:port")
	var fn = flag.Int("fn", 0, "function number")
	var dev = flag.Int("dev", 0, "device id")
	var addr = flag.Int("addr", 0, "address")
	var num = flag.Int("num", 1, "number of values")
	//var data = flag.String("data", "", "data to send")

	flag.Parse()
	s, err := modbus.NewModbusSender(*host)

	if err != nil {
		fmt.Printf("error: %s", err.Error())
		os.Exit(1)
	}

	defer s.Close()

	switch *fn {
	case 1:
		pdu := modbus.ReadCoils(byte(*dev), uint16(*addr), uint16(*num))
		fmt.Printf("request: %s\n", pdu.ReqString())
		fmt.Printf("request raw: %s\n", pdu)
		resp, err := s.Send(pdu)

		if err != nil {
			fmt.Printf("error: %s", err.Error())
			return
		}

		fmt.Printf("responce: %v\n", resp)

		res := modbus.DecodeCoils(resp)
		for i := 0; i < *num; i++ {
			fmt.Printf("  %d: %v\n", *addr+i, res[i])
		}

	case 3:
		pdu := modbus.ReadHoldingRegisters(byte(*dev), uint16(*addr), uint16(*num))
		fmt.Printf("request: %s\n", pdu.ReqString())
		fmt.Printf("request raw: %s\n", pdu)
		resp, err := s.Send(pdu)

		if err != nil {
			fmt.Printf("error: %s", err.Error())
			return
		}

		fmt.Printf("responce: %v\n", resp)
		res := modbus.DecodeValuse(pdu)

		for i := 0; i < *num; i++ {
			fmt.Printf("  %d: %#x\n", *addr+i, res[i])
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
