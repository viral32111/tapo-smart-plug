package protocol

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/viral32111/tapo-smart-plug/internal/helpers"
	"github.com/viral32111/tapo-smart-plug/internal/klap"
)

type Request struct {
	Method           string `json:"method"`
	RequestTimestamp int64  `json:"request_time_milis"`
	TerminalUUID     string `json:"terminal_uuid"`
	Params           any    `json:"params"`
}

type Response struct {
	ErrorCode int             `json:"error_code"`
	Result    json.RawMessage `json:"result"`
}

func Execute(klapSession *klap.Session, method string, params any) (json.RawMessage, error) {
	terminalUUID, err := helpers.DeriveUUID(klapSession.SessionCookie.Value)
	if err != nil {
		return nil, fmt.Errorf("Error deriving terminal UUID: %w", err)
	}

	payload, err := json.Marshal(Request{
		Method:           method,
		RequestTimestamp: time.Now().UnixMilli(),
		TerminalUUID:     terminalUUID,
		Params:           map[string]any{},
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating payload for '%s' request: %w", method, err)
	}

	response, err := klapSession.Request(payload)
	if err != nil {
		return nil, fmt.Errorf("Request for '%s' failed: %w", method, err)
	}

	var data Response
	if err := json.Unmarshal(response, &data); err != nil {
		return nil, fmt.Errorf("Error parsing '%s' response: %w", method, err)
	}

	if data.ErrorCode != 0 {
		return nil, fmt.Errorf("Device reported error code within '%s' response: %d", method, data.ErrorCode)
	}

	return data.Result, nil
}
