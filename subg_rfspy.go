package cc111x

// Command represents a command used by subg_rfspy firmware running on CC111x.
// See https://github.com/ps2/subg_rfspy
type Command byte

//go:generate stringer -type Command

const (
	CmdGetState       Command = 1
	CmdGetVersion     Command = 2
	CmdGetPacket      Command = 3
	CmdSendPacket     Command = 4
	CmdSendAndListen  Command = 5
	CmdUpdateRegister Command = 6
	CmdReset          Command = 7
	CmdLED            Command = 8
	CmdReadRegister   Command = 9
)

// ErrorCode represents an error that can be returned by subg_rfspy firmware.
type ErrorCode byte

//go:generate stringer -type ErrorCode

const (
	ErrorRXTimeout      ErrorCode = 0xaa
	ErrorCmdInterrupted ErrorCode = 0xbb
	ErrorZeroData       ErrorCode = 0xcc
)
