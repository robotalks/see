package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	logger "github.com/op/go-logging"
	vis "github.com/robotalks/simulator/vis"
	mqtt "github.com/robotalks/simulator/vis/mqtt"
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

	var source vis.MsgSource
	switch {
	case len(args) == 0:
		source = &vis.StreamMsgSource{Reader: os.Stdin, Writer: os.Stdout}
	case strings.HasPrefix(args[0], "mqtt://"):
		src, err := mqtt.NewMsgSourceFromURL("tcp" + args[0][4:])
		if err != nil {
			return err
		}
		if len(args) > 1 {
			src.ClientID = args[1]
		}
		if err = src.Connect(); err != nil {
			return err
		}
		source = src
	default:
		src, err := vis.NewExecMsgSource(args[0], args[1:]...)
		if err != nil {
			return err
		}
		if err = src.Cmd.Start(); err != nil {
			return err
		}
		src.Cmd.Release()
		source = src
	}
	srv.MsgSink = source

	c.logger.Noticef("Listen %s", ln.(*net.TCPListener).Addr().String())

	errCh := make(chan error)
	go c.runServer(srv, errCh)
	go c.processMsgs(source, srv, errCh)
	err = <-errCh
	if err == io.EOF {
		err = nil
	}
	return err
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

func (c *visCmd) runServer(srv *vis.Server, errCh chan error) {
	errCh <- srv.Serve()
}

func (c *visCmd) processMsgs(source vis.MsgSource, sink vis.MessageSink, errCh chan error) {
	for {
		err := source.ProcessMessages(sink)
		if err != nil {
			errCh <- err
			break
		}
	}
}
