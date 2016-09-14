package vis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	rice "github.com/GeertJohan/go.rice"
	logger "github.com/op/go-logging"
	websocket "golang.org/x/net/websocket"
	yaml "gopkg.in/yaml.v2"
)

// PageContext defines the extensive content in page
type PageContext struct {
	Stylesheets []string `json:"stylesheets" yaml:"stylesheets"`
	Scripts     []string `json:"scripts" yaml:"scripts"`
}

// PluginManifest is the content of plugin manifest file
type PluginManifest struct {
	Name       string      `json:"name" yaml:"name"`
	Visualizer PageContext `json:"visualizer" yaml:"visualizer"`
}

const (
	// PluginManifestFile is the filename of plugin manifest
	PluginManifestFile = "visualizer.plugin"
)

// LoadPluginManifest loads plugin manifest from specified directory
func LoadPluginManifest(dir string) (*PluginManifest, error) {
	fn := filepath.Join(dir, PluginManifestFile)
	raw, err := ioutil.ReadFile(fn)
	if err != nil {
		return nil, err
	}

	manifest := &PluginManifest{}
	// json file
	if bytes.HasPrefix(bytes.TrimSpace(raw), []byte{'{'}) {
		err = json.Unmarshal(raw, manifest)
	} else {
		err = yaml.Unmarshal(raw, manifest)
	}
	return manifest, err
}

type plugin struct {
	name    string
	dir     string
	fullDir string
}

// Server serve static pages and APIs
type Server struct {
	Host          string
	Port          int
	Listener      net.Listener
	States        StateStore
	MsgSink       MessageSink
	Logger        *logger.Logger
	WebContentDir string

	plugins []*plugin
	conns   map[*websocket.Conn]*websocket.Conn
	lock    sync.RWMutex
}

// LoadPlugin loads plugin from specified directory
func (s *Server) LoadPlugin(dir string) error {
	var name string
	pos := strings.Index(dir, ":")
	if pos > 0 {
		name = dir[0:pos]
		dir = dir[pos+1:]
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("%s: locate error: %v", dir, err)
	}
	mf, err := LoadPluginManifest(absDir)
	if err != nil {
		return fmt.Errorf("%s: load error: %v", dir, err)
	}
	if name == "" {
		name = mf.Name
	}
	if name == "" {
		name = filepath.Base(absDir)
	}

	for _, p := range s.plugins {
		if p.name == name {
			return fmt.Errorf("%s: name '%s' conflict with %s", dir, name, p.dir)
		}
	}

	s.plugins = append(s.plugins, &plugin{name: name, dir: dir, fullDir: absDir})
	return nil
}

// Serve runs the server
func (s *Server) Serve() error {
	h, err := s.Handler()
	if err != nil {
		return err
	}
	if s.Listener != nil {
		return http.Serve(s.Listener, h)
	}
	return http.ListenAndServe(fmt.Sprintf("%s:%d", s.Host, s.Port), h)
}

// Handler creates default http handler
func (s *Server) Handler() (http.Handler, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/objects", s.StatesHandler)
	mux.Handle("/ws", websocket.Handler(s.WebSocketHandler))
	for _, p := range s.plugins {
		prefix := "/plugins/" + p.name + "/"
		mux.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(p.fullDir))))
	}
	var fs http.FileSystem
	if s.WebContentDir != "" {
		s.Logger.Infof("Use Web Content: %s", s.WebContentDir)
		fs = http.Dir(s.WebContentDir)
	} else {
		box, err := rice.FindBox("www")
		if err != nil {
			return nil, err
		}
		fs = box.HTTPBox()
	}
	fsHandler := http.FileServer(fs)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet &&
			(r.URL.Path == "/" || r.URL.Path == "index.html") {
			s.HandleIndex(w, r, fs)
		} else {
			fsHandler.ServeHTTP(w, r)
		}
	})
	return mux, nil
}

// HandleIndex generates index HTML file
func (s *Server) HandleIndex(w http.ResponseWriter, r *http.Request, fs http.FileSystem) {
	content, err := s.GenerateIndexPage(fs)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(err.Error()))
		return
	}
	w.Write([]byte(content))
}

// GenerateIndexPage generates index.html
func (s *Server) GenerateIndexPage(fs http.FileSystem) (string, error) {
	f, err := fs.Open("/index.html")
	if err != nil {
		return "", err
	}
	raw, err := ioutil.ReadAll(f)
	f.Close()
	if err != nil {
		return "", err
	}
	tmpl, err := template.New("index").Parse(string(raw))
	if err != nil {
		return "", err
	}
	var ctx PageContext
	for _, p := range s.plugins {
		mf, mfErr := LoadPluginManifest(p.fullDir)
		if mfErr != nil {
			s.Logger.Warningf("Load plugin %s (%s) failed: %v", p.name, p.dir, mfErr)
			continue
		}
		for _, fn := range mf.Visualizer.Stylesheets {
			ctx.Stylesheets = append(ctx.Stylesheets, "plugins/"+p.name+"/"+fn)
		}
		for _, fn := range mf.Visualizer.Scripts {
			ctx.Scripts = append(ctx.Scripts, "plugins/"+p.name+"/"+fn)
		}
	}
	var out bytes.Buffer
	err = tmpl.Execute(&out, &ctx)
	return out.String(), err
}

// WebSocketHandler handles websocket connections
func (s *Server) WebSocketHandler(ws *websocket.Conn) {
	s.addConn(ws)
	defer s.rmConn(ws)
	objs, err := s.Objects()
	if err != nil {
		s.Logger.Errorf("States error: %v", err)
		return
	}
	msgs := make([]Msg, 0, len(objs))
	for _, obj := range objs {
		msgs = append(msgs, ObjectMsg(obj))
	}
	_, err = ws.Write(MustEncode(msgs))
	if err != nil {
		s.Logger.Errorf("Write error: %v", err)
		return
	}

	decoder := NewMsgDecoder(ws)
	for {
		var msgs []Msg
		msgs, err = decoder.Decode()
		if err != nil && err != io.EOF {
			s.Logger.Errorf("Read message error: %v", err)
		}
		if err != nil {
			return
		}
		for _, msg := range msgs {
			s.Logger.Infof("%s: %s", strings.ToUpper(msg.Action()), msg.MustEncode())
		}
		// forward to message sink
		if s.MsgSink != nil {
			s.MsgSink.RecvMessages(msgs)
		}
	}
}

func (s *Server) addConn(ws *websocket.Conn) {
	s.lock.Lock()
	if s.conns == nil {
		s.conns = make(map[*websocket.Conn]*websocket.Conn)
	}
	s.conns[ws] = ws
	s.lock.Unlock()
}

func (s *Server) rmConn(ws *websocket.Conn) {
	s.lock.Lock()
	if s.conns != nil {
		if s.conns[ws] == ws {
			delete(s.conns, ws)
		}
	}
	s.lock.Unlock()
	ws.Close()
}

func (s *Server) broadcastMessages(msgs []Msg) {
	encoded := MustEncode(msgs)
	var conns []*websocket.Conn
	s.lock.RLock()
	if s.conns != nil {
		for conn := range s.conns {
			conns = append(conns, conn)
		}
	}
	s.lock.RUnlock()
	for _, conn := range conns {
		conn.Write(encoded)
	}
}

// StatesHandler is the http handler manipulate object states
func (s *Server) StatesHandler(w http.ResponseWriter, r *http.Request) {
	var objects map[string]Object
	var err error
	switch r.Method {
	case http.MethodGet:
		objects, err = s.Objects()
	case http.MethodPost, http.MethodPut:
		var msgs []Msg
		msgs, err = NewMsgDecoder(r.Body).Decode()
		if err == nil {
			s.RecvMessages(msgs)
		}
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else if objects != nil {
		w.WriteHeader(http.StatusOK)
		w.Write(MustEncode(objects))
	} else {
		w.WriteHeader(http.StatusNoContent)
		w.Write(nil)
	}
}

// RecvMessages implements MessageSink
func (s *Server) RecvMessages(msgs []Msg) {
	for _, msg := range msgs {
		s.HandleMessage(msg)
	}
	s.broadcastMessages(msgs)
}

// HandleMessage processes one message
func (s *Server) HandleMessage(a Msg) (err error) {
	action := a.Action()
	switch action {
	case ActionObject:
		if obj := a.Object(); obj != nil {
			err = s.Update(obj)
		} else {
			err = fmt.Errorf("missing property object")
		}
	case ActionRemove:
		err = s.Remove(a.ID())
	default:
		err = fmt.Errorf("unknown action")
	}
	if err == nil {
		s.Logger.Infof("%s: %s", strings.ToUpper(action), a.MustEncode())
	} else {
		s.Logger.Errorf("%s: %s: %s", strings.ToUpper(action), err.Error(), a.MustEncode())
	}
	return
}

// Objects implements StateStore
func (s *Server) Objects() (map[string]Object, error) {
	return s.States.Objects()
}

// Update implements StateStore
func (s *Server) Update(objs ...Object) error {
	return s.States.Update(objs...)
}

// Remove implements StateStore
func (s *Server) Remove(ids ...string) error {
	return s.States.Remove(ids...)
}
