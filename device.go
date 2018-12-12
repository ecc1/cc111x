package cc111x

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

const (
	firmwarePrefix = "subg_rfspy"

	defaultTimeout = 50 * time.Millisecond

	verbose = false
)

var (
	errNoResponse = errors.New("cc111x: no response")
)

func init() {
	if verbose {
		log.SetFlags(log.Ltime | log.Lmicroseconds | log.LUTC)
	}
}

// Open opens the radio device.
func Open() *Radio {
	r := openRadio()
	if r.err != nil {
		return r
	}
	r.Reset()
	v := r.Version()
	if r.err != nil {
		r.Close()
		return r
	}
	if !strings.HasPrefix(v, firmwarePrefix) {
		r.err = fmt.Errorf("unexpected firmware version %q", v)
	}
	if r.err != nil {
		r.Close()
	}
	return r
}

// Close closes the radio device.
func (r *Radio) Close() {
	_ = r.device.Close()
}

// Name returns the radio's name.
func (r *Radio) Name() string {
	return "CC111x"
}

// State returns the radio's state.
func (r *Radio) State() string {
	return r.stringOp(CmdGetState)
}

// Version returns the radio's firmware version.
func (r *Radio) Version() string {
	return r.stringOp(CmdGetVersion)
}

func (r *Radio) stringOp(cmd Command) string {
	r.request(cmd)
	return string(r.response(defaultTimeout))
}

// Init initializes the radio device.
func (r *Radio) Init(frequency uint32) {
	r.SetFrequency(frequency)
}

// Error returns the error state of the radio device.
func (r *Radio) Error() error {
	return r.err
}

// SetError sets the error state of the radio device.
func (r *Radio) SetError(err error) {
	r.err = err
}
