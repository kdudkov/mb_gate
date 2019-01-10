package main

import (
	"io"
	"log"
	"net"
	"time"
)

const (
	IdleTimeout time.Duration = 5 * time.Second
)

type TcpHandler struct {
	conn         *net.Conn
	closeTimer   *time.Timer
	lastActivity *time.Time
}

func (app *App) ListenTCP(addressPort string) (err error) {
	listen, err := net.Listen("tcp", addressPort)
	if err != nil {
		log.Printf("Failed to Listen: %v\n", err)
		return err
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			log.Printf("Unable to accept connections: %#v\n", err)
			return err
		}

		go TcpHandler{conn: conn}.handle(app)
	}

	return err
}

func (h *TcpHandler) handle(app *App) {
	defer h.conn.Close()

	for {
		packet := make([]byte, tcpMaxLength)
		bytesRead, err := h.conn.Read(packet)
		if err != nil {
			if err != io.EOF {
				log.Printf("read error %v\n", err)
			}
			return
		}

		h.lastActivity = time.Now()
		h.startCloseTimer()

		transactionId, pdu, err := FromTCP(packet[:bytesRead])
		if err != nil {
			log.Printf("bad packet error %v\n", err)
			return
		}

		job := Job{Ch: make(chan bool)}
		job.TransactionId = transactionId
		job.Pdu = pdu

		app.Jobs <- job

		<-job.Ch
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
