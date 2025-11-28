/*
Author: Leonardo Rossi Leao
Created at: September 23rd, 2025
Last update: September 23rd, 2025
*/

package unicommtcp

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type TCPOptions struct {
	Host         string
	Port         uint
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	EndDelimiter string
}

type UnicommTCP struct {
	Options    TCPOptions
	Connection net.Conn

	mutex sync.Mutex
}

/*
Creates a new instance of Unicomm TCP communication
*/
func NewTCP(options TCPOptions) *UnicommTCP {
	if options.ReadTimeout == 0 {
		options.ReadTimeout = 100 * time.Millisecond
	}
	if options.WriteTimeout == 0 {
		options.WriteTimeout = 100 * time.Millisecond
	}
	return &UnicommTCP{
		Options: options,
	}
}

/*
Returns true if TCP connection is established
*/
func (ut *UnicommTCP) IsConnected() bool {
	if ut.Connection == nil {
		return false
	}

	ut.mutex.Lock()
	defer ut.mutex.Unlock()

	if _, err := ut.Connection.Write([]byte{}); err != nil {
		return false
	}
	return true
}

/*
Establishes a connection with the desired host and port
*/
func (ut *UnicommTCP) Connect() error {
	var url string
	host := ut.Options.Host
	port := ut.Options.Port

	if ut.IsConnected() {
		return fmt.Errorf("there is a connection already established")
	}

	if strings.Contains(host, ":") {
		url = fmt.Sprintf("[%s]:%d", host, port)
	} else {
		url = fmt.Sprintf("%s:%d", host, port)
	}

	ut.mutex.Lock()
	defer ut.mutex.Unlock()

	connection, err := net.DialTimeout("tcp", url, 500*time.Millisecond)
	if err != nil {
		ut.Connection = nil
		return err
	}

	ut.Connection = connection
	return nil
}

/*
Closes the connection with the current TCP server
*/
func (ut *UnicommTCP) Disconnect() error {
	if !ut.IsConnected() {
		return fmt.Errorf("there is no connection established")
	}

	ut.mutex.Lock()
	defer ut.mutex.Unlock()

	if err := ut.Connection.Close(); err != nil {
		return err
	}

	ut.Connection = nil
	return nil
}

/*
Reads a number of bytes from the TCP server
*/
func (ut *UnicommTCP) Read(n uint) ([]byte, error) {
	buffer := make([]byte, n)

	if !ut.IsConnected() {
		return nil, fmt.Errorf("there is no port connected")
	}

	ut.mutex.Lock()
	timeout := time.Now().Add(ut.Options.ReadTimeout)
	ut.Connection.SetReadDeadline(timeout)
	nReaded, err := ut.Connection.Read(buffer)
	ut.mutex.Unlock()

	if err != nil {
		return nil, err
	}
	return buffer[:nReaded], nil
}

/*
Reads data from the TCP server until a target
delimiter is found
*/
func (ut *UnicommTCP) ReadUntil(endDelimiter string) ([]byte, error) {
	buffer := make([]byte, 0)
	singleByte := make([]byte, 1)
	errorChan := make(chan error, 1)
	resultChan := make(chan []byte, 1)

	if !ut.IsConnected() {
		return nil, fmt.Errorf("there is no port connected")
	}

	ut.mutex.Lock()
	defer ut.mutex.Unlock()

	timeout := time.Now().Add(ut.Options.ReadTimeout)
	ut.Connection.SetReadDeadline(timeout)

	go func() {
		for {
			nReaded, err := ut.Connection.Read(singleByte)
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
	case <-time.After(ut.Options.ReadTimeout):
		return buffer, fmt.Errorf("read until timeout")
	}
}

/*
Writes an array of bytes to the serial port
*/
func (ut *UnicommTCP) Write(message []byte) error {
	msgStr := string(message)
	endDelimiter := ut.Options.EndDelimiter

	if !ut.IsConnected() {
		return fmt.Errorf("there is no port connected")
	}
	if endDelimiter != "" && !strings.HasSuffix(msgStr, endDelimiter) {
		message = append(message, []byte(endDelimiter)...)
	}

	ut.mutex.Lock()
	timeout := time.Now().Add(ut.Options.WriteTimeout)
	ut.Connection.SetReadDeadline(timeout)
	nWrited, err := ut.Connection.Write(message)
	ut.mutex.Unlock()

	if err != nil {
		return err
	}
	if nWrited != len(message) {
		return fmt.Errorf("writed %d bytes, expected %d", nWrited, len(message))
	}
	return nil
}
