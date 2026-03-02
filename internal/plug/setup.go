package plug

import (
	"fmt"

	"github.com/viral32111/tapo-smart-plug/internal/config"
	"github.com/viral32111/tapo-smart-plug/internal/klap"
	"github.com/viral32111/tapo-smart-plug/internal/protocol/commands"
)

func Setup(plugConfig config.PlugConfig, emailAddress string, password string) ([]*commands.DeviceInfo, *klap.Session, error) {
	klapSession, err := klap.NewSession(plugConfig.IPAddress, emailAddress, password)
	if err != nil {
		return nil, nil, fmt.Errorf("Error communicating with plug '%s' (%s): %w", plugConfig.Name, plugConfig.IPAddress, err)
	}

	deviceInfo, err := commands.GetDeviceInfo(klapSession)
	if err != nil {
		return nil, nil, fmt.Errorf("Error fetching device info from '%s' (%s): %w", plugConfig.Name, plugConfig.IPAddress, err)
	}

	var devices []*commands.DeviceInfo

	// Only add child plugs for extension cords
	if deviceInfo.Model == "P304M" {
		childList, err := commands.GetChildDeviceList(klapSession)
		if err != nil {
			return nil, nil, fmt.Errorf("Error fetching child device list from '%s' (%s): %w", plugConfig.Name, plugConfig.IPAddress, err)
		}

		for _, child := range childList.Devices {
			childDeviceInfo, err := commands.GetChildDeviceInfo(klapSession, child.ID)
			if err != nil {
				return nil, nil, fmt.Errorf("Error fetching device info for child '%s' of '%s' (%s): %w", child.ID, plugConfig.Name, plugConfig.IPAddress, err)
			}

			// Fixes for missing fields that should use parent device's values
			childDeviceInfo.FirmwareDescription = deviceInfo.FirmwareDescription
			childDeviceInfo.HardwareDescription = deviceInfo.HardwareDescription
			childDeviceInfo.IPAddress = deviceInfo.IPAddress
			childDeviceInfo.MACAddress = child.MACAddress
			childDeviceInfo.WirelessSSID = deviceInfo.WirelessSSID
			childDeviceInfo.WirelessSignalLevel = deviceInfo.WirelessSignalLevel
			childDeviceInfo.Language = deviceInfo.Language

			devices = append(devices, childDeviceInfo)
		}
	} else {
		devices = append(devices, deviceInfo)
	}

	return devices, klapSession, nil
}
