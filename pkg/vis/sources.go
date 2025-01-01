package vis

import (
	"io"
	"net"
	"os"
	"os/exec"
	"sync"
)

// StreamMsgSource implements MsgSource simply using
// Reader/Writer for line-based JSON encoded messages
// in plain text
type StreamMsgSource struct {
	Reader io.Reader
	Writer io.Writer
}

// RecvMessages implements MessageSink
func (s *StreamMsgSource) RecvMessages(msgs []Msg) {
	s.Writer.Write([]byte(string(MustEncode(msgs)) + "\n"))
}

// ProcessMessages implements MsgSource
func (s *StreamMsgSource) ProcessMessages(sink MessageSink) error {
	decoder := NewMsgDecoder(s.Reader)
	for {
		msgs, err := decoder.Decode()
		if err != nil {
			return err
		}
		sink.RecvMessages(msgs)
	}
}

type ListenerSource struct {
	ln net.Listener
	clientsLock sync.RWMutex
	clients map[net.Conn]struct{}
}

func NewListenerSource(ln net.Listener) *ListenerSource {
	return &ListenerSource{ln: ln, clients: make(map[net.Conn]struct{})}
}

func (s *ListenerSource) RecvMessages(msgs []Msg) {
	s.clientsLock.RLock()
	conns := make([]net.Conn, 0, len(s.clients))
	for conn := range s.clients {
		conns = append(conns, conn)
	}
	s.clientsLock.RUnlock()
	if len(conns) == 0 {
		return
	}
	data := []byte(string(MustEncode(msgs)) + "\n")
	for _, conn := range conns {
		conn.Write(data)
	}
}

func (s *ListenerSource) ProcessMessages(sink MessageSink) error {
	defer s.ln.Close()
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			return io.EOF
		}
		s.clientsLock.Lock()
		s.clients[conn] = struct{}{}
		s.clientsLock.Unlock()
		go s.serveConn(conn, sink)
	}
}

func (s *ListenerSource) serveConn(conn net.Conn, sink MessageSink) {
	stream := &StreamMsgSource{Reader: conn}
	stream.ProcessMessages(sink)
	s.clientsLock.Lock()
	delete(s.clients, conn)
	s.clientsLock.Unlock()
	conn.Close()
}

// ExecMsgSource spawns an external process and use stdin/stdout to
// exchange messages
type ExecMsgSource struct {
	Cmd *exec.Cmd
	rw  StreamMsgSource
}

// NewExecMsgSource creates a new ExecMsgSource using command line
func NewExecMsgSource(prog string, args ...string) (s *ExecMsgSource, err error) {
	s = &ExecMsgSource{Cmd: exec.Command(prog, args...)}
	if s.rw.Reader, err = s.Cmd.StdoutPipe(); err != nil {
		return
	}
	if s.rw.Writer, err = s.Cmd.StdinPipe(); err != nil {
		s.rw.Reader.(io.Closer).Close()
		s.rw.Reader = nil
		return
	}
	s.Cmd.Stderr = os.Stderr
	s.Cmd.Env = os.Environ()
	return
}

// RecvMessages implements MessageSink
func (s *ExecMsgSource) RecvMessages(msgs []Msg) {
	s.rw.RecvMessages(msgs)
}

// ProcessMessages implements MsgSource
func (s *ExecMsgSource) ProcessMessages(sink MessageSink) error {
	return s.rw.ProcessMessages(sink)
}

// Close closes direction pipes
func (s *ExecMsgSource) Close() error {
	if r := s.rw.Reader; r != nil {
		r.(io.Closer).Close()
		s.rw.Reader = nil
	}
	if w := s.rw.Writer; w != nil {
		w.(io.Closer).Close()
		s.rw.Writer = nil
	}
	return nil
}
