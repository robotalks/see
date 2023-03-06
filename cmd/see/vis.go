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
	vis "github.com/robotalks/see/pkg/vis"
	"github.com/robotalks/see/pkg/vis/mqhub"
	mqtt "github.com/robotalks/see/pkg/vis/mqtt"
)

// Version number
const Version = "0.1.0"

// VersionSuffix provides suffix info
var VersionSuffix = "-dev"

// PrintVersion prints version
func PrintVersion() {
	fmt.Println(Version + VersionSuffix)
}

type visCmd struct {
	Port       int
	Quiet      bool
	PluginDirs []string `n:"plugin-dir"`
	Title      string
	Version    bool

	logger *logger.Logger
}

func (c *visCmd) Execute(args []string) error {
	if c.Version {
		PrintVersion()
		return nil
	}

	c.logger = logger.MustGetLogger("see")
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
		Title:         c.Title,
		LocalWebDir:   ".vis.www",
		WebContentDir: os.Getenv("SEE_WEB_ROOT"),
		Logger:        c.logger,
	}

	if err = c.loadPlugins(srv); err != nil {
		return err
	}

	var source vis.MsgSource
	switch {
	case len(args) == 0:
		source = &vis.StreamMsgSource{Reader: os.Stdin, Writer: os.Stdout}
	case strings.HasPrefix(args[0], "mqhub://"):
		if len(args) < 2 {
			return fmt.Errorf("mqhub expects schema file as second argument")
		}
		src, e := mqhub.NewMsgSource("mqtt"+args[0][5:], args[1])
		if e != nil {
			return e
		}
		if err = src.Connect(); err != nil {
			return err
		}
		source = src
	case strings.HasPrefix(args[0], "mqtt://"):
		src, e := mqtt.NewMsgSourceFromURL("tcp" + args[0][4:])
		if e != nil {
			return e
		}
		if len(args) > 1 {
			src.ClientID = args[1]
		}
		if err = src.Connect(); err != nil {
			return err
		}
		source = src
	default:
		if len(args) == 0 || args[0] == "" {
			args = []string{"./.vis.exec"}
		}
		src, e := vis.NewExecMsgSource(args[0], args[1:]...)
		if e != nil {
			return e
		}
		if err = src.Cmd.Start(); err != nil {
			return err
		}
		src.Cmd.Process.Release()
		source = src
	}
	srv.MsgSink = source

	c.logger.Noticef("Listen %s", ln.(*net.TCPListener).Addr().String())

	errCh := make(chan error)
	go c.runServer(source, srv, errCh)
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
		srv.LoadPlugin(filepath.Join(usr.HomeDir, ".robotalks"))
	}
	wd, err := os.Getwd()
	if err == nil {
		srv.LoadPlugin(wd)
	}
	if dirs := os.Getenv("SEE_PLUGIN_PATH"); dirs != "" {
		for _, dir := range filepath.SplitList(dirs) {
			if srv.LoadPlugin(dir) == nil {
				c.logger.Infof("Loaded %s", dir)
			}
		}
	}
	for _, dir := range c.PluginDirs {
		if err = srv.LoadPlugin(dir); err != nil {
			return err
		}
		c.logger.Infof("Loaded %s", dir)
	}
	return nil
}

func (c *visCmd) runServer(ext interface{}, srv *vis.Server, errCh chan error) {
	srvExt, _ := ext.(vis.ServerExt)
	errCh <- srv.Serve(srvExt)
}

func (c *visCmd) processMsgs(source vis.MsgSource, sink vis.MessageSink, errCh chan error) {
	for {
		err := source.ProcessMessages(sink)
		if err != nil {
			if err != io.EOF {
				errCh <- err
			}
			break
		}
	}
}
