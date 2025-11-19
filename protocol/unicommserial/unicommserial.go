/*
Author: Leonardo Rossi Leao
Created at: September 22nd, 2025
Last update: September 23rd, 2025
*/

package unicommserial

import (
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"go.bug.st/serial"
)

type Port = serial.Port
type Parity = serial.Parity
type StopBits = serial.StopBits

type SerialOptions struct {
	PortName string
	BaudRate int
	Parity Parity
	DataBits int
	StopBits StopBits
	ReadTimeout time.Duration
	WriteTimeout time.Duration
	StartDelimiter string
	EndDelimiter string
	RetryConnect bool
}

type UnicommSerial struct {
	Options SerialOptions
	Connection Port

	mutex sync.Mutex  // Protect port instance
}

const (
	NoParity Parity = iota
	OddParity
	EvenParity
	MarkParity
	SpaceParity
)

const (
	OneStopBit StopBits = iota
	OnePointFiveStopBits
	TwoStopBits
)

/*
Creates a new instance of Unicomm Serial communication
*/
func NewSerial(options SerialOptions) *UnicommSerial {
	if options.ReadTimeout == 0 {
		options.ReadTimeout = 100 * time.Millisecond
	}
	if options.WriteTimeout == 0 {
		options.WriteTimeout = 100 * time.Millisecond
	}
	return &UnicommSerial{
		Options: options,
	}
}

/*
Returns true if the target port is available
*/
func (us *UnicommSerial) isPortAvailable(portName string) error {
	available, err := serial.GetPortsList()
	if err != nil {
		return fmt.Errorf("was not possible to validate the port")
	}
	if !slices.Contains(available, portName) {
		return fmt.Errorf("port is not available")
	}

	port, err := serial.Open(portName, &serial.Mode{})
	port.Close()

	return err
}

/*
Returns true if the serial port is connected
*/
func (us *UnicommSerial) IsConnected() bool {
	if us.Connection == nil {
		return false
	}

	us.mutex.Lock()
	defer us.mutex.Unlock()

	if _, err := us.Connection.Write([]byte{}); err != nil {
		return false
	}
	return true
}

/*
Establishes a connection with the desired serial port
*/
func (us *UnicommSerial) Connect() error {
	portName := us.Options.PortName
	serialMode := &serial.Mode{
		BaudRate: us.Options.BaudRate,
		Parity: us.Options.Parity,
		DataBits: us.Options.DataBits,
		StopBits: us.Options.StopBits,
	}

	if us.IsConnected() {
		return fmt.Errorf("there is a port already connected")
	}
	if err := us.isPortAvailable(portName); err != nil {
		return err
	}

	us.mutex.Lock()
	defer us.mutex.Unlock()

	port, err := serial.Open(portName, serialMode)
	if err != nil {
		us.Connection = nil
		return err
	}
	port.SetReadTimeout(us.Options.ReadTimeout)

	us.Connection = port
	return nil
}

/*
Closes the connection with the current serial port
*/
func (us *UnicommSerial) Disconnect() error {
	if !us.IsConnected() {
		return fmt.Errorf("there is no port connected")
	}
	
	us.mutex.Lock()
	defer us.mutex.Unlock()

	if err := us.Connection.Close(); err != nil {
		return err
	}
	
	us.Connection = nil
	return nil
}

/*
Reads a number of bytes from the serial port
*/
func (us *UnicommSerial) Read(n uint) ([]byte, error) {
	buffer := make([]byte, n)

	if !us.IsConnected() {
		return nil, fmt.Errorf("there is no port connected")
	}

	us.mutex.Lock()
	nReaded, err := us.Connection.Read(buffer)
	us.mutex.Unlock()

	if err != nil {
		return nil, err
	}
	return buffer[:nReaded], nil
}

/*
Reads data from the serial port until a target
delimiter is found
*/
func (us *UnicommSerial) ReadUntil(endDelimiter string) ([]byte, error) {
	buffer := make([]byte, 0)
	singleByte := make([]byte, 1)
	errorChan := make(chan error, 1)
	resultChan := make(chan []byte, 1)

	if !us.IsConnected() {
		return nil, fmt.Errorf("there is no port connected")
	}

	us.mutex.Lock()
	defer us.mutex.Unlock()

	go func() {
		for {
			nReaded, err := us.Connection.Read(singleByte)
			if err != nil {
				errorChan <- err
				return
			}
			if nReaded == 1 {
				buffer = append(buffer, singleByte...)
			} 
			if nReaded == 0 || strings.Contains(string(buffer), endDelimiter) {
				resultChan <- buffer
				return
			}
		}
	}()
	
	select {
		case err := <-errorChan:
			return nil, err
		case result := <-resultChan:
			return result, nil
		case <- time.After(us.Options.ReadTimeout):
			return buffer, fmt.Errorf("read until timeout")
	}
}

/*
Writes an array of bytes to the serial port
*/
func (us *UnicommSerial) Write(message []byte) error {
	var errorChan = make(chan error)

	msgStr := string(message)
	endDelimiter := us.Options.EndDelimiter

	if !us.IsConnected() {
		return fmt.Errorf("there is no port connected")
	}
	if endDelimiter != "" && !strings.HasSuffix(msgStr, endDelimiter) {
		message = append(message, []byte(endDelimiter)...)
	}

	timer := time.NewTimer(us.Options.WriteTimeout)
	defer timer.Stop()

	us.mutex.Lock()
	defer us.mutex.Unlock()

	go func() {
		us.Connection.ResetInputBuffer()
		us.Connection.ResetOutputBuffer()
		nWrited, err := us.Connection.Write(message)
		if nWrited != len(message) {
			errorChan <-fmt.Errorf("writed %d bytes, expected %d", nWrited, len(message))
			return
		}
		errorChan <-err
	}()

	for{
		select{
		case err := <-errorChan:
			return err
		case <-timer.C:
			return fmt.Errorf("write timeout")
		}
	}
}