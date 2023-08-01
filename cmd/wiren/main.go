package main

import (
	"flag"
	"fmt"
	"github.com/kdudkov/mb_gate/modbus"
	"os"
)

func main() {
	var host = flag.String("host", "192.168.1.2:1502", "host:port")

	flag.Parse()
	s := modbus.NewClient(*host)

	err := s.Connect()
	if err != nil {
		fmt.Printf("error: %s", err.Error())
		os.Exit(1)
	}

	defer s.Close()

	SearchDevices(s)
}

func SearchDevices(s *modbus.MbClient) {
	for addr := 1; addr < 254; addr++ {
		var t1, t2 bool

		if _, err := s.ReadCoils(byte(addr), 1, 1); err == nil {
			t1 = true
		}

		if _, err := s.ReadHoldingRegisters(byte(addr), 1, 1); err == nil {
			t2 = true
		}

		if !(t1 || t2) {
			continue
		}

		if v, err := GetWirenVersion(s, uint16(addr)); err == nil {
			fmt.Printf("%d: Wiren board %s\n", addr, v)
		} else {
			fmt.Printf("%d: not wiren device: %s\n", addr, err.Error())
		}
	}
}

func CheckDevice(s *modbus.MbClient, addr int) {
	s.WriteCoil(byte(addr), 1, false)
	dat, err := s.ReadCoils(byte(addr), 1, 6)

	if err != nil {
		fmt.Println(err.Error())
	}

	for _, i := range dat {
		fmt.Println(i)
	}
}

func GetWirenVersion(s *modbus.MbClient, addr uint16) (string, error) {
	result := ""

	if v, err := s.ReadString(byte(addr), 200, 6); err == nil {
		result += fmt.Sprintf("model: %s", v)
	} else {
		return result, err
	}

	if v, err := s.ReadString(byte(addr), 250, 16); err == nil {
		result += fmt.Sprintf("fw version: %s", v)
	} else {
		return result, err
	}

	return result, nil
}
