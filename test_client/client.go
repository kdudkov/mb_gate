package main

import (
	"fmt"
	"mb_gate/modbus"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:1502")
	if err != nil {
		panic("can't write")
	}

	pdu := modbus.ReadInputRegisters(1, 2, 1)
	data := pdu.ToTCP(1)

	fmt.Println(pdu)
	fmt.Println(data)
	conn.Write(pdu.ToTCP(1))
	res := make([]byte, 255)

	n, err := conn.Read(res)
	if err != nil {
		fmt.Printf("error %v\n", err)
		return
	}

	fmt.Println(res[:n])
}
