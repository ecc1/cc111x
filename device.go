package cc111x

import (
	"bytes"
	"fmt"
	"log"
	"math/bits"
	"strings"
	"time"

	"github.com/ecc1/radio"
	"github.com/ecc1/spi"
)

const (
	firmwarePrefix = "subg_rfspy"
	spiSpeed       = 62500 // Hz

	verbose    = false
	verboseSPI = false
)

func init() {
	if verbose || verboseSPI {
		log.SetFlags(log.Ltime | log.Lmicroseconds | log.LUTC)
	}
}

// Radio represents an open radio device.
type Radio struct {
	device        *spi.Device
	receiveBuffer bytes.Buffer
	stats         radio.Statistics
	err           error
}

// Open opens the radio device.
func Open() *Radio {
	r := &Radio{}
	r.device, r.err = spi.Open(spiDevice, spiSpeed, 0)
	if r.err != nil {
		return r
	}
	r.Flush()
	v := r.Version()
	if !strings.HasPrefix(v, firmwarePrefix) {
		r.err = fmt.Errorf("unexpected firmware version %q", v)
	}
	return r
}

// Close closes the radio device.
func (r *Radio) Close() {
	r.err = r.device.Close()
}

// Name returns the radio's name.
func (r *Radio) Name() string {
	return "CC111x"
}

// Device returns the pathname of the radio's device.
func (r *Radio) Device() string {
	return spiDevice
}

// Version returns the radio's state.
func (r *Radio) State() string {
	r.request(CmdGetState)
	return string(r.response())
}

// Version returns the radio's firmware version.
func (r *Radio) Version() string {
	r.request(CmdGetVersion)
	return string(r.response())
}

func (r *Radio) Reset() {
	r.request(CmdReset)
}

// Init initializes the radio device.
func (r *Radio) Init(frequency uint32) {
	r.SetFrequency(frequency)
}

// Statistics returns the byte and packet counts for the radio device.
func (r *Radio) Statistics() radio.Statistics {
	return r.stats
}

// Error returns the error state of the radio device.
func (r *Radio) Error() error {
	return r.err
}

// SetError sets the error state of the radio device.
func (r *Radio) SetError(err error) {
	r.err = err
}

// Hardware returns the radio's hardware information.
func (r *Radio) Hardware() *radio.Hardware {
	panic("unimplemented")
}

var buf = make([]byte, 1)

func (r *Radio) xfer(b byte) byte {
	buf[0] = bits.Reverse8(b)
	r.err = r.device.Transfer(buf)
	c := bits.Reverse8(buf[0])
	if verboseSPI {
		log.Printf("xfer %02X -> %02X", b, c)
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
		r.err = r.receiveBuffer.WriteByte(rx)
		count--
	}
}

func (r *Radio) request(cmd Command, params ...byte) {
	data := make([]byte, 1+len(params))
	data[0] = byte(cmd)
	copy(data[1:], params)
	r.sendRequest(data)
}

func (r *Radio) Flush() {
	r.xfer(0x99)
	count := r.xfer(0)
	for count != 0 {
		r.xfer(0)
		count--
	}
}

// TODO: add timeout
func (r *Radio) response() []byte {
	for {
		b := r.receiveBuffer.Bytes()
		n := len(b)
		if n != 0 && b[n-1] == 0 {
			p := make([]byte, n-1)
			_, r.err = r.receiveBuffer.Read(p)
			r.receiveBuffer.Reset()
			if verbose {
				log.Printf("received %d-byte message %q", n-1, p)
			}
			return p
		}
		// Haven't received terminating 0 byte yet.
		time.Sleep(time.Millisecond)
		r.xfer(0x99)
		count := r.xfer(0)
		for count != 0 {
			rx := r.xfer(0)
			r.err = r.receiveBuffer.WriteByte(rx)
			count--
		}
	}
}
