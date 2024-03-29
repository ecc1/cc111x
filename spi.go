// +build !uart

package cc111x

import (
	"bytes"
	"log"
	"math/bits"
	"time"

	"github.com/ecc1/gpio"
	"github.com/ecc1/spi"
)

const (
	spiSpeed = 62500 // Hz

	verboseSPI = false
)

// Radio represents an open radio device.
type Radio struct {
	device        *spi.Device
	resetPin      gpio.OutputPin
	receiveBuffer bytes.Buffer
	snd           []byte
	rcv           []byte
	err           error
}

func openRadio() *Radio {
	r := &Radio{}
	r.device, r.err = spi.Open(spiDevice, spiSpeed, customCS)
	if r.err != nil {
		return r
	}
	r.resetPin, r.err = gpio.Output(resetPin, true, false)
	if r.err != nil {
		r.Close()
		return r
	}
	r.snd = make([]byte, 1)
	r.rcv = make([]byte, 1)
	return r
}

// Device returns the pathname of the radio's device.
func (*Radio) Device() string {
	return spiDevice
}

// Reset resets the CC111x hardware.
func (r *Radio) Reset() {
	if r.Error() != nil {
		return
	}
	if verbose {
		log.Printf("resetting CC111x")
	}
	r.resetPin.Write(true)
	time.Sleep(100 * time.Microsecond)
	r.err = r.resetPin.Write(false)
	time.Sleep(1 * time.Second)
}

var ()

func (r *Radio) xfer(b byte) byte {
	r.snd[0] = bits.Reverse8(b)
	r.err = r.device.Transfer(r.snd, r.rcv)
	c := bits.Reverse8(r.rcv[0])
	if verboseSPI {
		if r.err != nil {
			log.Printf("xfer %02X -> %02X (%v)", b, c, r.err)
		} else {
			log.Printf("xfer %02X -> %02X", b, c)
		}
	}
	return c
}

func (r *Radio) sendRequest(data []byte) {
	if verbose {
		log.Printf("request: % X", data)
	}
	n := len(data)
	if n > 0xFF {
		panic("request too long")
	}
	r.xfer(0x99)
	count := r.xfer(byte(n))
	for _, b := range data {
		rx := r.xfer(b)
		if count > 0 {
			r.err = r.receiveBuffer.WriteByte(rx)
			count--
		}
	}
	for count != 0 {
		rx := r.xfer(0)
		err := r.receiveBuffer.WriteByte(rx)
		if r.err == nil {
			r.err = err
		}
		count--
	}
}

func (r *Radio) request(cmd Command, params ...byte) {
	data := make([]byte, 1+len(params))
	data[0] = byte(cmd)
	copy(data[1:], params)
	r.sendRequest(data)
}

// Drain reads and discards any pending input from the subg_rfspy firmware.
func (r *Radio) drain() {
	r.xfer(0x99)
	count := r.xfer(0)
	for count != 0 {
		r.xfer(0)
		count--
	}
}

// Read any pending input from the subg_rfspy firmware into buf.
func (r *Radio) readResponse(buf *bytes.Buffer) {
	r.xfer(0x99)
	count := int(r.xfer(0))
	for i := 0; i < count; i++ {
		rx := r.xfer(0)
		err := buf.WriteByte(rx)
		if r.err == nil {
			r.err = err
		}
	}
	if verbose && count != 0 {
		n := buf.Len()
		data := buf.Bytes()
		if n != 0 {
			log.Printf("read %d bytes: % X", count, data[n-count:])
		}
	}
}

func (r *Radio) response(timeout time.Duration) []byte {
	const pollInterval = 1 * time.Millisecond
	buf := &r.receiveBuffer
	for timeout > 0 {
		r.readResponse(buf)
		b := buf.Bytes()
		i := bytes.LastIndexByte(b, 0)
		if i >= 0 {
			p := make([]byte, i)
			copy(p, b[:i])
			buf.Reset()
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
