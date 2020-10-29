package modbus

import (
	"encoding/binary"
	"fmt"
	"net"
)

type MbClient struct {
	addr string
	conn net.Conn
	trId uint16
}

func NewClient(addr string) *MbClient {
	return &MbClient{addr: addr}
}

func (s *MbClient) Connect() error {
	if s.conn != nil {
		return nil
	}

	conn, err := net.Dial("tcp", s.addr)
	s.conn = conn
	return err
}

func (s *MbClient) Send(pdu *ProtocolDataUnit) (*ProtocolDataUnit, error) {
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

	_, ans, err := FromTCP(res[:n])
	return ans, err
}

func (s *MbClient) ReadCoils(slaveId byte, addr, count uint16) ([]bool, error) {
	pdu := ReadCoils(slaveId, addr, count)

	ans, err := s.Send(pdu)
	if err != nil {
		return nil, err
	}

	if ans.ErrString() != "" {
		return nil, fmt.Errorf("error %s", ans.ErrString())
	}
	return DecodeCoils(ans)
}

func (s *MbClient) ReadHoldingRegisters(slaveId byte, addr, count uint16) ([]uint16, error) {
	pdu := ReadHoldingRegisters(slaveId, addr, count)

	ans, err := s.Send(pdu)
	if err != nil {
		return nil, err
	}

	if ans.ErrString() != "" {
		return nil, fmt.Errorf("error %s", ans.ErrString())
	}
	return DecodeValues(ans)
}

func (s *MbClient) ReadString(slaveId byte, addr, count uint16) (string, error) {
	pdu := ReadHoldingRegisters(slaveId, addr, count)
	resp, err := s.Send(pdu)

	if err != nil {
		return "", err
	}

	if resp == nil {
		return "", fmt.Errorf("empty resp")
	}

	if resp.ErrString() != "" {
		return "", fmt.Errorf("error: %s", resp.ErrString())
	}

	return getString(resp)
}

func (s *MbClient) WriteCoil(slaveId byte, addr uint16, value bool) error {
	pdu := WriteSingleCoil(slaveId, addr, value)
	resp, err := s.Send(pdu)

	if err != nil {
		return err
	}

	if resp == nil {
		return fmt.Errorf("empty resp")
	}

	if resp.ErrString() != "" {
		return fmt.Errorf("error: %s", resp.ErrString())
	}

	return nil
}

func (s *MbClient) WriteHoldingRegister(slaveId byte, addr uint16, value uint16) error {
	pdu := WriteSingleRegister(slaveId, addr, value)
	resp, err := s.Send(pdu)

	if err != nil {
		return err
	}

	if resp == nil {
		return fmt.Errorf("empty resp")
	}

	if resp.ErrString() != "" {
		return fmt.Errorf("error: %s", resp.ErrString())
	}

	return nil
}

func (s *MbClient) WriteHoldingRegisters(slaveId byte, addr uint16, values []uint16) error {
	pdu := WriteMultipleRegisters(slaveId, addr, uint16(len(values)), values)
	resp, err := s.Send(pdu)

	if err != nil {
		return err
	}

	if resp == nil {
		return fmt.Errorf("empty resp")
	}

	if resp.ErrString() != "" {
		return fmt.Errorf("error: %s", resp.ErrString())
	}

	return nil
}

func getString(pdu *ProtocolDataUnit) (string, error) {
	var i byte
	var s string

	if len(pdu.Data) == 0 {
		return "", fmt.Errorf("no data")
	}

	size := pdu.Data[0]

	if len(pdu.Data) < 2*int(size)+1 {
		return "", fmt.Errorf("no data")
	}

	for i = 0; i < size; i++ {
		val := binary.BigEndian.Uint16(pdu.Data[1+i*2:])
		if val == 0 {
			return s, nil
		}
		s += fmt.Sprint(val)
	}
	return s, nil
}

func (s *MbClient) Close() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}
