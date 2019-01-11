package main

import (
	"flag"
	"github.com/sirupsen/logrus"
	"mb_gate/modbus"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var log = logrus.New()

type Job struct {
	TransactionId uint16
	Pdu           *modbus.ProtocolDataUnit
	Ch            chan bool
}

type App struct {
	Done       chan bool
	Jobs       chan *Job
	SerialPort *modbus.SerialPort
	httpPort   string
	tcpPort    string
}

func NewApp(port string, httpPort string, tcpPort string) (app *App) {
	app = &App{
		Done:       make(chan bool),
		Jobs:       make(chan *Job, 10),
		SerialPort: modbus.NewSerial(port, 19200),
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
				log.WithFields(logrus.Fields{"tr_id": job.TransactionId}).Info("got job")
				time.Sleep(time.Second)
				job.Ch <- true
			case <-app.Done:
				return
			}

			runtime.Gosched()
		}
	}()

}

func (app *App) Run() {

	flag.Parse()

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
	flag.Parse()

	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{})
	app := NewApp(*port, *httpPort, *tcpPort)

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
