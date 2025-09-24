# Unicomm

A Go library that provides an unified interface to communicate with devices over multiple protocols, including Serial and Ethernet, allowing developers to use the same methods regardless of the underlying transport.

## Features

- **Unified Interface**: Single API for both Serial and TCP communication
- **Protocol Abstraction**: Switch between protocols without changing your application logic
- **Thread-Safe**: Built-in mutex protection for concurrent operations
- **Configurable Timeouts**: Separate read/write timeouts for fine-tuned control
- **Connection Management**: Connection handling with status checking

## Installation

```bash
go get github.com/devicehub-go/unicomm
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "time"
    
    "github.com/devicehub-go/unicomm"
    "github.com/devicehub-go/unicomm/protocol/unicommserial"
    "github.com/devicehub-go/unicomm/protocol/unicommtcp"
    "go.bug.st/serial"
)

func main() {
    // Create a new Unicomm instance for Serial communication
    comm := unicomm.New(unicomm.UnicommOptions{
        Protocol: unicomm.Serial,
        Serial: unicommserial.SerialOptions{
            PortName:       "/dev/ttyUSB0",
            BaudRate:       9600,
            Parity:         serial.NoParity,
            DataBits:       8,
            StopBits:       serial.OneStopBit,
            ReadTimeout:    time.Second,
            WriteTimeout:   time.Second,
            EndDelimiter:   "\r\n",
        },
    })

    // Connect
    if err := comm.Connect(); err != nil {
        panic(err)
    }
    defer comm.Disconnect()

    // Write data
    err := comm.Write([]byte("Hello Device"))
    if err != nil {
        fmt.Printf("Write error: %v\n", err)
        return
    }

    // Read response
    response, err := comm.ReadUntil("\r\n")
    if err != nil {
        fmt.Printf("Read error: %v\n", err)
        return
    }

    fmt.Printf("Response: %s\n", string(response))
}
```

### Serial Communication

```go
comm := unicomm.New(unicomm.UnicommOptions{
    Protocol: unicomm.Serial,
    Serial: unicommserial.SerialOptions{
        PortName:        "/dev/ttyUSB0",  // Windows: "COM1"
        BaudRate:        115200,
        Parity:          serial.NoParity,
        DataBits:        8,
        StopBits:        serial.OneStopBit,
        ReadTimeout:     500 * time.Millisecond,
        WriteTimeout:    500 * time.Millisecond,
        StartDelimiter:  "",
        EndDelimiter:    "\n",
        RetryConnect:    true,
    },
})
```

### TCP Communication

```go
comm := unicomm.New(unicomm.UnicommOptions{
    Protocol: unicomm.TCP,
    TCP: unicommtcp.TCPOptions{
        Host:         "192.168.1.100",
        Port:         8080,
        ReadTimeout:  time.Second,
        WriteTimeout: time.Second,
        EndDelimiter: "\r\n",
    },
})
```

### IPv6 Support

```go
comm := unicomm.New(unicomm.UnicommOptions{
    Protocol: unicomm.TCP,
    TCP: unicommtcp.TCPOptions{
        Host:         "2001:db8::1",  // IPv6 address
        Port:         8080,
        ReadTimeout:  time.Second,
        WriteTimeout: time.Second,
        EndDelimiter: "\r\n",
    },
})
```

## API Reference

### Unicomm Interface

All communication protocols implement the `Unicomm` interface:

```go
type Unicomm interface {
    Connect() error
    Disconnect() error
    IsConnected() bool
    Read(size uint) ([]byte, error)
    ReadUntil(delimiter string) ([]byte, error)
    Write(message []byte) error
}
```

#### Methods

- **`Connect()`**: Establishes connection to the target device/server
- **`Disconnect()`**: Closes the current connection
- **`IsConnected()`**: Returns `true` if connection is active
- **`Read(size uint)`**: Reads a specific number of bytes
- **`ReadUntil(delimiter string)`**: Reads data until delimiter is found
- **`Write(message []byte)`**: Writes data to the connection

### Configuration Options

#### SerialOptions

```go
type SerialOptions struct {
    PortName       string        // Port name (e.g., "/dev/ttyUSB0", "COM1")
    BaudRate       int           // Baud rate (e.g., 9600, 115200)
    Parity         Parity        // Parity setting
    DataBits       int           // Data bits (5, 6, 7, 8)
    StopBits       StopBits      // Stop bits
    ReadTimeout    time.Duration // Read operation timeout
    WriteTimeout   time.Duration // Write operation timeout
    StartDelimiter string        // Message start delimiter
    EndDelimiter   string        // Message end delimiter
    RetryConnect   bool          // Enable connection retry
}
```

#### TCPOptions

```go
type TCPOptions struct {
    Host         string        // Target host (IP address or hostname)
    Port         uint          // Target port
    ReadTimeout  time.Duration // Read operation timeout
    WriteTimeout time.Duration // Write operation timeout
    EndDelimiter string        // Message end delimiter
}
```

## Examples

### Reading Fixed-Size Data

```go
// Read exactly 10 bytes
data, err := comm.Read(10)
if err != nil {
    log.Printf("Read error: %v", err)
    return
}
fmt.Printf("Received: %s\n", string(data))
```

### Reading Until Delimiter

```go
// Read until newline character
response, err := comm.ReadUntil("\n")
if err != nil {
    log.Printf("Read error: %v", err)
    return
}
fmt.Printf("Response: %s\n", string(response))
```

### Connection Status Checking

```go
if comm.IsConnected() {
    fmt.Println("Connection is active")
} else {
    fmt.Println("Connection is closed")
    if err := comm.Connect(); err != nil {
        log.Printf("Reconnection failed: %v", err)
    }
}
```

## Thread Safety

Unicomm is thread-safe. All operations are protected by internal mutexes, making it safe to use from multiple goroutines simultaneously.

```go
// Safe to call from multiple goroutines
go func() {
    comm.Write([]byte("message1"))
}()

go func() {
    data, _ := comm.Read(10)
    fmt.Println(string(data))
}()
```

## Default Values

- **ReadTimeout**: 100ms
- **WriteTimeout**: 100ms
- **TCP Connection Timeout**: 500ms

## Error Handling

Common errors and their meanings:

- `"there is no port connected"`: Connection is not established
- `"port is not available"`: Serial port doesn't exist or is in use
- `"read until timeout"`: Read operation exceeded timeout
- `"there is a connection already established"`: Attempting to connect when already connected

## License

This project is authored by Leonardo Rossi Leao and was created on September 22nd, 2025.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
