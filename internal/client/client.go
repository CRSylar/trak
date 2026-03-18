package client

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/CRSylar/trak/internal/protocol"
)

// Send sends a request to the daemon and returns the response message.
// Returns an error if the daemon is unreachable or the command failed.
func Send(cmd, payload string) (string, error) {
	conn, err := net.DialTimeout("unix", protocol.SocketPath, 2*time.Second)
	if err != nil {
		return "", fmt.Errorf("daemon not running — start your workday with 'trak start'")
	}
	defer conn.Close()

	req := protocol.Request{Command: cmd, Payload: payload}
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	var resp protocol.Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if !resp.OK {
		return "", fmt.Errorf("%s", resp.Message)
	}
	return resp.Message, nil
}
