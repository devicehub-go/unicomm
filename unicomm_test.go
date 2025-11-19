package unicomm_test

import (
	"log"
	"testing"
	"time"

	"github.com/devicehub-go/unicomm"
	"github.com/devicehub-go/unicomm/protocol/unicommserial"
)

func TestSerial(t *testing.T) {
	options := unicomm.Options{
		Protocol: unicomm.Serial,
		Serial: unicommserial.SerialOptions{
			PortName:     "COM6",
			BaudRate:     115200,
			DataBits:     8,
			StopBits:     unicommserial.OneStopBit,
			Parity:       unicommserial.NoParity,
			ReadTimeout:  time.Second * 5,
			WriteTimeout: time.Second * 5,
		},
	}
	serial := unicomm.New(options)
	if err := serial.Connect(); err != nil {
		log.Fatal(err)
	}
}