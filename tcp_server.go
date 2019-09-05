package main

import (
	"io"
	"net"
	"time"

	"github.com/kdudkov/mb_gate/modbus"
)

const (
	IdleTimeout = 10 * time.Second
)

type TcpHandler struct {
	conn         net.Conn
	closeTimer   *time.Timer
	lastActivity time.Time
	app          *App
}

func (app *App) ListenTCP(addressPort string) (err error) {
	listen, err := net.Listen("tcp", addressPort)
	if err != nil {
		app.Logger.Errorf("Failed to Listen: %v", err)
		return err
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			app.Logger.Errorf("Unable to accept connections: %#v", err)
			return err
		}

		h := TcpHandler{conn: conn, app: app}
		go h.handle()
	}
}

func (h *TcpHandler) handle() {
	defer h.conn.Close()

	for {
		packet := make([]byte, 255)
		bytesRead, err := h.conn.Read(packet)
		if err != nil {
			if err != io.EOF {
				h.app.Logger.Errorf("read error %v", err)
			}
			if h.closeTimer != nil {
				h.closeTimer.Stop()
			}
			return
		}
		h.setActivity()

		transactionId, pdu, err := modbus.FromTCP(packet[:bytesRead])
		l := h.app.Logger.With("tr_id", transactionId)

		if err != nil {
			h.app.Logger.Errorf("bad packet error %v", err)
			return
		}
		l.Debugf("request: %v", pdu)

		ans, err := h.app.processPdu(transactionId, pdu)
		if err != nil {
			l.Errorf("error processing pdu: %s", err.Error())
		}

		if ans == nil {
			ans = modbus.NewModbusError(pdu, modbus.ExceptionCodeServerDeviceBusy)
		}
		l.Debugf("answer: %v", ans)

		if _, err := h.conn.Write(ans.MakeTCP(transactionId)); err != nil {
			l.Error("error sending answer")
		}
		h.setActivity()
	}
}

func (h *TcpHandler) setActivity() {
	h.lastActivity = time.Now()

	if h.closeTimer == nil {
		h.closeTimer = time.AfterFunc(IdleTimeout, h.closeIdle)
	} else {
		h.closeTimer.Reset(IdleTimeout)
	}
}

// closeIdle closes the connection if last activity is passed behind IdleTimeout.
func (h *TcpHandler) closeIdle() {
	idle := time.Now().Sub(h.lastActivity)

	if idle >= IdleTimeout {
		h.app.Logger.Debugf("modbus: closing tcp connection due to idle timeout: %v", idle)
		h.conn.Close()
	}
}
