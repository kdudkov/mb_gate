package modbus

import (
	"encoding/binary"
	"io"
	"log"
	"os"
	"time"

	"github.com/goburrow/serial"
)

const (
	serialTimeout     = 5 * time.Second
	serialIdleTimeout = 60 * time.Second
)

type SerialPort struct {
	serial.Config

	Logger      *log.Logger
	IdleTimeout time.Duration

	port         io.ReadWriteCloser
	lastActivity time.Time
	closeTimer   *time.Timer
}

func NewSerial(device string, baudrate int) (s *SerialPort) {
	s = &SerialPort{}
	s.Address = device
	s.BaudRate = baudrate
	s.DataBits = 8
	s.Parity = "N"
	s.StopBits = 1
	s.Timeout = serialTimeout
	s.Logger = log.New(os.Stdout, "serial: ", log.LstdFlags)
	s.IdleTimeout = serialIdleTimeout
	return
}

func (mb *SerialPort) connect() error {
	if mb.port == nil {
		port, err := serial.Open(&mb.Config)
		if err != nil {
			return err
		}
		mb.port = port
	}
	return nil
}

func (mb *SerialPort) close() (err error) {
	if mb.port != nil {
		err = mb.port.Close()
		mb.port = nil
	}
	return
}

func (mb *SerialPort) logf(format string, v ...interface{}) {
	if mb.Logger != nil {
		mb.Logger.Printf(format, v...)
	}
}

func (mb *SerialPort) startCloseTimer() {
	if mb.IdleTimeout <= 0 {
		return
	}
	if mb.closeTimer == nil {
		mb.closeTimer = time.AfterFunc(mb.IdleTimeout, mb.closeIdle)
	} else {
		mb.closeTimer.Reset(mb.IdleTimeout)
	}
}

// closeIdle closes the connection if last activity is passed behind IdleTimeout.
func (mb *SerialPort) closeIdle() {
	if mb.IdleTimeout <= 0 {
		return
	}
	idle := time.Now().Sub(mb.lastActivity)
	if idle >= mb.IdleTimeout {
		mb.logf("modbus: closing connection due to idle timeout: %v", idle)
		mb.close()
	}
}

func (mb *SerialPort) Send(aduRequest []byte) (aduResponse []byte, err error) {
	// Make sure port is connected
	if err = mb.connect(); err != nil {
		return
	}
	// Start the timer to close when idle
	mb.lastActivity = time.Now()
	mb.startCloseTimer()

	// Send the request
	mb.logf("modbus: sending % x\n", aduRequest)
	if _, err = mb.port.Write(aduRequest); err != nil {
		return
	}
	function := aduRequest[1]
	functionFail := aduRequest[1] & 0x80
	bytesToRead := calculateResponseLength(aduRequest)
	time.Sleep(mb.calculateDelay(len(aduRequest) + bytesToRead))

	var n int
	var n1 int
	var data [rtuMaxSize]byte
	//We first read the minimum length and then read either the full package
	//or the error package, depending on the error status (byte 2 of the response)
	n, err = io.ReadAtLeast(mb.port, data[:], rtuMinSize)
	if err != nil {
		return
	}
	//if the function is correct
	if data[1] == function {
		//we read the rest of the bytes
		if n < bytesToRead {
			if bytesToRead > rtuMinSize && bytesToRead <= rtuMaxSize {
				if bytesToRead > n {
					n1, err = io.ReadFull(mb.port, data[n:bytesToRead])
					n += n1
				}
			}
		}
	} else if data[1] == functionFail {
		//for error we need to read 5 bytes
		if n < rtuExceptionSize {
			n1, err = io.ReadFull(mb.port, data[n:rtuExceptionSize])
		}
		n += n1
	}

	if err != nil {
		return
	}
	aduResponse = data[:n]
	mb.logf("modbus: received % x\n", aduResponse)
	return
}

// calculateDelay roughly calculates time needed for the next frame.
// See MODBUS over Serial Line - Specification and Implementation Guide (page 13).
func (mb *SerialPort) calculateDelay(chars int) time.Duration {
	var characterDelay, frameDelay int // us

	if mb.BaudRate <= 0 || mb.BaudRate > 19200 {
		characterDelay = 750
		frameDelay = 1750
	} else {
		characterDelay = 15000000 / mb.BaudRate
		frameDelay = 35000000 / mb.BaudRate
	}
	return time.Duration(characterDelay*chars+frameDelay) * time.Microsecond
}

func calculateResponseLength(adu []byte) int {
	length := rtuMinSize
	switch adu[1] {
	case FuncCodeReadDiscreteInputs,
		FuncCodeReadCoils:
		count := int(binary.BigEndian.Uint16(adu[4:]))
		length += 1 + count/8
		if count%8 != 0 {
			length++
		}
	case FuncCodeReadInputRegisters,
		FuncCodeReadHoldingRegisters,
		FuncCodeReadWriteMultipleRegisters:
		count := int(binary.BigEndian.Uint16(adu[4:]))
		length += 1 + count*2
	case FuncCodeWriteSingleCoil,
		FuncCodeWriteMultipleCoils,
		FuncCodeWriteSingleRegister,
		FuncCodeWriteMultipleRegisters:
		length += 4
	case FuncCodeMaskWriteRegister:
		length += 6
	case FuncCodeReadFIFOQueue:
		// undetermined
	default:
	}
	return length
}
