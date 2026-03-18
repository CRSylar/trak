package protocol

// Command types sent from CLI to daemon
const (
	CmdStart      = "start"
	CmdEnd        = "end"
	CmdSwitch     = "switch"
	CmdNext       = "next"
	CmdRest       = "rest"
	CmdStatus     = "status"
	CmdProjects   = "projects"
	CmdRegister   = "register"
	CmdUnregister = "unregister"
)

// Request is the message sent from CLI to daemon
type Request struct {
	Command string `json:"command"`
	Payload string `json:"payload,omitempty"` // e.g. project name for switch/register
}

// Response is what the daemon sends back
type Response struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`          // human-readable output to print
	Data    any    `json:"data,omitempty"`   // optional structured data
}

// SocketPath is the unix socket location
const SocketPath = "/tmp/trak.sock"
