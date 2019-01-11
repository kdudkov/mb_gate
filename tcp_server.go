package main

import (
	"io"
	"mb_gate/modbus"
	"net"
	"time"
)

const (
	IdleTimeout time.Duration = 5 * time.Second
)

type TcpHandler struct {
	conn         net.Conn
	closeTimer   *time.Timer
	lastActivity time.Time
}

func (app *App) ListenTCP(addressPort string) (err error) {
	listen, err := net.Listen("tcp", addressPort)
	if err != nil {
		log.Errorf("Failed to Listen: %v", err)
		return err
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Errorf("Unable to accept connections: %#v", err)
			return err
		}

		h := TcpHandler{conn: conn}
		go h.handle(app)
	}
}

func (h *TcpHandler) handle(app *App) {
	defer h.conn.Close()

	for {
		packet := make([]byte, modbus.TcpHeaderSize)
		bytesRead, err := h.conn.Read(packet)
		if err != nil {
			if err != io.EOF {
				log.Errorf("read error %v", err)
			}
			return
		}

		h.lastActivity = time.Now()
		h.startCloseTimer()

		transactionId, pdu, err := modbus.FromTCP(packet[:bytesRead])
		if err != nil {
			log.Errorf("bad packet error %v", err)
			return
		}

		job := &Job{Ch: make(chan bool)}
		job.TransactionId = transactionId
		job.Pdu = pdu

		select {
		case app.Jobs <- job:
			<-job.Ch
		default:
			log.Error("buffer is full")
		}
		close(job.Ch)
	}
}

func (h *TcpHandler) startCloseTimer() {
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
		log.Printf("modbus: closing tcp connection due to idle timeout: %v", idle)
		h.conn.Close()
	}
}
