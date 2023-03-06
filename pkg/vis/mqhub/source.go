package mqhub

import (
	"fmt"
	"io"
	"net/http"
	"os"

	hub "github.com/robotalks/mqhub.go/mqhub"
	// load mqtt impl
	_ "github.com/robotalks/mqhub.go/mqtt"
	"github.com/robotalks/see/pkg/vis"
)

// MsgSource implements MsgSource backed by mqhub
type MsgSource struct {
	Connector hub.Connector
	Schema    *Schema

	msgCh chan hub.Message
}

// NewMsgSource creates MsgSource
func NewMsgSource(mqttURL, schemaFile string) (s *MsgSource, err error) {
	s = &MsgSource{msgCh: make(chan hub.Message)}
	if s.Connector, err = hub.NewConnector(mqttURL); err != nil {
		return
	}
	if s.Schema, err = LoadSchemaFile(schemaFile); err != nil {
		return
	}
	return
}

// Connect connects to MQTT
func (s *MsgSource) Connect() error {
	if err := s.Connector.Connect().Wait(); err != nil {
		return err
	}
	_, err := s.Connector.Watch(hub.MessageSinkFunc(s.handleMsg))
	return err
}

// RecvMessages implements vis.MessageSink
func (s *MsgSource) RecvMessages(msgs []vis.Msg) {
	for _, msg := range msgs {
		action := msg.Action()
		if action == "" {
			continue
		}
		sch := s.Schema.FindAction(msg)
		if sch == nil {
			continue
		}
		data, err := sch.Render(s.Schema.context, msg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "render data for reactor %s/%s error: %v\n",
				sch.Component, sch.Endpoint, err)
			continue
		}
		s.Connector.
			Describe(sch.Component).
			Endpoint(sch.Endpoint).
			ConsumeMessage(hub.StreamMessage(data))
	}
}

// ProcessMessages implements vis.MsgSource
func (s *MsgSource) ProcessMessages(sink vis.MessageSink) error {
	sink.RecvMessages(s.Schema.Refresh())
	for {
		msg, ok := <-s.msgCh
		if !ok {
			return io.EOF
		}
		component := msg.Component()
		endpoint := msg.Endpoint()
		encoded, ok := msg.(hub.EncodedPayload)
		if component == "" || endpoint == "" || !ok {
			continue
		}
		payload, err := encoded.Payload()
		if err != nil {
			continue
		}
		msgs := s.Schema.UpdateObject(component, endpoint, payload)
		if msgs != nil {
			sink.RecvMessages(msgs)
		}
	}
}

// AddHandlers implements ServerExt
func (s *MsgSource) AddHandlers(mux *http.ServeMux) error {
	mux.Handle("/mqhub/states/",
		http.StripPrefix("/mqhub/states/", http.HandlerFunc(s.serveStates)))
	return nil
}

func (s *MsgSource) handleMsg(msg hub.Message) hub.Future {
	s.msgCh <- msg
	return nil
}

func (s *MsgSource) serveStates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("only GET is allowed"))
		return
	}
	sch := s.Schema.FindStateSchema(r.URL.Path)
	if sch != nil {
		if val, exist := s.Schema.FindState(sch.Component, sch.Endpoint); exist {
			contentType := sch.ContentType
			if contentType == "" {
				contentType = "application/octet-stream"
			}
			w.Header().Add("Content-type", contentType)
			w.Write(val)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("not found"))
	return
}
