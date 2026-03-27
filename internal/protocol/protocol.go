package protocol

// Command types sent from CLI to daemon
const (
	CmdStart           = "start"
	CmdEnd             = "end"
	CmdSwitch          = "switch"
	CmdNext            = "next"
	CmdRest            = "rest"
	CmdEdit            = "edit"
	CmdStatus          = "status"
	CmdProjects        = "projects"
	CmdRegister        = "register"
	CmdUnregister      = "unregister"
	CmdCheckResume     = "check-resume"
	CmdResume          = "resume"
	CmdDiscardAndStart = "discard-and-start"
)

// Request is the message sent from CLI to daemon
type Request struct {
	Command string `json:"command"`
	Payload string `json:"payload,omitempty"` // e.g. project name for switch/register
}

// Response is what the daemon sends back
type Response struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`        // human-readable output to print
	Data    any    `json:"data,omitempty"` // optional structured data
}

// ResumeCandidate is returned by CheckResume when an unfinished session exists
type ResumeCandidate struct {
	SessPath      string
	ActiveProject string
	SessAt        string // formatted last segment start
}


// SocketPath is the unix socket location
const SocketPath = "/tmp/trak.sock"
