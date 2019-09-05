package client

import (
	"net"

	"github.com/kdudkov/mb_gate/modbus"
)

type ModbusSender struct {
	conn net.Conn
	trId uint16
}

func NewModbusSender(addr string) (s *ModbusSender, err error) {
	s = &ModbusSender{}

	conn, err := net.Dial("tcp", addr)
	s.conn = conn
	return s, err
}

func (s *ModbusSender) Send(pdu *modbus.ProtocolDataUnit) (*modbus.ProtocolDataUnit, error) {
	data := pdu.MakeTCP(s.trId)
	s.trId++

	if _, err := s.conn.Write(data); err != nil {
		return nil, err
	}

	res := make([]byte, 255)

	n, err := s.conn.Read(res)
	if err != nil {
		return nil, err
	}

	_, ans, _ := modbus.FromTCP(res[:n])
	return ans, nil
}

func (s *ModbusSender) Close() error {
	return s.conn.Close()
}
