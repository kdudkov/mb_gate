package modbus

import (
	"encoding/binary"
	"fmt"
)

const (
	// read
	FuncCodeReadCoils            = 1
	FuncCodeReadDiscreteInputs   = 2
	FuncCodeReadHoldingRegisters = 3
	FuncCodeReadInputRegisters   = 4
	// write one
	FuncCodeWriteSingleCoil     = 5
	FuncCodeWriteSingleRegister = 6
	// diag
	FuncCodeReadExceptionStatus = 7
	FuncCodeDiagnostic          = 8
	FuncCodeGetComEventCounter  = 11
	FuncCodeGetComEventLog      = 12
	FuncCodeReportSlaveId       = 17
	// write many
	FuncCodeWriteMultipleCoils     = 15
	FuncCodeWriteMultipleRegisters = 16
	// files
	FuncCodeReadFileRecord  = 20
	FuncCodeWriteFileRecord = 21
	// change
	FuncCodeMaskWriteRegister          = 22
	FuncCodeReadWriteMultipleRegisters = 23
	// fifo
	FuncCodeReadFIFOQueue = 24
	// other
	FuncCodeEncapsulatedInterfaceTransport = 43

	ExceptionCodeIllegalFunction                    = 1
	ExceptionCodeIllegalDataAddress                 = 2
	ExceptionCodeIllegalDataValue                   = 3
	ExceptionCodeServerDeviceFailure                = 4
	ExceptionCodeAcknowledge                        = 5
	ExceptionCodeServerDeviceBusy                   = 6
	ExceptionCodeMemoryParityError                  = 8
	ExceptionCodeGatewayPathUnavailable             = 10
	ExceptionCodeGatewayTargetDeviceFailedToRespond = 11

	RtuMinSize       = 4
	RtuMaxSize       = 256
	RtuExceptionSize = 5

	TcpHeaderSize = 7
	TcpMaxLength  = 260
)

type ProtocolDataUnit struct {
	SlaveId      byte
	FunctionCode byte
	Data         []byte
}

func (pdu *ProtocolDataUnit) String() string {
	if pdu.FunctionCode&0x80 == 0 {
		s := ""
		for _, d := range pdu.Data {
			if s != "" {
				s += " "
			}
			s += fmt.Sprintf("%#.2x", d)
		}
		return fmt.Sprintf("slaveId: %d, fn: %#.2x, data: %s", pdu.SlaveId, pdu.FunctionCode, s)
	} else {
		return fmt.Sprintf("error: slave_id: %d, fn: %#.2x, code: %#.2x", pdu.SlaveId, pdu.FunctionCode&0x7f, pdu.Data[0])
	}
}

func (pdu *ProtocolDataUnit) ReqString() string {
	var name string

	switch pdu.FunctionCode {
	case FuncCodeReadCoils:
		name = fmt.Sprintf("read coils, addr %#x, num %d", binary.BigEndian.Uint16(pdu.Data[0:]), binary.BigEndian.Uint16(pdu.Data[2:]))
	case FuncCodeReadDiscreteInputs:
		name = fmt.Sprintf("read discrete inputs, addr %#x, num %d", binary.BigEndian.Uint16(pdu.Data[0:]), binary.BigEndian.Uint16(pdu.Data[2:]))
	case FuncCodeReadHoldingRegisters:
		name = fmt.Sprintf("read holding registers, addr %#x, num %d", binary.BigEndian.Uint16(pdu.Data[0:]), binary.BigEndian.Uint16(pdu.Data[2:]))
	case FuncCodeReadInputRegisters:
		name = fmt.Sprintf("read input registers, addr %#x, num %d", binary.BigEndian.Uint16(pdu.Data[0:]), binary.BigEndian.Uint16(pdu.Data[2:]))

	case FuncCodeWriteSingleCoil:
		name = fmt.Sprintf("write coil, addr %#x, val %#.4x", binary.BigEndian.Uint16(pdu.Data[0:]), binary.BigEndian.Uint16(pdu.Data[2:]))
	case FuncCodeWriteSingleRegister:
		name = fmt.Sprintf("write register, addr %#x, val %#.4x", binary.BigEndian.Uint16(pdu.Data[0:]), binary.BigEndian.Uint16(pdu.Data[2:]))

	case FuncCodeWriteMultipleCoils:
		name = fmt.Sprintf("write coils, addr %#x, num %d", binary.BigEndian.Uint16(pdu.Data[0:]), binary.BigEndian.Uint16(pdu.Data[2:]))
	case FuncCodeWriteMultipleRegisters:
		name = fmt.Sprintf("write registers, addr %#x, num %d", binary.BigEndian.Uint16(pdu.Data[0:]), binary.BigEndian.Uint16(pdu.Data[2:]))

	default:
		name = "unknown"
	}

	if pdu.FunctionCode&0x80 == 0 {
		return fmt.Sprintf("slave_id: %d, fn: %#.2x %s", pdu.SlaveId, pdu.FunctionCode, name)
	} else {
		return fmt.Sprintf("error: slave_id: %d, fn: %#.2x, code: %#.2x", pdu.SlaveId, pdu.FunctionCode, pdu.Data[0])
	}
}

func readManyPDU(slaveId byte, fn byte, addr uint16, count uint16) (pdu *ProtocolDataUnit) {
	pdu = &ProtocolDataUnit{SlaveId: slaveId, FunctionCode: fn}
	pdu.Data = make([]byte, 4)
	binary.BigEndian.PutUint16(pdu.Data, addr)
	binary.BigEndian.PutUint16(pdu.Data[2:], count)
	return
}

func ReadCoils(slaveId byte, addr uint16, count uint16) *ProtocolDataUnit {
	return readManyPDU(slaveId, FuncCodeReadCoils, addr, count)
}

func ReadDiscteteInputs(slaveId byte, addr uint16, count uint16) *ProtocolDataUnit {
	return readManyPDU(slaveId, FuncCodeReadDiscreteInputs, addr, count)
}

func ReadHoldingRegisters(slaveId byte, addr uint16, count uint16) (pdu *ProtocolDataUnit) {
	return readManyPDU(slaveId, FuncCodeReadHoldingRegisters, addr, count)
}

func ReadInputRegisters(slaveId byte, addr uint16, count uint16) (pdu *ProtocolDataUnit) {
	return readManyPDU(slaveId, FuncCodeReadInputRegisters, addr, count)
}

func WriteSingleCoil(slaveId byte, addr uint16, val bool) (pdu *ProtocolDataUnit) {
	var v uint16 = 0
	if val {
		v = 0xff00
	}
	return WriteSingleCoilRaw(slaveId, addr, v)
}

func WriteSingleCoilRaw(slaveId byte, addr uint16, val uint16) (pdu *ProtocolDataUnit) {
	pdu = &ProtocolDataUnit{SlaveId: slaveId, FunctionCode: FuncCodeWriteSingleCoil}
	pdu.Data = make([]byte, 4)
	binary.BigEndian.PutUint16(pdu.Data, addr)
	binary.BigEndian.PutUint16(pdu.Data[2:], val)
	return
}

func WriteMultipleCoilsRaw(slaveId byte, addr uint16, num uint16, data []byte) (pdu *ProtocolDataUnit) {
	pdu = &ProtocolDataUnit{SlaveId: slaveId, FunctionCode: FuncCodeWriteMultipleCoils}
	pdu.Data = make([]byte, 6+len(data))
	binary.BigEndian.PutUint16(pdu.Data, addr)
	binary.BigEndian.PutUint16(pdu.Data[2:], num)
	pdu.Data[4] = byte(len(data))
	var i uint16
	for i = 0; i < uint16(len(data)); i++ {
		pdu.Data[5+i] = data[i]
	}
	return
}

func WriteSingleRegister(slaveId byte, addr uint16, val uint16) (pdu *ProtocolDataUnit) {
	pdu = &ProtocolDataUnit{SlaveId: slaveId, FunctionCode: FuncCodeWriteSingleRegister}
	pdu.Data = make([]byte, 4)
	binary.BigEndian.PutUint16(pdu.Data, addr)
	binary.BigEndian.PutUint16(pdu.Data[2:], val)
	return
}

func WriteMultipleRegisters(slaveId byte, addr uint16, count uint16, values []uint16) (pdu *ProtocolDataUnit) {
	pdu = &ProtocolDataUnit{SlaveId: slaveId, FunctionCode: FuncCodeWriteMultipleRegisters}
	pdu.Data = make([]byte, 5+2*count)
	binary.BigEndian.PutUint16(pdu.Data, addr)
	binary.BigEndian.PutUint16(pdu.Data[2:], count)
	pdu.Data[4] = byte(2 * len(values))
	var i uint16
	for i = 0; i < count; i++ {
		binary.BigEndian.PutUint16(pdu.Data[5+2*i:], values[i])
	}

	return
}

func NewModbusError(pdu *ProtocolDataUnit, errorCode byte) (e *ProtocolDataUnit) {
	e = &ProtocolDataUnit{}
	e.SlaveId = pdu.SlaveId
	e.FunctionCode = pdu.FunctionCode | 0x80
	e.Data = []byte{errorCode}
	return
}

func (pdu *ProtocolDataUnit) MakeRtu() (adu []byte, err error) {
	length := len(pdu.Data) + 4
	if length > RtuMaxSize {
		err = fmt.Errorf("modbus: length of data '%v' must not be bigger than '%v'", length, RtuMaxSize)
		return
	}
	adu = make([]byte, length)

	adu[0] = pdu.SlaveId
	adu[1] = pdu.FunctionCode
	copy(adu[2:], pdu.Data)

	var crc crc
	crc.reset().pushBytes(adu[0 : length-2])
	checksum := crc.value()

	adu[length-1] = byte(checksum >> 8)
	adu[length-2] = byte(checksum)
	return
}

func FromRtu(adu []byte) (pdu *ProtocolDataUnit, err error) {
	length := len(adu)
	// Calculate checksum
	var crc crc
	crc.reset().pushBytes(adu[0 : length-2])
	checksum := uint16(adu[length-1])<<8 | uint16(adu[length-2])

	if checksum != crc.value() {
		err = fmt.Errorf("modbus: response crc '%v' does not match expected '%v'", checksum, crc.value())
		return
	}

	// Function code & data
	pdu = &ProtocolDataUnit{}
	pdu.SlaveId = adu[0]
	pdu.FunctionCode = adu[1]
	pdu.Data = adu[2 : length-2]
	return
}

func (pdu *ProtocolDataUnit) MakeTCP(transactionId uint16) (adu []byte) {
	adu = make([]byte, TcpHeaderSize+1+len(pdu.Data))

	// Transaction identifier
	binary.BigEndian.PutUint16(adu, transactionId)
	// Protocol identifier
	binary.BigEndian.PutUint16(adu[2:], 0)
	// Length = sizeof(SlaveId) + sizeof(FunctionCode) + Data
	length := uint16(1 + 1 + len(pdu.Data))
	binary.BigEndian.PutUint16(adu[4:], length)
	adu[6] = pdu.SlaveId
	adu[7] = pdu.FunctionCode
	copy(adu[8:], pdu.Data)
	return
}

func FromTCP(adu []byte) (transactionId uint16, pdu *ProtocolDataUnit, err error) {
	transactionId = binary.BigEndian.Uint16(adu)
	// readManyPDU length value in the header
	length := binary.BigEndian.Uint16(adu[4:])
	pduLength := len(adu) - TcpHeaderSize

	if pduLength <= 0 || pduLength != int(length-1) {
		err = fmt.Errorf("modbus: length in response '%v' does not match pdu data length '%v'", length-1, pduLength)
		return
	}

	pdu = &ProtocolDataUnit{}
	pdu.SlaveId = adu[6]
	// The first byte after header is function code
	pdu.FunctionCode = adu[TcpHeaderSize]
	pdu.Data = adu[TcpHeaderSize+1:]
	return
}

func DecodeCoils(pdu *ProtocolDataUnit) []bool {
	var i byte

	if pdu.FunctionCode != FuncCodeReadCoils && pdu.FunctionCode != FuncCodeReadDiscreteInputs {
		return nil
	}

	res := make([]bool, pdu.Data[0]*8)

	for i = 0; i < pdu.Data[0]*8; i++ {
		res[i] = pdu.Data[1+i>>3]&(1<<(i&7)) > 0
	}

	return res
}

func DecodeValuse(pdu *ProtocolDataUnit) []uint16 {
	var i byte

	if pdu.FunctionCode != FuncCodeReadInputRegisters && pdu.FunctionCode != FuncCodeReadHoldingRegisters {
		return nil
	}

	res := make([]uint16, pdu.Data[0])

	for i = 0; i < pdu.Data[0]; i++ {
		res[i] = binary.BigEndian.Uint16(pdu.Data[1+i*2:])
	}

	return res
}
