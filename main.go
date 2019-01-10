package main

import (
	"fmt"
	"log"
	"mb_gate/modbus"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

type Job struct {
	Payload string
	Ch      chan bool
}

type App struct {
	Done       chan bool
	Jobs       chan *Job
	Logger     *log.Logger
	SerialPort *modbus.SerialPort
}

func NewApp() (app *App) {
	app = &App{
		Done:       make(chan bool),
		Jobs:       make(chan *Job, 10),
		Logger:     log.New(os.Stdout, "app: ", log.LstdFlags),
		SerialPort: NewSerial("/dev/ttyS1", 19200),
	}
	http.HandleFunc("/", app.handleIndex())
	return
}

func (app *App) logf(format string, v ...interface{}) {
	if app.Logger != nil {
		app.Logger.Printf(format, v...)
	}
}

func (app *App) StartWorker() {
	go func() {
		for {
			select {
			case job := <-app.Jobs:
				app.SerialPort.Send()
				app.logf("got job")
				time.Sleep(time.Second)
				job.Payload = "OK"
				job.Ch <- true
			case <-app.Done:
				return
			}

			runtime.Gosched()
		}
	}()

}

func (app *App) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job := Job{Ch: make(chan bool)}
		app.Jobs <- &job
		<-job.Ch
		close(job.Ch)

		fmt.Println(job.Payload)
	}
}

func (app *App) Run() {
	app.StartWorker()

	app.logf("running")
	err := http.ListenAndServe(":8081", nil)

	if err != nil {
		panic(err.Error())
	}
}

func main() {
	app := NewApp()

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		app.logf("exiting...")
		app.Done <- true
		os.Exit(1)
	}()

	app.Run()
}
