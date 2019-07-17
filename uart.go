// +build uart

package cc111x

import (
	"bytes"
	"log"
	"time"

	"github.com/ecc1/serial"
)

const (
	serialDevice = "/dev/serial0"
	serialSpeed  = 19200
)

// Radio represents an open radio device.
type Radio struct {
	device *serial.Port
	err    error
}

func openRadio() *Radio {
	r := &Radio{}
	r.device, r.err = serial.Open(serialDevice, serialSpeed)
	return r
}

// Device returns the pathname of the radio's device.
func (*Radio) Device() string {
	return serialDevice
}

// Reset resets the CC111x hardware.
// (Can't reset when connected only via serial port.)
func (*Radio) Reset() {
}

func (r *Radio) request(cmd Command, params ...byte) {
	data := make([]byte, 1+len(params))
	data[0] = byte(cmd)
	copy(data[1:], params)
	if verbose {
		log.Printf("request: % X", data)
	}
	r.err = r.device.Write(data)
}

func (r *Radio) response(timeout time.Duration) []byte {
	const pollInterval = 1 * time.Millisecond
	buf := make([]byte, 256)
	off := 0
	for timeout > 0 {
		n, err := r.device.ReadAvailable(buf[off:])
		if err != nil {
			r.SetError(err)
			return nil
		}
		off += n
		i := bytes.LastIndexByte(buf[:off], 0)
		if i >= 0 {
			p := buf[:i]
			if verbose {
				log.Printf("received %d-byte response % X", i, p)
			}
			return p
		}
		// No terminating 0 byte; wait for more data.
		time.Sleep(pollInterval)
		timeout -= pollInterval
	}
	if verbose {
		log.Printf("no response")
	}
	r.SetError(errNoResponse)
	return nil
}
