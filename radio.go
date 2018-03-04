package cc111x

import (
	"fmt"
	"time"
)

// CC111x hardware-related constants.
const (
	FXOSC = 24000000 // Crystal frequency in Hz

	FREQ2 = 0x09 // Frequency control word, high byte
	FREQ1 = 0x0A // Frequency control word, middle byte
	FREQ0 = 0x0B // Frequency control word, low byte
)

// ReadRegister returns the value of a CC111x register.
// This is only available on subg_rfspy 1.0 or later.
func (r *Radio) ReadRegister(addr byte) byte {
	r.request(CmdReadRegister, addr)
	b := r.response(defaultTimeout)
	if len(b) == 0 {
		return 0
	}
	return b[0]
}

// WriteRegister writes a value to a CC111x register.
func (r *Radio) WriteRegister(addr byte, value byte) {
	r.request(CmdUpdateRegister, addr, value)
	_ = r.response(defaultTimeout)
}

// Frequency returns the radio's current frequency, in Hertz.
func (r *Radio) Frequency() uint32 {
	f2 := r.ReadRegister(FREQ2)
	f1 := r.ReadRegister(FREQ1)
	f0 := r.ReadRegister(FREQ0)
	f := uint32(f2)<<16 + uint32(f1)<<8 + uint32(f0)
	return uint32(uint64(f) * FXOSC >> 16)
}

// SetFrequency sets the radio to the given frequency, in Hertz.
func (r *Radio) SetFrequency(freq uint32) {
	f := (uint64(freq)<<16 + FXOSC/2) / FXOSC
	r.WriteRegister(FREQ2, byte(f>>16))
	r.WriteRegister(FREQ1, byte(f>>8))
	r.WriteRegister(FREQ0, byte(f))
}

// Send transmits the given packet.
func (r *Radio) Send(data []byte) {
	if r.Error() != nil {
		return
	}
	// Terminate packet with a zero byte.
	params := make([]byte, 3+len(data)+1)
	params[0] = 0 // channel
	params[1] = 0 // repeat
	params[2] = 0 // delay
	copy(params[3:], data)
	r.request(CmdSendPacket, params...)
	timeout := time.Duration(len(params)) * time.Millisecond
	if timeout < defaultTimeout {
		timeout = defaultTimeout
	}
	_ = r.response(timeout)
}

// Receive listens with the given timeout for an incoming packet.
// It returns the packet and the associated RSSI.
func (r *Radio) Receive(timeout time.Duration) ([]byte, int) {
	if r.Error() != nil {
		return nil, 0
	}
	channel := 0
	ms := uint32(timeout / time.Millisecond)
	params := append([]byte{byte(channel)}, marshalUint32(ms)...)
	r.request(CmdGetPacket, params...)
	return r.finishReceive(timeout)
}

// SendAndReceive sends the given packet,
// then listens with the given timeout for an incoming packet.
// It returns the packet and the associated RSSI.
func (r *Radio) SendAndReceive(p []byte, timeout time.Duration) ([]byte, int) {
	if r.Error() != nil {
		return nil, 0
	}
	ms := uint32(timeout / time.Millisecond)
	// Terminate packet with a zero byte.
	params := make([]byte, 9+len(p)+1)
	params[0] = 0 // send_channel
	params[1] = 0 // repeat_count
	params[2] = 0 // delay_ms
	params[3] = 0 // listen_channel
	copy(params[4:8], marshalUint32(ms))
	params[8] = 0 // retry_count
	copy(params[9:], p)
	r.request(CmdSendAndListen, params...)
	return r.finishReceive(timeout)
}

const rssiOffset = 73 // see data sheet section 13.10.3, table 68

func (r *Radio) finishReceive(timeout time.Duration) ([]byte, int) {
	var data []byte
	for r.Error() == nil {
		// Wait a little longer than the firmware does.
		data = r.response(timeout + 5*time.Millisecond)
		if len(data) == 0 {
			break
		}
		if len(data) == 1 {
			code := ErrorCode(data[0])
			switch code {
			case ErrorRXTimeout:
			case ErrorCmdInterrupted:
				continue
			default:
				r.SetError(fmt.Errorf("Receive: %v", code))
			}
			break
		}
		rssi := int(data[0])
		if rssi >= 128 {
			rssi -= 256
		}
		rssi = rssi/2 - rssiOffset
		return data[2:], rssi
	}
	return nil, -128
}
