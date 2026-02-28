package commands

import (
	"encoding/json"
	"fmt"

	"github.com/viral32111/tapo-smart-plug/internal/klap"
	"github.com/viral32111/tapo-smart-plug/internal/protocol"
)

type ChildDeviceList struct {
	Total   int          `json:"sum"`
	Index   int          `json:"start_index"`
	Devices []DeviceInfo `json:"child_device_list"`
}

func GetChildDeviceList(klapSession *klap.Session) (*ChildDeviceList, error) {
	result, err := protocol.Execute(klapSession, "get_child_device_list", map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("Error while fetching device info: %w", err)
	}

	var data ChildDeviceList
	if err := json.Unmarshal(result, &data); err != nil {
		return nil, fmt.Errorf("Error parsing device info response: %w", err)
	}

	return &data, nil
}
