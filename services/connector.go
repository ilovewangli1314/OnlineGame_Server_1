package services

import (
	"context"
	"fmt"

	pbcommon "github.com/ilovewangli1314/OnlineGame_Server_1/protos"
	"github.com/topfreegames/pitaya/component"
)

// ConnectorRemote is a remote that will receive rpc's
type ConnectorRemote struct {
	component.Base
}

// Connector struct
type Connector struct {
	component.Base
}

// SessionData is the session data struct
type SessionData struct {
	Data map[string]interface{} `json:"data"`
}

// RemoteFunc is a function that will be called remotelly
func (c *ConnectorRemote) RemoteFunc(ctx context.Context, message []byte) (*pbcommon.Response, error) {
	fmt.Printf("received a remote call with this message: %s\n", message)
	return &pbcommon.Response{
		Msg: string(message),
	}, nil
}
