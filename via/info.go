package via

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

type command struct {
	XMLName  xml.Name `xml:"P"`
	Username string   `xml:"UN"`
	Password string   `xml:"Pwd"`
	Command  string   `xml:"Cmd"`
	Param1   string   `xml:"P1,omitempty"`
	Param2   string   `xml:"P2,omitempty"`
	Param3   string   `xml:"P3,omitempty"`
	Param4   string   `xml:"P4,omitempty"`
	Param5   string   `xml:"P5,omitempty"`
	Param6   string   `xml:"P6,omitempty"`
	Param7   string   `xml:"P7,omitempty"`
	Param8   string   `xml:"P8,omitempty"`
	Param9   string   `xml:"P9,omitempty"`
	Param10  string   `xml:"P10,omitempty"`
}

// HardwareInfo contains the common information for device hardware information
type HardwareInfo struct {
	Hostname              string           `json:"hostname,omitempty"`
	ModelName             string           `json:"model_name,omitempty"`
	SerialNumber          string           `json:"serial_number,omitempty"`
	BuildDate             string           `json:"build_date,omitempty"`
	FirmwareVersion       string           `json:"firmware_version,omitempty"`
	ProtocolVersion       string           `json:"protocol_version,omitempty"`
	NetworkInfo           NetworkInfo      `json:"network_information,omitempty"`
	FilterStatus          string           `json:"filter_status,omitempty"`
	WarningStatus         []string         `json:"warning_status,omitempty"`
	ErrorStatus           []string         `json:"error_status,omitempty"`
	PowerStatus           string           `json:"power_status,omitempty"`
	PowerSavingModeStatus string           `json:"power_saving_mode_status,omitempty"`
	TimerInfo             []map[string]int `json:"timer_info,omitempty"`
	Temperature           string           `json:"temperature,omitempty"`
}

// NetworkInfo contains the network information for the device
type NetworkInfo struct {
	IPAddress  string   `json:"ip_address,omitempty"`
	MACAddress string   `json:"mac_address,omitempty"`
	Gateway    string   `json:"gateway,omitempty"`
	DNS        []string `json:"dns,omitempty"`
}

// Info gets the info for the via
func (v *Via) Info(ctx context.Context) ({}interface, error) {
	v.Log.Info("Getting hardware info", zap.String("address", v.Address))

	var toReturn HardwareInfo
	var cmd command

	// get serial number
	cmd.Command = "GetSerialNo"

	serial, err := v.sendCommand(ctx, cmd)
	if err != nil {
		return toReturn, fmt.Errorf("failed to get serial number from %s", v.Address)
	}

	toReturn.SerialNumber = parseResponse(serial, "|")

	// get firmware version
	cmd.Command = "GetVersion"

	version, err := v.sendCommand(ctx, cmd)
	if err != nil {
		return toReturn, fmt.Errorf("failed to get the firmware version of %s", v.Address)
	}

	toReturn.FirmwareVersion = parseResponse(version, "|")

	// get MAC address
	cmd.Command = "GetMacAdd"

	macAddr, err := v.sendCommand(ctx, cmd)
	if err != nil {
		return toReturn, fmt.Errorf("failed to get the MAC address of %s", v.Address)
	}

	// get IP information
	cmd.Command = "IpInfo"

	ipInfo, err := v.sendCommand(ctx, cmd)
	if err != nil {
		return toReturn, fmt.Errorf("failed to get the IP information from %s", v.Address)
	}

	hostname, network := parseIPInfo(ipInfo)

	toReturn.Hostname = hostname
	network.MACAddress = parseResponse(macAddr, "|")
	toReturn.NetworkInfo = network

	return toReturn, nil
}

func parseResponse(resp string, delimiter string) string {
	pieces := strings.Split(resp, delimiter)

	var msg string

	if len(pieces) < 2 {
		msg = pieces[0]
	} else {
		msg = pieces[1]
	}

	return strings.Trim(msg, "\r\n")
}

func parseIPInfo(ip string) (hostname string, network NetworkInfo) {
	ipList := strings.Split(ip, "|")

	for _, item := range ipList {
		if strings.Contains(item, "IP") {
			network.IPAddress = strings.Split(item, ":")[1]
		}
		if strings.Contains(item, "GAT") {
			network.Gateway = strings.Split(item, ":")[1]
		}
		if strings.Contains(item, "DNS") {
			network.DNS = []string{strings.Split(item, ":")[1]}
		}
		if strings.Contains(item, "Host") {
			hostname = strings.Trim(strings.Split(item, ":")[1], "\r\n")
		}
	}

	return hostname, network
}

func (v *Via) Healthy(ctx context.Context) error {
	_, err := v.Volumes(ctx, []string{})
	if err != nil {
		return fmt.Errorf("unable to get volume (not healthy): %s", err)
	}

	return nil
}
