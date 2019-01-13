package main

import (
	"flag"
	"mb_gate/modbus"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

type Job struct {
	TransactionId uint16
	Pdu           *modbus.ProtocolDataUnit
	Answer        *modbus.ProtocolDataUnit
	Ch            chan bool
}

type App struct {
	Done       chan bool
	Jobs       chan *Job
	SerialPort *modbus.SerialPort
	httpPort   string
	tcpPort    string
}

func NewApp(port string, portSpeed int, httpPort string, tcpPort string) (app *App) {
	app = &App{
		Done:       make(chan bool),
		Jobs:       make(chan *Job, 10),
		SerialPort: modbus.NewSerial(port, portSpeed),
		httpPort:   httpPort,
		tcpPort:    tcpPort,
	}
	http.HandleFunc("/", app.handleIndex())
	return
}

func (app *App) StartWorker() {
	go func() {
		for {
			select {
			case job := <-app.Jobs:
				if job.Pdu == nil {
					log.Error("nil job pdu")
					continue
				}
				d, _ := job.Pdu.MakeRtu()
				ans, err := app.SerialPort.Send(d)
				if err != nil {
					log.WithFields(logrus.Fields{"tr_id": job.TransactionId}).Errorf("error %v", err)
					job.Answer = modbus.NewModbusError(job.Pdu, modbus.ExceptionCodeServerDeviceFailure)
				} else {
					job.Answer, _ = modbus.FromRtu(ans)
					log.WithFields(logrus.Fields{"tr_id": job.TransactionId}).Debugf("answer %v", job.Answer)
				}
				job.Ch <- true
			case <-app.Done:
				return
			}

			runtime.Gosched()
		}
	}()

}

func (app *App) Run() {
	app.StartWorker()

	log.Infof("start http server on %s", app.httpPort)

	go func() {
		if err := http.ListenAndServe(app.httpPort, nil); err != nil {
			log.Panic("can't start tcp listener", err)
		}
	}()

	log.Infof("start tcp server on %s", app.tcpPort)
	if err := app.ListenTCP(app.tcpPort); err != nil {
		log.Panic("can't start tcp listener", err)
	}
}

func main() {
	var httpPort = flag.String("http", ":8080", "hpst:port for http")
	var tcpPort = flag.String("tcp", ":1502", "hpst:port for modbus tcp")
	var port = flag.String("port", "/dev/ttyS0", "serial port")
	var portSpeed = flag.Int("speed", 19200, "serial port speed")
	flag.Parse()

	logrus.SetLevel(logrus.DebugLevel)
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{})
	app := NewApp(*port, *portSpeed, *httpPort, *tcpPort)

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		log.Info("exiting...")
		app.Done <- true
		os.Exit(1)
	}()

	app.Run()
}
