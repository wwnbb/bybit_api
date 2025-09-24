package bybit_api

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"nhooyr.io/websocket"
)

type HTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

type ApiMockServer struct {
	wsServer    *http.Server
	httpServer  *http.Server
	upgrader    *http.Server // Placeholder for compatibility
	wsAddr      string
	httpAddr    string
	wsPort      int
	httpPort    int
	WsInput     chan []byte
	WsOutput    chan []byte
	HTTPInput   chan HTTPResponse
	wsConns     map[*websocket.Conn]bool
	mu          sync.Mutex
	addrMu      sync.RWMutex
	serverMu    sync.Mutex
	ctx         context.Context
	cancel      context.CancelFunc
	router      *mux.Router
	defaultResp HTTPResponse
	wsReady     chan struct{}
	httpReady   chan struct{}
}

func NewFreedomApiMockServer() *ApiMockServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &ApiMockServer{
		upgrader:  nil, // nhooyr.io/websocket doesn't use upgrader pattern
		WsInput:   make(chan []byte, 100),
		WsOutput:  make(chan []byte, 100),
		HTTPInput: make(chan HTTPResponse, 100),
		wsConns:   make(map[*websocket.Conn]bool),
		ctx:       ctx,
		cancel:    cancel,
		router:    mux.NewRouter(),
		defaultResp: HTTPResponse{
			StatusCode: http.StatusOK,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       []byte(`{"message": "Default response"}`),
		},
		wsReady:   make(chan struct{}),
		httpReady: make(chan struct{}),
	}
}

func (s *ApiMockServer) Start() error {
	errChan := make(chan error, 2)

	go func() {
		err := s.startWsServer()
		if err != nil {
			errChan <- fmt.Errorf("WebSocket server error: %v", err)
		}
	}()

	go func() {
		err := s.startHttpServer()
		if err != nil {
			errChan <- fmt.Errorf("HTTP server error: %v", err)
		}
	}()

	// Wait for both servers to be ready or for an error
	for i := 0; i < 2; i++ {
		select {
		case err := <-errChan:
			return err
		case <-s.wsReady:
		case <-s.httpReady:
		case <-time.After(5 * time.Second):
			return fmt.Errorf("timeout waiting for servers to start")
		}
	}

	go s.broadcastRoutine()
	return nil
}

func (s *ApiMockServer) startWsServer() error {
	listener, err := s.getListener(s.wsPort)
	if err != nil {
		return err
	}

	s.addrMu.Lock()
	s.wsAddr = listener.Addr().String()
	s.wsPort = listener.Addr().(*net.TCPAddr).Port
	s.addrMu.Unlock()

	mux := http.NewServeMux()
	mux.HandleFunc("/*", s.handleWebSocket)

	s.serverMu.Lock()
	s.wsServer = &http.Server{
		Handler: mux,
	}
	s.serverMu.Unlock()

	close(s.wsReady)

	return s.wsServer.Serve(listener)
}

func (s *ApiMockServer) startHttpServer() error {
	listener, err := s.getListener(s.httpPort)
	if err != nil {
		return err
	}

	s.addrMu.Lock()
	s.httpAddr = listener.Addr().String()
	s.httpPort = listener.Addr().(*net.TCPAddr).Port
	s.addrMu.Unlock()

	s.setupRESTRoutes()

	s.serverMu.Lock()
	s.httpServer = &http.Server{
		Handler: s.router,
	}
	s.serverMu.Unlock()

	close(s.httpReady)

	return s.httpServer.Serve(listener)
}

func (s *ApiMockServer) getListener(port int) (net.Listener, error) {
	if port == 0 {
		return net.Listen("tcp4", "127.0.0.1:0")
	}
	return net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
}

func (s *ApiMockServer) Stop() error {
	s.cancel()

	s.mu.Lock()
	defer s.mu.Unlock()

	for conn := range s.wsConns {
		conn.Close(websocket.StatusNormalClosure, "server shutdown")
		delete(s.wsConns, conn)
	}

	s.serverMu.Lock()
	defer s.serverMu.Unlock()

	var wsErr, httpErr error
	if s.wsServer != nil {
		wsErr = s.wsServer.Shutdown(context.Background())
	}
	if s.httpServer != nil {
		httpErr = s.httpServer.Shutdown(context.Background())
	}

	if wsErr != nil {
		return wsErr
	}
	return httpErr
}

func (s *ApiMockServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	s.mu.Lock()
	s.wsConns[conn] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.wsConns, conn)
		s.mu.Unlock()
		conn.Close(websocket.StatusNormalClosure, "connection closed")
	}()

	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			_, message, err := conn.Read(s.ctx)
			log.Println("Received message:", string(message))
			if err != nil {
				log.Println("read error:", err)
				return
			}
			s.WsOutput <- message
		}
	}
}

func (s *ApiMockServer) broadcastRoutine() {
	fmt.Println("Starting broadcast routine")
	for {
		select {
		case <-s.ctx.Done():
			return
		case message := <-s.WsInput:
			s.mu.Lock()
			for conn := range s.wsConns {
				err := conn.Write(s.ctx, websocket.MessageText, message)
				if err != nil {
					log.Println("write error:", err)
					delete(s.wsConns, conn)
					conn.Close(websocket.StatusInternalError, "write error")
				}
			}
			s.mu.Unlock()
		}
	}
}

func (s *ApiMockServer) WsURL() string {
	s.addrMu.RLock()
	defer s.addrMu.RUnlock()
	return fmt.Sprintf("ws://%s/ws/", s.wsAddr)
}

func (s *ApiMockServer) HttpURL() string {
	s.addrMu.RLock()
	defer s.addrMu.RUnlock()
	return fmt.Sprintf("http://%s/", s.httpAddr)
}

func (s *ApiMockServer) Restart() error {
	err := s.Stop()
	if err != nil {
		return err
	}

	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.wsConns = make(map[*websocket.Conn]bool)
	s.router = mux.NewRouter()
	s.wsReady = make(chan struct{})
	s.httpReady = make(chan struct{})

	return s.Start()
}

func (s *ApiMockServer) setupRESTRoutes() {
	s.router.PathPrefix("/").HandlerFunc(s.handleHTTPRequest)
}

func (s *ApiMockServer) handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	select {
	case resp := <-s.HTTPInput:
		for key, value := range resp.Headers {
			w.Header().Set(key, value)
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(resp.Body)
	case <-s.ctx.Done():
		http.Error(w, "Server is shutting down", http.StatusServiceUnavailable)
	default:
		for key, value := range s.defaultResp.Headers {
			w.Header().Set(key, value)
		}
		w.WriteHeader(s.defaultResp.StatusCode)
		w.Write(s.defaultResp.Body)
	}
}

func (s *ApiMockServer) SetDefaultResponse(statusCode int, headers map[string]string, body []byte) {
	s.defaultResp = HTTPResponse{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body,
	}
}
