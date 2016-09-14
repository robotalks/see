package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"

	logger "github.com/op/go-logging"
	vis "github.com/robotalks/simulator/vis"
)

type visCmd struct {
	Port          int
	Quiet         bool
	PluginDirs    []string `n:"plugin-dir"`
	WebContentDir string   `n:"web-content-dir"`
	Version       bool

	logger *logger.Logger
}

func (c *visCmd) Execute(args []string) error {
	if c.Version {
		PrintVersion()
		return nil
	}

	c.logger = logger.MustGetLogger("visualizer")
	if c.Quiet {
		logger.SetLevel(logger.NOTICE, c.logger.Module)
	} else {
		logger.SetLevel(logger.INFO, c.logger.Module)
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", c.Port))
	if err != nil {
		return err
	}
	srv := &vis.Server{
		Listener:      ln,
		States:        &vis.MemStateStore{},
		WebContentDir: c.WebContentDir,
		Logger:        c.logger,
	}

	if err = c.loadPlugins(srv); err != nil {
		return err
	}

	prog, err := c.startProgram(args...)
	if err != nil {
		return err
	}

	srv.MsgSink = vis.SinkMessage(func(msgs []vis.Msg) {
		c.forwardMsgs(prog.w, msgs)
	})

	c.logger.Noticef("Listen %s", ln.(*net.TCPListener).Addr().String())

	errCh := make(chan error)
	go c.runServer(srv, errCh)
	go c.processOutput(prog.r, srv, errCh)
	return <-errCh
}

func (c *visCmd) loadPlugins(srv *vis.Server) error {
	usr, err := user.Current()
	if err == nil {
		srv.LoadPlugin(filepath.Join(usr.HomeDir, ".sim-ng"))
	}
	wd, err := os.Getwd()
	if err == nil {
		srv.LoadPlugin(wd)
	}
	for _, dir := range c.PluginDirs {
		if err = srv.LoadPlugin(dir); err != nil {
			return err
		}
		c.logger.Infof("Loaded %s", dir)
	}
	return nil
}

type program struct {
	w io.WriteCloser
	r io.ReadCloser
}

func (p *program) close() {
	if p.w != nil {
		p.w.Close()
	}
	if p.r != nil {
		p.r.Close()
	}
}

func (c *visCmd) startProgram(args ...string) (p *program, err error) {
	p = &program{w: os.Stdout, r: os.Stdin}
	if len(args) == 0 || args[0] == "-" || args[0] == "" {
		return
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Env = os.Environ()
	cmd.Stderr = os.Stderr
	if p.w, err = cmd.StdinPipe(); err != nil {
		return
	}
	if p.r, err = cmd.StdoutPipe(); err != nil {
		p.w.Close()
		return
	}
	if err = cmd.Start(); err != nil {
		p.w.Close()
		p.r.Close()
		return
	}
	cmd.Process.Release()
	return
}

func (c *visCmd) forwardMsgs(w io.Writer, msgs []vis.Msg) {
	w.Write([]byte(string(vis.MustEncode(msgs)) + "\n"))
}

func (c *visCmd) runServer(srv *vis.Server, errCh chan error) {
	errCh <- srv.Serve()
}

func (c *visCmd) processOutput(r io.Reader, sink vis.MessageSink, errCh chan error) {
	decoder := vis.NewMsgDecoder(r)
	for {
		msgs, err := decoder.Decode()
		if err == io.EOF {
			break
		}
		if err != nil {
			errCh <- err
			break
		}
		sink.RecvMessages(msgs)
	}
}
