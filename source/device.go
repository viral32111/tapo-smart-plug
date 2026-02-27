package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

/*
P110: {
	"device_id": "00000000000000000000000000000000",
	"oem_id": "00000000000000000000000000000000",

	"type": "SMART.TAPOPLUG"
	"model": "P110M",

	"nickname": "BASE64",
	"avatar": "house",


	"fw_id": "00000000000000000000000000000000",
	"fw_ver": "1.2.3 Build 240617 Rel.153525",

	"hw_id": "00000000000000000000000000000000",
	"hw_ver": "1.0",


	"ssid": "BASE64",
	"rssi": -44,
	"signal_level": 3,

	"ip": "192.168.10.10",
	"mac": "00-00-00-00-00-00",

	"region": "Europe/London",
	"lang": "en_US",
	"has_set_location_info": true,
	"latitude": 00000,
	"longitude": -00000,

	"auto_off_remain_time": 0,
	"auto_off_status": "off",
	"charging_status": "normal",
	"default_states": {
		"state": {
			"on": true
		},
		"type": "custom"
	},
	"device_on": true,
	"on_time": 1699706,
	"time_diff": 0,
	"overcurrent_status": "normal",
	"overheat_status": "normal",
	"power_protection_status": "normal",
	"specs": "",
}

P304M: {
	"avatar": "",
	"device_id": "00000000000000000000000000000000",
	"fw_id": "00000000000000000000000000000000",
	"fw_ver": "1.0.9 Build 251021 Rel.100234",
	"has_set_location_info": true,
	"hw_id": "00000000000000000000000000000000",
	"hw_ver": "1.0",
	"ip": "192.168.100.10",
	"lang": "en_US",
	"latitude": 00000,
	"longitude": -0000,
	"mac": "00-00-00-00-00-00",
	"model": "P304M",
	"nickname": "",
	"oem_id": "00000000000000000000000000000000",
	"region": "Europe/London",
	"rssi": -49,
	"signal_level": 3,
	"specs": "",
	"ssid": "BASE64",
	"time_diff": 0,
	"type": "SMART.TAPOPLUG"
}

*/

type DeviceInfo struct {
	// Only what we care about

	ID string `json:"device_id"`

	Type  string  `json:"type"`
	Model string  `json:"model"`
	Name  *string `json:"nickname"` // Base64-encoded, empty for P304M
	Slot  *int    `json:"position"` // Only for child devices of P304M

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
}

type ChildDeviceList struct {
	Total      int          `json:"sum"`
	StartIndex int          `json:"start_index"`
	Devices    []DeviceInfo `json:"child_device_list"`
}

type ControlChild_Response struct {
	ResponseData KLAPResponse `json:"responseData"`
}

type EnergyUsage struct {
	CurrentEnergy int `json:"current_power"` // milliwatts

	TodayEnergy  int `json:"today_energy"`  // watt-hours
	TodayRuntime int `json:"today_runtime"` // minutes

	MonthEnergy  int `json:"month_energy"`  // watt-hours
	MonthRuntime int `json:"month_runtime"` // minutes
}

type KLAPRequest struct {
	Method           string `json:"method"`
	RequestTimestamp int64  `json:"request_time_milis"`
	TerminalUUID     string `json:"terminal_uuid"`
	Params           any    `json:"params"`
}

type KLAPResponse struct {
	ErrorCode int             `json:"error_code"`
	Result    json.RawMessage `json:"result"`
}

/************************************************************/

func GetDeviceInfo(session *KLAPSession) (*DeviceInfo, error) {
	payload, err := json.Marshal(KLAPRequest{
		Method:           "get_device_info",
		RequestTimestamp: time.Now().UnixMilli(),
		TerminalUUID:     session.terminalUUID,
		Params:           map[string]any{},
	})
	if err != nil {
		return nil, err
	}

	response, err := session.Request(payload)
	if err != nil {
		return nil, fmt.Errorf("KLAP request for 'get_device_info' failed: %w", err)
	}

	var data KLAPResponse
	if err := json.Unmarshal(response, &data); err != nil {
		return nil, fmt.Errorf("Error parsing 'get_device_info' response: %w", err)
	}

	if data.ErrorCode != 0 {
		return nil, fmt.Errorf("Device reported error code during 'get_device_info': %d", data.ErrorCode)
	}

	var deviceInfo DeviceInfo
	if err := json.Unmarshal(data.Result, &deviceInfo); err != nil {
		return nil, fmt.Errorf("Error parsing 'get_device_info' data: %w", err)
	}

	// Nickname is base64-encoded
	if deviceInfo.Name != nil {
		nickname, err := base64.StdEncoding.DecodeString(*deviceInfo.Name)
		if err != nil {
			return nil, fmt.Errorf("Error base64 decoding device name: %w", err)
		}

		decodedName := string(nickname)
		deviceInfo.Name = &decodedName
	}

	// SSID is base64-encoded
	ssid, err := base64.StdEncoding.DecodeString(deviceInfo.WirelessSSID)
	if err != nil {
		return nil, fmt.Errorf("Error base64 decoding wireless SSID: %w", err)
	}
	deviceInfo.WirelessSSID = string(ssid)

	// Remove hyphen in P110 MAC addresses for consistency with P304M child devices
	deviceInfo.MACAddress = strings.ReplaceAll(deviceInfo.MACAddress, "-", "")

	return &deviceInfo, nil
}

func GetChildDeviceList(session *KLAPSession) (*ChildDeviceList, error) {
	payload, err := json.Marshal(KLAPRequest{
		Method:           "get_child_device_list",
		RequestTimestamp: time.Now().UnixMilli(),
		TerminalUUID:     session.terminalUUID,
		Params:           map[string]any{},
	})
	if err != nil {
		return nil, err
	}

	response, err := session.Request(payload)
	if err != nil {
		return nil, fmt.Errorf("KLAP request for 'get_child_device_list' failed: %w", err)
	}

	var data KLAPResponse
	if err := json.Unmarshal(response, &data); err != nil {
		return nil, fmt.Errorf("Error parsing 'get_child_device_list' response: %w", err)
	}

	if data.ErrorCode != 0 {
		return nil, fmt.Errorf("Device reported error code during 'get_child_device_list': %d", data.ErrorCode)
	}

	var childDeviceList *ChildDeviceList
	if err := json.Unmarshal(data.Result, &childDeviceList); err != nil {
		return nil, fmt.Errorf("Error parsing 'get_child_device_list' data: %w", err)
	}

	return childDeviceList, nil
}

func ControlChild_GetDeviceInfo(session *KLAPSession, childDeviceId string) (*DeviceInfo, error) {
	payload, err := json.Marshal(KLAPRequest{
		Method:           "control_child",
		RequestTimestamp: time.Now().UnixMilli(),
		TerminalUUID:     session.terminalUUID,
		Params: map[string]any{
			"device_id": childDeviceId,
			"requestData": map[string]any{
				"method": "get_device_info",
				"params": map[string]any{},
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("Error creating payload for 'control_child' request: %w", err)
	}

	response, err := session.Request(payload)
	if err != nil {
		return nil, fmt.Errorf("KLAP request for 'control_child' failed: %w", err)
	}

	var outerData KLAPResponse
	if err := json.Unmarshal(response, &outerData); err != nil {
		return nil, fmt.Errorf("Error parsing 'control_child' response: %w", err)
	}

	if outerData.ErrorCode != 0 {
		return nil, fmt.Errorf("Device reported error code during 'control_child': %d", outerData.ErrorCode)
	}

	var innerData ControlChild_Response
	if err := json.Unmarshal(outerData.Result, &innerData); err != nil {
		return nil, fmt.Errorf("Error parsing 'control_child' inner response: %w", err)
	}

	if innerData.ResponseData.ErrorCode != 0 {
		return nil, fmt.Errorf("Device reported error code during 'control_child' inner request: %d", innerData.ResponseData.ErrorCode)
	}

	var deviceInfo DeviceInfo
	if err := json.Unmarshal(innerData.ResponseData.Result, &deviceInfo); err != nil {
		return nil, fmt.Errorf("Error parsing 'control_child' data: %w", err)
	}

	// Name is base64-encoded
	if deviceInfo.Name != nil {
		nickname, err := base64.StdEncoding.DecodeString(*deviceInfo.Name)
		if err != nil {
			return nil, fmt.Errorf("Error base64 decoding child device name: %w", err)
		}

		decodedName := string(nickname)
		deviceInfo.Name = &decodedName
	}

	return &deviceInfo, nil
}

func GetEnergyUsage(session *KLAPSession) (*EnergyUsage, error) {
	payload, err := json.Marshal(KLAPRequest{
		Method:           "get_energy_usage",
		RequestTimestamp: time.Now().UnixMilli(),
		TerminalUUID:     session.terminalUUID,
		Params:           map[string]any{},
	})
	if err != nil {
		return nil, err
	}

	response, err := session.Request(payload)
	if err != nil {
		return nil, fmt.Errorf("KLAP request for 'get_energy_usage' failed: %w", err)
	}

	var data KLAPResponse
	if err := json.Unmarshal(response, &data); err != nil {
		return nil, fmt.Errorf("Error parsing 'get_energy_usage' response: %w", err)
	}

	if data.ErrorCode != 0 {
		return nil, fmt.Errorf("Device reported error code during 'get_energy_usage': %d", data.ErrorCode)
	}

	var energyUsage EnergyUsage
	if err := json.Unmarshal(data.Result, &energyUsage); err != nil {
		return nil, fmt.Errorf("Error parsing 'get_energy_usage' data: %w", err)
	}

	return &energyUsage, nil
}

func ControlChild_GetEnergyUsage(session *KLAPSession, childDeviceId string) (*EnergyUsage, error) {
	payload, err := json.Marshal(KLAPRequest{
		Method:           "control_child",
		RequestTimestamp: time.Now().UnixMilli(),
		TerminalUUID:     session.terminalUUID,
		Params: map[string]any{
			"device_id": childDeviceId,
			"requestData": map[string]any{
				"method": "get_energy_usage",
				"params": map[string]any{},
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("Error creating payload for 'control_child' request: %w", err)
	}

	response, err := session.Request(payload)
	if err != nil {
		return nil, fmt.Errorf("KLAP request for 'control_child' failed: %w", err)
	}

	var outerData KLAPResponse
	if err := json.Unmarshal(response, &outerData); err != nil {
		return nil, fmt.Errorf("Error parsing 'control_child' response: %w", err)
	}

	if outerData.ErrorCode != 0 {
		return nil, fmt.Errorf("Device reported error code during 'control_child': %d", outerData.ErrorCode)
	}

	var innerData ControlChild_Response
	if err := json.Unmarshal(outerData.Result, &innerData); err != nil {
		return nil, fmt.Errorf("Error parsing 'control_child' inner response: %w", err)
	}

	if innerData.ResponseData.ErrorCode != 0 {
		return nil, fmt.Errorf("Device reported error code during 'control_child' inner request: %d", innerData.ResponseData.ErrorCode)
	}

	var energyUsage EnergyUsage
	if err := json.Unmarshal(innerData.ResponseData.Result, &energyUsage); err != nil {
		return nil, fmt.Errorf("Error parsing 'control_child' energy usage data: %w", err)
	}

	return &energyUsage, nil
}
