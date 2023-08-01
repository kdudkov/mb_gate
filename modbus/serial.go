package modbus

import (
	"encoding/binary"
	"io"
	"time"

	"go.uber.org/zap"

	"github.com/goburrow/serial"
)

const (
	serialTimeout     = 500 * time.Millisecond
	serialIdleTimeout = 60 * time.Second
)

type SerialPort struct {
	serial.Config

	IdleTimeout  time.Duration
	port         io.ReadWriteCloser
	lastActivity time.Time
	closeTimer   *time.Timer
	Logger       *zap.SugaredLogger
}

func NewSerial(device string, baudrate int, data int, parity string, stop int) (s *SerialPort) {
	s = &SerialPort{}
	s.Address = device
	s.BaudRate = baudrate
	s.DataBits = data
	s.Parity = parity
	s.StopBits = stop
	s.Timeout = serialTimeout
	s.IdleTimeout = serialIdleTimeout
	return
}

func (sp *SerialPort) connect() error {
	if sp.port == nil {
		port, err := serial.Open(&sp.Config)
		if err != nil {
			return err
		}
		sp.port = port
	}
	return nil
}

func (sp *SerialPort) close() (err error) {
	if sp.port != nil {
		err = sp.port.Close()
		sp.port = nil
	}
	return
}

func (sp *SerialPort) startCloseTimer() {
	if sp.IdleTimeout <= 0 {
		return
	}
	if sp.closeTimer == nil {
		sp.closeTimer = time.AfterFunc(sp.IdleTimeout, sp.closeIdle)
	} else {
		sp.closeTimer.Reset(sp.IdleTimeout)
	}
}

// closeIdle closes the connection if last activity is passed behind IdleTimeout.
func (sp *SerialPort) closeIdle() {
	if sp.IdleTimeout <= 0 {
		return
	}
	idle := time.Now().Sub(sp.lastActivity)
	if idle >= sp.IdleTimeout {
		sp.Logger.Errorf("serial: closing connection due to idle timeout: %v", idle)
		sp.close()
	}
}

func (sp *SerialPort) Send(aduRequest []byte) (aduResponse []byte, err error) {
	// Make sure port is connected
	if err = sp.connect(); err != nil {
		sp.Logger.Errorf("serial: can't connect: %s", err.Error())
		return
	}
	// Start the timer to close when idle
	sp.lastActivity = time.Now()
	sp.startCloseTimer()

	// Send the request
	sp.Logger.Debugf("serial: sending %x", aduRequest)
	if _, err = sp.port.Write(aduRequest); err != nil {
		sp.Logger.Errorf("serial: write error %s", err.Error())
		return
	}
	function := aduRequest[1]
	bytesToRead := calculateResponseLength(aduRequest)
	time.Sleep(sp.calculateDelay(len(aduRequest) + bytesToRead))

	var n int
	var n1 int
	var data [RtuMaxSize]byte
	//We first read the minimum length and then read either the full package
	//or the error package, depending on the error status (byte 2 of the response)
	n, err = io.ReadAtLeast(sp.port, data[:], RtuMinSize)
	if err != nil {
		sp.Logger.Errorf("serial: read header error %s", err.Error())
		return
	}
	//if the function is correct
	if data[1] == function {
		//we read the rest of the bytes
		if n < bytesToRead {
			if bytesToRead > RtuMinSize && bytesToRead <= RtuMaxSize {
				if bytesToRead > n {
					n1, err = io.ReadFull(sp.port, data[n:bytesToRead])
					n += n1
				}
			}
		}
	} else if data[1]&0x80 != 0 {
		//for error we need to read 5 bytes
		if n < RtuExceptionSize {
			n1, err = io.ReadFull(sp.port, data[n:RtuExceptionSize])
		}
		n += n1
	}

	if err != nil {
		sp.Logger.Errorf("serial: read error %s", err.Error())
		return
	}
	aduResponse = data[:n]
	sp.Logger.Debugf("serial: received %x", aduResponse)
	return
}

// calculateDelay roughly calculates time needed for the next frame.
// See MODBUS over Serial Line - Specification and Implementation Guide (page 13).
func (sp *SerialPort) calculateDelay(chars int) time.Duration {
	var characterDelay, frameDelay int // us

	if sp.BaudRate <= 0 || sp.BaudRate > 19200 {
		characterDelay = 750
		frameDelay = 1750
	} else {
		characterDelay = 15000000 / sp.BaudRate
		frameDelay = 35000000 / sp.BaudRate
	}
	return time.Duration(characterDelay*chars+frameDelay) * time.Microsecond
}

func calculateResponseLength(adu []byte) int {
	length := RtuMinSize
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
