package main

import (
	"io"
	"mb_gate/modbus"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	IdleTimeout = 5 * time.Second
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
		packet := make([]byte, 255)
		bytesRead, err := h.conn.Read(packet)
		if err != nil {
			if err != io.EOF {
				log.Errorf("read error %v", err)
			}
			if h.closeTimer != nil {
				h.closeTimer.Stop()
			}
			return
		}
		h.setActivity()

		transactionId, pdu, err := modbus.FromTCP(packet[:bytesRead])
		if err != nil {
			log.Errorf("bad packet error %v", err)
			return
		}
		log.Debug(pdu)
		job := &Job{Ch: make(chan bool)}
		job.TransactionId = transactionId
		job.Pdu = pdu

		defer close(job.Ch)

		select {
		case app.Jobs <- job:
			<-job.Ch
			if _, err := h.conn.Write(job.Answer.ToTCP(transactionId)); err != nil {
				log.WithFields(logrus.Fields{"tr_id": job.TransactionId}).Error("error sending answer")
			}
			h.setActivity()
		default:
			log.WithFields(logrus.Fields{"tr_id": job.TransactionId}).Error("buffer is full")
			ans := modbus.NewModbusError(pdu, modbus.ExceptionCodeServerDeviceBusy)

			if _, err := h.conn.Write(ans.ToTCP(transactionId)); err != nil {
				log.WithFields(logrus.Fields{"tr_id": job.TransactionId}).Error("error sending answer")
			}
			h.setActivity()
		}
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
		log.Printf("modbus: closing tcp connection due to idle timeout: %v", idle)
		h.conn.Close()
	}
}
