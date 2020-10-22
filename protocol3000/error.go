package protocol3000

import (
	"fmt"
)

// https://k.kramerav.com/downloads/protocols/protocol_3000_3.0_master_user.pdf, page 133
var errorText = map[int]string{
	1:  "Protocol Syntax",
	2:  "Command not available",
	3:  "Parameter out of range",
	4:  "Unauthorized access",
	5:  "Internal FW Error",
	6:  "Protocol busy",
	7:  "Wrong CRC",
	8:  "Timeout",
	9:  "(Reserved)",
	10: "Not enough space for data (firmware, FPGA...)",
	11: "Not enough space - file system",
	12: "File does not exist",
	13: "File can't be created",
	14: "File can't open",
	15: "(Reserved)",
	16: "(Reserved)",
	17: "(Reserved)",
	18: "(Reserved)",
	19: "(Reserved)",
	20: "(Reserved)",
	21: "Packet CRC error",
	22: "Packet number isn't expected (missing packet)",
	23: "Packet size is wrong",
	24: "(Reserved)",
	25: "(Reserved)",
	26: "(Reserved)",
	27: "(Reserved)",
	28: "(Reserved)",
	29: "(Reserved)",
	30: "EDID corrupted",
	31: "Device specific errors",
	32: "File has the same CRC - no changed",
	33: "Wrong operation mode",
	34: "Device/chip was not initalized",
}

func Error(code int) string {
	return fmt.Sprintf("%d: %s", code, errorText[code])
}
