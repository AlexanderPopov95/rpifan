package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/stianeikeland/go-rpio"
)

const (
	pin      = 2
	lowTemp  = 42
	highTemp = 55

	sensorPath = "/sys/devices/virtual/thermal/thermal_zone0/temp"
)

var fanEnabled = false

func main() {
	runtime.GOMAXPROCS(1)
	term := make(chan os.Signal)
	signal.Notify(term, syscall.SIGINT, syscall.SIGTERM)

	if err := rpio.Open(); err != nil {
		panic(err)
	}
	pin := rpio.Pin(pin)
	pin.Output()

	defer func() {
		pin.Low()
		rpio.Close()
	}()

	go tempControl(pin)
	<-term
}

func tempControl(pin rpio.Pin) {
	for {
		<-time.After(time.Second)
		t := temp()
		if t > highTemp && !fanEnabled {
			pin.High()
			fanEnabled = true
			continue
		}
		if t < lowTemp && fanEnabled {
			pin.Low()
			fanEnabled = false
		}
	}
}

func temp() int {
	b, err := ioutil.ReadFile(sensorPath)
	if err != nil {
		fmt.Println(err)
		return highTemp + 1
	}
	f, err := strconv.Atoi(string(b[:len(b)-1]))
	if err != nil {
		fmt.Println(err)
		return highTemp + 1
	}
	return f / 1000
}
