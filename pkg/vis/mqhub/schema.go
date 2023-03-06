package mqhub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"text/template"

	"github.com/easeway/langx.go/mapper"
	"github.com/robotalks/see/pkg/vis"
	yaml "gopkg.in/yaml.v3"
)

// StateSchema defines a cached state
type StateSchema struct {
	Component   string `json:"component"`
	Endpoint    string `json:"endpoint"`
	APIPath     string `json:"api-path"`
	ContentType string `json:"content-type"`
}

// MuteSchema defines states without notifying object change
type MuteSchema struct {
	Component string `json:"component"`
	Endpoint  string `json:"endpoint"`
}

// ActionSchema maps action to reactor on mqhub
type ActionSchema struct {
	Action    string          `json:"action"`
	Matches   []ActionMatcher `json:"matches"`
	Component string          `json:"component"`
	Endpoint  string          `json:"endpoint"`
	Template  string          `json:"data"`

	template *template.Template
}

// ActionMatcher matches an action
type ActionMatcher struct {
	Keys  []string `json:"keys"`
	Value string   `json:"value"`
}

// Init initializes internal states of ActionSchema
func (s *ActionSchema) Init() error {
	t, err := template.New("data").Funcs(template.FuncMap{
		"scalef": func(val, factor float64) float64 {
			return val * factor
		},
		"scalei": func(val, factor float64) int64 {
			return int64(val * factor)
		},
	}).Parse(s.Template)
	if err != nil {
		return err
	}
	s.template = t.Templates()[0]
	return nil
}

// MatchAction determine if this schema matches the action
func (s *ActionSchema) MatchAction(msg vis.Msg) bool {
	if s.Action != msg.Action() {
		return false
	}
	for _, m := range s.Matches {
		v := msg.ByKeys(m.Keys...)
		if v == nil {
			return false
		}
		if fmt.Sprintf("%v", v) != m.Value {
			return false
		}
	}
	return true
}

// Render renders the data template
func (s *ActionSchema) Render(ctx *Context, msg vis.Msg) ([]byte, error) {
	var buf bytes.Buffer
	err := s.template.Execute(&buf, map[string]interface{}{
		"action": msg,
		"ctx":    ctx,
	})
	return buf.Bytes(), err
}

// TopSchema is the top-level schema
type TopSchema struct {
	Template string          `json:"objects"`
	States   []*StateSchema  `json:"states"`
	Muted    []*MuteSchema   `json:"mute"`
	Actions  []*ActionSchema `json:"actions"`
}

// IsMuted indicates and object/endpoint is muted
func (s *TopSchema) IsMuted(component, endpoint string) bool {
	for _, m := range s.Muted {
		if m.Component == component && m.Endpoint == endpoint {
			return true
		}
	}
	return false
}

// Context is the current rendering context
type Context struct {
	Objects map[string]vis.Object
	States  map[string]map[string][]byte

	lock sync.RWMutex
}

// NewSchemaCtx creates a Context
func NewSchemaCtx() *Context {
	return &Context{
		Objects: make(map[string]vis.Object),
		States:  make(map[string]map[string][]byte),
	}
}

// UpdateProperty updates a single property of an object
func (c *Context) UpdateProperty(id, property string, value []byte) {
	var parsed interface{}
	if json.Unmarshal(value, &parsed) != nil {
		parsed = nil
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	obj := c.Objects[id]
	if obj == nil {
		obj = make(vis.Object)
		c.Objects[id] = obj
	}

	if parsed != nil {
		obj[property] = parsed
	} else {
		delete(obj, property)
	}

	s := c.States[id]
	if s == nil {
		s = make(map[string][]byte)
		c.States[id] = s
	}
	s[property] = value
}

// FindState gets the raw state of specified object/property
func (c *Context) FindState(id, property string) (val []byte, exist bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if s := c.States[id]; s != nil {
		val, exist = s[property]
	}
	return
}

// Render applies the context
func (c *Context) Render(t *template.Template) ([]vis.Object, error) {
	var buf bytes.Buffer
	c.lock.RLock()
	err := t.Execute(&buf, c)
	c.lock.RUnlock()
	if err != nil {
		return nil, err
	}
	var objs []vis.Object
	if err = json.Unmarshal(buf.Bytes(), &objs); err != nil {
		return nil, err
	}
	return objs, nil
}

func normalizeMap(val interface{}) interface{} {
	switch v := val.(type) {
	case []interface{}:
		for n, item := range v {
			v[n] = normalizeMap(item)
		}
	case map[interface{}]interface{}:
		m := make(map[string]interface{})
		for key, value := range v {
			m[fmt.Sprintf("%v", key)] = normalizeMap(value)
		}
		val = m
	case map[string]interface{}:
		for key, value := range v {
			v[key] = normalizeMap(value)
		}
	}
	return val
}

// Schema is a loaded schema
type Schema struct {
	schema   TopSchema
	template *template.Template
	context  *Context
	current  []vis.Object
}

// LoadSchemaFile load schema from file
func LoadSchemaFile(filename string) (*Schema, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var out map[string]interface{}
	if err = yaml.Unmarshal(content, &out); err != nil {
		return nil, err
	}
	out = normalizeMap(out).(map[string]interface{})

	s := &Schema{context: NewSchemaCtx()}
	if err = mapper.Map(&s.schema, out); err != nil {
		return nil, err
	}

	t, err := template.New("schema").Funcs(template.FuncMap{
		"object": func(ctx *Context, id string) vis.Object {
			return ctx.Objects[id]
		},
	}).Parse(s.schema.Template)
	if err != nil {
		return nil, err
	}
	s.template = t.Templates()[0]

	for _, a := range s.schema.Actions {
		if err = a.Init(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Refresh applies the context and generates messages
func (s *Schema) Refresh() (msgs []vis.Msg) {
	objs, err := s.context.Render(s.template)
	if err != nil {
		fmt.Fprintf(os.Stderr, "render template failed: %v", err)
		return
	}
	activeObjs := make(map[string]bool)
	for _, obj := range objs {
		msg := make(vis.Msg)
		msg[vis.PropAction] = vis.ActionObject
		msg[vis.PropObject] = obj
		msgs = append(msgs, msg)
		activeObjs[obj.ID()] = true
	}

	if curr := s.current; curr != nil {
		for _, obj := range curr {
			if !activeObjs[obj.ID()] {
				msg := make(vis.Msg)
				msg[vis.PropAction] = vis.ActionRemove
				msg[vis.PropID] = obj.ID()
				msgs = append(msgs, msg)
			}
		}
	}
	s.current = objs
	return
}

// UpdateObject updates one property of the object and gets update messages
func (s *Schema) UpdateObject(id, property string, value []byte) []vis.Msg {
	s.context.UpdateProperty(id, property, value)
	if !s.schema.IsMuted(id, property) {
		return s.Refresh()
	}
	return nil
}

// FindStateSchema looks up the state schema by query path
func (s *Schema) FindStateSchema(requestPath string) *StateSchema {
	id, endpoint := path.Split(requestPath)
	id = strings.Trim(id, "/")
	for _, sch := range s.schema.States {
		if (sch.APIPath != "" && sch.APIPath == requestPath) ||
			(sch.APIPath == "" && sch.Component == id && sch.Endpoint == endpoint) {
			return sch
		}
	}
	return nil
}

// FindState gets the raw state of specified object/property
func (s *Schema) FindState(id, property string) ([]byte, bool) {
	return s.context.FindState(id, property)
}

// FindAction matches the message to action schema
func (s *Schema) FindAction(msg vis.Msg) *ActionSchema {
	for _, sch := range s.schema.Actions {
		if sch.MatchAction(msg) {
			return sch
		}
	}
	return nil
}
