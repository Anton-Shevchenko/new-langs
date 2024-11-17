package callback_builder

import (
	"encoding/json"
	"fmt"
	"strings"
)

const separator = "|"

type CallBackMsg struct {
	Action string
	Params map[string]string
}

func (c *CallBackMsg) String() string {
	params, _ := json.Marshal(c.Params)

	return fmt.Sprintf("%s/%s", c.Action, string(params))
}

func Parse(call string) *CallBackMsg {
	var params map[string]string

	parts := strings.Split(call, "/")

	if len(parts) > 2 {
		_ = json.Unmarshal([]byte(parts[2]), &params)
	}

	return &CallBackMsg{
		Action: parts[1],
		Params: params,
	}
}
