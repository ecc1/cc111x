// Code generated by "stringer -type Command"; DO NOT EDIT.

package cc111x

import "strconv"

const _Command_name = "CmdGetStateCmdGetVersionCmdGetPacketCmdSendPacketCmdSendAndListenCmdUpdateRegisterCmdResetCmdLEDCmdReadRegister"

var _Command_index = [...]uint8{0, 11, 24, 36, 49, 65, 82, 90, 96, 111}

func (i Command) String() string {
	i -= 1
	if i >= Command(len(_Command_index)-1) {
		return "Command(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _Command_name[_Command_index[i]:_Command_index[i+1]]
}
