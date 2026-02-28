package protocol

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/viral32111/tapo-smart-plug/internal/helpers"
	"github.com/viral32111/tapo-smart-plug/internal/klap"
)

type ControlChildResponse struct {
	ResponseData Response `json:"responseData"`
}

func ControlChild(klapSession *klap.Session, childDeviceId string, method string, params any) (json.RawMessage, error) {
	terminalUUID, err := helpers.DeriveUUID(klapSession.SessionCookie.Value)
	if err != nil {
		return nil, fmt.Errorf("Error deriving terminal UUID: %w", err)
	}

	payload, err := json.Marshal(Request{
		Method:           "control_child",
		RequestTimestamp: time.Now().UnixMilli(),
		TerminalUUID:     terminalUUID,
		Params: map[string]any{
			"device_id": childDeviceId,
			"requestData": map[string]any{
				"method": method,
				"params": params,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating payload for '%s' (control_child) request: %w", method, err)
	}

	response, err := klapSession.Request(payload)
	if err != nil {
		return nil, fmt.Errorf("Request for '%s' (control_child) failed: %w", method, err)
	}

	// fmt.Printf("\ncontrol_child -> %s\n%s\n\n", method, string(response))

	// Step 1
	var data Response
	if err := json.Unmarshal(response, &data); err != nil {
		return nil, fmt.Errorf("Error parsing '%s' (control_child) response: %w", method, err)
	}

	if data.ErrorCode != 0 {
		return nil, fmt.Errorf("Device reported error code within '%s' response: %d", method, data.ErrorCode)
	}

	// Step 2
	var child ControlChildResponse
	if err := json.Unmarshal(data.Result, &child); err != nil {
		return nil, fmt.Errorf("Error parsing '%s' (control_child) response: %w", method, err)
	}

	if child.ResponseData.ErrorCode != 0 {
		return nil, fmt.Errorf("Device reported error code within '%s' (control_child) response: %d", method, child.ResponseData.ErrorCode)
	}

	return child.ResponseData.Result, nil
}
