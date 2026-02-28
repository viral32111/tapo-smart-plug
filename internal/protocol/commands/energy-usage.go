package commands

import (
	"encoding/json"
	"fmt"

	"github.com/viral32111/tapo-smart-plug/internal/klap"
	"github.com/viral32111/tapo-smart-plug/internal/protocol"
)

type EnergyUsage struct {
	// TODO: For some reason P304M child devices do not return a 'current_power' property?

	CurrentEnergy int `json:"current_power"` // milliwatts

	TodayEnergy  int `json:"today_energy"`  // watt-hours
	TodayRuntime int `json:"today_runtime"` // minutes

	MonthEnergy  int `json:"month_energy"`  // watt-hours
	MonthRuntime int `json:"month_runtime"` // minutes
}

// https://github.com/python-kasa/python-kasa/blob/c9c80619d76e21bd1458f10084c6d2ef2daae8f1/tests/fixtures/smart/P110(UK)_1.0_1.3.0.json#L334C6-L334C22
func GetEnergyUsage(klapSession *klap.Session) (*EnergyUsage, error) {
	result, err := protocol.Execute(klapSession, "get_energy_usage", map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("Error while fetching energy usage: %w", err)
	}

	var data EnergyUsage
	if err := json.Unmarshal(result, &data); err != nil {
		return nil, fmt.Errorf("Error parsing energy usage response: %w", err)
	}

	return &data, nil
}

// https://github.com/python-kasa/python-kasa/blob/c9c80619d76e21bd1458f10084c6d2ef2daae8f1/tests/fixtures/smart/P304M(UK)_1.0_1.0.3.json#L298
func GetChildEnergyUsage(klapSession *klap.Session, childDeviceId string) (*EnergyUsage, error) {
	result, err := protocol.ControlChild(klapSession, childDeviceId, "get_energy_usage", map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("Error while fetching child device energy usage: %w", err)
	}

	var data EnergyUsage
	if err := json.Unmarshal(result, &data); err != nil {
		return nil, fmt.Errorf("Error parsing child device energy usage response: %w", err)
	}

	return &data, nil
}
