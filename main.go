package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"go.uber.org/zap"

	"mb_gate/modbus"
)

const (
	readTimeout = time.Second * 5
)

var (
	gitRevision string = "unknown"
	gitBranch   string = "unknown"
)

type Job struct {
	TransactionId uint16
	Pdu           *modbus.ProtocolDataUnit
	Answer        *modbus.ProtocolDataUnit
	Ch            chan bool
}

type App struct {
	Done        chan bool
	Jobs        chan *Job
	SerialPort  *modbus.SerialPort
	httpPort    string
	tcpPort     string
	translators map[byte]Translator
	Logger      *zap.SugaredLogger
}

func NewApp(port string, portSpeed int, httpPort string, tcpPort string, logger *zap.SugaredLogger) (app *App) {
	app = &App{
		Done:        make(chan bool),
		Jobs:        make(chan *Job, 10),
		SerialPort:  modbus.NewSerial(port, portSpeed),
		httpPort:    httpPort,
		tcpPort:     tcpPort,
		translators: make(map[byte]Translator),
		Logger:      logger,
	}

	app.SerialPort.Logger = app.Logger.Named("serial")
	// addr 5
	app.translators[5] = NewSimpleChinese()
	// addr 100
	app.translators[100] = NewFakeTranslator()

	app.setRoute()
	return
}

func (app *App) StartWorker() {
	go func() {
		for {
			select {
			case job := <-app.Jobs:
				if job.Pdu == nil {
					app.Logger.Error("nil job pdu")
					continue
				}
				d, _ := job.Pdu.MakeRtu()
				ans, err := app.SerialPort.Send(d)
				if err != nil {
					app.Logger.Errorf("error %v", err, zap.Uint16("tr_id", job.TransactionId))
					job.Answer = modbus.NewModbusError(job.Pdu, modbus.ExceptionCodeServerDeviceFailure)
				} else {
					job.Answer, _ = modbus.FromRtu(ans)
					app.Logger.Debugf("answer %v", job.Answer, zap.Uint16("tr_id", job.TransactionId))
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

	app.Logger.Infof("start http server on %s", app.httpPort)

	go func() {
		if err := http.ListenAndServe(app.httpPort, nil); err != nil {
			app.Logger.Panic("can't start tcp listener", err)
		}
	}()

	app.Logger.Infof("start tcp server on %s", app.tcpPort)
	if err := app.ListenTCP(app.tcpPort); err != nil {
		app.Logger.Panic("can't start tcp listener", err)
	}
}

func (app *App) processPdu(transactionId uint16, pdu *modbus.ProtocolDataUnit) (*modbus.ProtocolDataUnit, error) {
	tr, ok := app.translators[pdu.SlaveId]
	if ok {
		dontSend := tr.Translate(pdu)
		if dontSend {
			return pdu, nil
		}
	}

	job := &Job{Ch: make(chan bool), TransactionId: transactionId, Pdu: pdu}
	defer close(job.Ch)

	select {
	case app.Jobs <- job:
		select {
		case <-job.Ch:
			return job.Answer, nil
		case <-time.After(readTimeout):
			ans := modbus.NewModbusError(pdu, modbus.ExceptionCodeServerDeviceFailure)
			return ans, fmt.Errorf("timeout")
		}
	default:
		ans := modbus.NewModbusError(pdu, modbus.ExceptionCodeServerDeviceBusy)
		return ans, fmt.Errorf("buffer is full")
	}
}

func main() {
	fmt.Printf("version %s:%s\n", gitBranch, gitRevision)

	var httpPort = flag.String("http", ":8080", "hpst:port for http")
	var tcpPort = flag.String("tcp", ":1502", "hpst:port for modbus tcp")
	var port = flag.String("port", "/dev/ttyS0", "serial port")
	var portSpeed = flag.Int("speed", 19200, "serial port speed")

	flag.Parse()

	cfg := zap.NewProductionConfig()
	logger, _ := cfg.Build()
	app := NewApp(*port, *portSpeed, *httpPort, *tcpPort, logger.Sugar())

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		app.Logger.Info("exiting...")
		app.Done <- true
		os.Exit(1)
	}()

	app.Run()
}
