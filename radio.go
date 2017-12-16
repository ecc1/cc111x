package cc111x

import (
	"fmt"
	"time"
)

const (
	FXOSC = 24000000 // Crystal frequency in Hz

	FREQ2 = 0x09 // Frequency control word, high byte
	FREQ1 = 0x0A // Frequency control word, middle byte
	FREQ0 = 0x0B // Frequency control word, low byte
)

func (r *Radio) ReadRegister(addr byte) byte {
	r.request(CmdReadRegister, addr)
	b := r.response(defaultTimeout)
	return b[0]
}

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
	// Terminate packet with 2 zero bytes.
	params := make([]byte, 3+len(data)+2)
	params[0] = 0 // channel
	params[1] = 1 // repeat
	params[2] = 0 // delay
	copy(params[3:], data)
	r.request(CmdSendPacket, params...)
	_ = r.response(defaultTimeout)
	if r.Error() == nil {
		r.stats.Packets.Sent++
		r.stats.Bytes.Sent += len(data)
	}
}

// Receive listens with the given timeout for an incoming packet.
// It returns the packet and the associated RSSI.
func (r *Radio) Receive(timeout time.Duration) ([]byte, int) {
	const rssiOffset = 73 // see data sheet section 13.10.3, table 68
	if r.Error() != nil {
		return nil, 0
	}
	channel := 0
	ms := uint32(timeout / time.Millisecond)
	params := append([]byte{byte(channel)}, marshalUint32(ms)...)
	r.request(CmdGetPacket, params...)
	data := r.response(timeout)
	if len(data) == 0 {
		return nil, 0
	}
	if len(data) <= 2 {
		r.SetError(fmt.Errorf("Receive: %v", ErrorCode(data[0])))
		return nil, 0
	}
	r.stats.Packets.Received++
	r.stats.Bytes.Received += len(data)
	rssi := int(data[0])
	if rssi >= 128 {
		rssi -= 256
	}
	rssi = rssi/2 - rssiOffset
	return data[2:], rssi
}
