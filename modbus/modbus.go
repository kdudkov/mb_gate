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

type ModbusMessage interface {
	ToTCP(transactionId uint16) (adu []byte)
}

// ModbusError implements error interface.
type ModbusError struct {
	SlaveId       byte
	FunctionCode  byte
	ExceptionCode byte
}

type ProtocolDataUnit struct {
	SlaveId      byte
	FunctionCode byte
	Data         []byte
}

func (pdu *ProtocolDataUnit) String() string {
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

	return fmt.Sprintf("slave_id: %d, fn: %#.2x %s", pdu.SlaveId, pdu.FunctionCode, name)
}

func ReadDiscteteInputs(slaveId byte, addr uint16, count uint16) (pdu *ProtocolDataUnit) {
	pdu = &ProtocolDataUnit{SlaveId: slaveId, FunctionCode: FuncCodeReadDiscreteInputs}
	pdu.Data = make([]byte, 4)
	binary.BigEndian.PutUint16(pdu.Data, addr)
	binary.BigEndian.PutUint16(pdu.Data[2:], count)
	return
}

func ReadHoldingRegisters(slaveId byte, addr uint16, count uint16) (pdu *ProtocolDataUnit) {
	pdu = &ProtocolDataUnit{SlaveId: slaveId, FunctionCode: FuncCodeReadHoldingRegisters}
	pdu.Data = make([]byte, 4)
	binary.BigEndian.PutUint16(pdu.Data, addr)
	binary.BigEndian.PutUint16(pdu.Data[2:], count)
	return
}

func ReadInputRegisters(slaveId byte, addr uint16, count uint16) (pdu *ProtocolDataUnit) {
	pdu = &ProtocolDataUnit{SlaveId: slaveId, FunctionCode: FuncCodeReadInputRegisters}
	pdu.Data = make([]byte, 4)
	binary.BigEndian.PutUint16(pdu.Data, addr)
	binary.BigEndian.PutUint16(pdu.Data[2:], count)
	return
}

func NewModbusError(pdu *ProtocolDataUnit, errorCode byte) (e *ModbusError) {
	e = &ModbusError{}
	e.SlaveId = pdu.SlaveId
	e.FunctionCode = pdu.FunctionCode
	e.ExceptionCode = errorCode
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

func (pdu *ProtocolDataUnit) ToTCP(transactionId uint16) (adu []byte) {
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

func (pdu *ModbusError) ToTCP(transactionId uint16) (adu []byte) {
	adu = make([]byte, TcpHeaderSize+2)

	// Transaction identifier
	binary.BigEndian.PutUint16(adu, transactionId)
	// Protocol identifier
	binary.BigEndian.PutUint16(adu[2:], 0)
	binary.BigEndian.PutUint16(adu[4:], 2)
	adu[6] = pdu.SlaveId
	adu[7] = pdu.FunctionCode
	adu[8] = pdu.ExceptionCode

	return
}

func FromTCP(adu []byte) (transactionId uint16, pdu *ProtocolDataUnit, err error) {
	transactionId = binary.BigEndian.Uint16(adu)
	// Read length value in the header
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
