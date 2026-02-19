package ipc

import (
	"encoding/json"
	"log"
	"net"
	"src/api"

	"github.com/Microsoft/go-winio"
)

type Server struct {
	apiServer *api.Server
}

func NewServer(apiServer *api.Server) *Server {
	return &Server{apiServer: apiServer}
}

func (s *Server) Start() error {
	address := GetIPCAddress()

	// Use go-winio for Windows Named Pipes
	config := &winio.PipeConfig{}
	listener, err := winio.ListenPipe(address, config)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("IPC Server listening on named pipe: %s", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var req Request
		if err := decoder.Decode(&req); err != nil {
			return
		}

		resp := s.dispatch(req)
		if err := encoder.Encode(resp); err != nil {
			log.Printf("Error encoding response: %v", err)
			return
		}
	}
}

func (s *Server) dispatch(req Request) Response {
	var result interface{}
	var err error

	switch req.Method {
	case "GetAppLeaderboard":
		var params struct {
			Since string `json:"since"`
			Until string `json:"until"`
		}
		json.Unmarshal(req.Params, &params)
		result, err = s.apiServer.Apps.GetAppLeaderboard(params.Since, params.Until)

	case "GetAppBlocklist":
		result, err = s.apiServer.GetAppBlocklist()

	case "BlockApps":
		var names []string
		json.Unmarshal(req.Params, &names)
		err = s.apiServer.BlockApps(names)

	case "UnblockApps":
		var names []string
		json.Unmarshal(req.Params, &names)
		err = s.apiServer.UnblockApps(names)

	case "GetScreenTime":
		result, err = s.apiServer.Apps.GetScreenTime()

	case "GetTotalScreenTime":
		result, err = s.apiServer.Apps.GetTotalScreenTime()

	case "CheckChromeExtension":
		// TODO: Implement logic from bindings.go
		result = false

	case "Shutdown":
		s.apiServer.Shutdown()
		result = true

	default:
		return Response{ID: req.ID, Error: "Unknown method: " + req.Method}
	}

	if err != nil {
		return Response{ID: req.ID, Error: err.Error()}
	}

	return Response{ID: req.ID, Result: result}
}
