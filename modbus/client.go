package modbus

import (
	"net"
)

type ModbusSender struct {
	addr string
	conn net.Conn
	trId uint16
}

func NewModbusSender(addr string) *ModbusSender {
	return &ModbusSender{addr: addr}
}

func (s *ModbusSender) Connect() error {
	if s.conn != nil {
		return nil
	}

	conn, err := net.Dial("tcp", s.addr)
	s.conn = conn
	return err
}

func (s *ModbusSender) Send(pdu *ProtocolDataUnit) (*ProtocolDataUnit, error) {
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

	_, ans, _ := FromTCP(res[:n])
	return ans, nil
}

func (s *ModbusSender) Close() error {
	return s.conn.Close()
}
