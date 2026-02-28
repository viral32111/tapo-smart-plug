package commands

import (
	"encoding/json"
	"fmt"

	"github.com/viral32111/tapo-smart-plug/internal/klap"
	"github.com/viral32111/tapo-smart-plug/internal/protocol"
)

type CurrentPower struct {
	CurrentEnergy int `json:"current_power"` // watts
}

func GetCurrentPower(klapSession *klap.Session) (*CurrentPower, error) {
	result, err := protocol.Execute(klapSession, "get_current_power", map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("Error while fetching current power: %w", err)
	}

	var data CurrentPower
	if err := json.Unmarshal(result, &data); err != nil {
		return nil, fmt.Errorf("Error parsing current power response: %w", err)
	}

	return &data, nil
}

func GetChildCurrentPower(klapSession *klap.Session, childDeviceId string) (*CurrentPower, error) {
	result, err := protocol.ControlChild(klapSession, childDeviceId, "get_current_power", map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("Error while fetching child device current power: %w", err)
	}

	var data CurrentPower
	if err := json.Unmarshal(result, &data); err != nil {
		return nil, fmt.Errorf("Error parsing child device current power response: %w", err)
	}

	return &data, nil
}
