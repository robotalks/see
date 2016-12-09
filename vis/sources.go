package vis

import (
	"io"
	"os"
	"os/exec"
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
