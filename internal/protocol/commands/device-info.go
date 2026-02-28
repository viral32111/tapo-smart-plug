package commands

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/viral32111/tapo-smart-plug/internal/klap"
	"github.com/viral32111/tapo-smart-plug/internal/protocol"
)

type DeviceInfo struct {
	ID string `json:"device_id"`

	Type  string  `json:"type"`
	Model string  `json:"model"`
	Name  *string `json:"nickname"` // Base64-encoded, empty for P304M

	FirmwareDescription string `json:"fw_ver"`
	HardwareDescription string `json:"hw_ver"`

	WirelessSSID        string `json:"ssid"` // Base64-encoded
	WirelessSignalLevel int    `json:"signal_level"`

	IPAddress  string `json:"ip"`
	MACAddress string `json:"mac"`

	Region   string `json:"region"`
	Language string `json:"lang"`

	SwitchStatus  *bool `json:"device_on"` // Missing for P304M
	UptimeSeconds *int  `json:"on_time"`   // Missing for P304M

	// Only for P304M child devices
	Slot *int `json:"position"`
}

func GetDeviceInfo(klapSession *klap.Session) (*DeviceInfo, error) {
	result, err := protocol.Execute(klapSession, "get_device_info", map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("Error while fetching device info: %w", err)
	}

	var data DeviceInfo
	if err := json.Unmarshal(result, &data); err != nil {
		return nil, fmt.Errorf("Error parsing device info response: %w", err)
	}

	// Nickname is base64-encoded
	if data.Name != nil {
		nickname, err := base64.StdEncoding.DecodeString(*data.Name)
		if err != nil {
			return nil, fmt.Errorf("Error base64 decoding device name: %w", err)
		}

		decodedName := string(nickname)
		data.Name = &decodedName
	}

	// SSID is base64-encoded
	ssid, err := base64.StdEncoding.DecodeString(data.WirelessSSID)
	if err != nil {
		return nil, fmt.Errorf("Error base64 decoding wireless SSID: %w", err)
	}
	data.WirelessSSID = string(ssid)

	// Remove hyphen in P110 MAC addresses for consistency with P304M child devices
	data.MACAddress = strings.ReplaceAll(data.MACAddress, "-", "")

	return &data, nil
}

func GetChildDeviceInfo(klapSession *klap.Session, childDeviceId string) (*DeviceInfo, error) {
	result, err := protocol.ControlChild(klapSession, childDeviceId, "get_device_info", map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("Error while fetching child device info: %w", err)
	}

	var data DeviceInfo
	if err := json.Unmarshal(result, &data); err != nil {
		return nil, fmt.Errorf("Error parsing child device info response: %w", err)
	}

	// Nickname is base64-encoded
	if data.Name != nil {
		nickname, err := base64.StdEncoding.DecodeString(*data.Name)
		if err != nil {
			return nil, fmt.Errorf("Error base64 decoding child device name: %w", err)
		}

		decodedName := string(nickname)
		data.Name = &decodedName
	}

	return &data, nil
}
