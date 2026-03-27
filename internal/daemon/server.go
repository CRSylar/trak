package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/CRSylar/trak/internal/protocol"
)

// Server wraps State and listens on a unix socket
type Server struct {
	state    *State
	listener net.Listener
}

func NewServer() (*Server, error) {
	state, err := New()
	if err != nil {
		return nil, err
	}

	os.Remove(protocol.SocketPath)

	l, err := net.Listen("unix", protocol.SocketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on socket: %w", err)
	}

	return &Server{state: state, listener: l}, nil
}

func (s *Server) Serve() {
	defer os.Remove(protocol.SocketPath)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handleConn(conn)
	}
}

func (s *Server) Stop() {
	s.listener.Close()
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	var req protocol.Request
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		writeResponse(conn, protocol.Response{OK: false, Message: "invalid request: " + err.Error()})
		return
	}

	resp := s.dispatch(req)
	writeResponse(conn, resp)

	if req.Command == protocol.CmdEnd && resp.OK {
		s.Stop()
	}
}

func (s *Server) dispatch(req protocol.Request) protocol.Response {
	var msg string
	var err error

	switch req.Command {

	case protocol.CmdCheckResume:
		candidate, e := s.state.CheckResume()
		if e != nil {
			return protocol.Response{OK: false, Message: e.Error()}
		}
		if candidate == nil {
			return protocol.Response{OK: true, Message: ""}
		}
		data, e := json.Marshal(candidate)
		if e != nil {
			return protocol.Response{OK: false, Message: "failed to marshal resume candidate: " + e.Error()}
		}

		return protocol.Response{OK: true, Message: string(data)}

	case protocol.CmdResume:
		msg, err = s.state.Resume(req.Payload)

	case protocol.CmdDiscardAndStart:
		msg, err = s.state.DiscardAndStart(req.Payload)

	case protocol.CmdStart:
		msg, err = s.state.Start()

	case protocol.CmdEnd:
		msg, err = s.state.End()

	case protocol.CmdNext:
		msg, err = s.state.Next()

	case protocol.CmdRest:
		msg, err = s.state.Rest()

	case protocol.CmdSwitch:
		msg, err = s.state.Switch(req.Payload)

	case protocol.CmdEdit:
		shift, e := time.ParseDuration(req.Payload)
		if e != nil {
			return protocol.Response{OK: false, Message: "invalid duration: " + e.Error()}
		}
		msg, err = s.state.Edit(shift)

	case protocol.CmdStatus:
		msg, err = s.state.Status()

	case protocol.CmdProjects:
		if req.Payload == "names" {
			names := s.state.ListProjectNames()
			data, err := json.Marshal(names)
			if err != nil {
				return protocol.Response{OK: false, Message: fmt.Errorf("CmdProjects failed to Marshal names: %w", err).Error()}
			}

			return protocol.Response{OK: true, Message: string(data)}
		}
		msg, err = s.state.ListProjects()

	case protocol.CmdRegister:
		msg, err = s.state.Register(req.Payload)

	case protocol.CmdUnregister:
		msg, err = s.state.Unregister(req.Payload)

	default:
		err = fmt.Errorf("unknown command: %s", req.Command)
	}

	if err != nil {
		return protocol.Response{OK: false, Message: err.Error()}
	}
	return protocol.Response{OK: true, Message: msg}
}

func writeResponse(conn net.Conn, resp protocol.Response) {
	json.NewEncoder(conn).Encode(resp)
}
