/*
Author: Leonardo Rossi Leao
Createdt at: September 22nd, 2025
Last update: September 22nd, 2025
*/

package unicomm

import (
	"github.com/devicehub-go/unicomm/protocol/unicommserial"
	"github.com/devicehub-go/unicomm/protocol/unicommtcp"
)

type Protocol uint8

type UnicommOptions struct {
	Protocol Protocol
	Serial unicommserial.SerialOptions
	TCP unicommtcp.TCPOptions
	Delimiter string
}

type Unicomm interface {
	Connect() error
	Disconnect() error
	IsConnected() bool
	Read(size uint) ([]byte, error)
	ReadUntil(delimiter string) ([]byte, error)
	Write(message []byte) error
}

const (
	Serial Protocol = 0
	TCP    Protocol = 1
)

/*
Creates a new instance of unified communication based on
the target protocol
*/
func New(options UnicommOptions) Unicomm {
	switch(options.Protocol) {
	case Serial:
		return unicommserial.NewSerial(options.Serial)
	case TCP:
		return unicommtcp.NewTCP(options.TCP)
	}
	return nil
}