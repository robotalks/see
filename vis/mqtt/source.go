package mqtt

import (
	"bytes"
	"io"
	"net/url"
	"strings"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/robotalks/see/vis"
	"github.com/rs/xid"
)

const (
	// MessagesTopic is the topic name for update messages
	MessagesTopic = "msgs"
	// EventsTopic is the topic name for events
	EventsTopic = "events"
)

// MsgSource processes messages from MQTT bus
type MsgSource struct {
	Server   *url.URL
	Prefix   string
	ClientID string
	Client   paho.Client

	msgCh chan paho.Message
}

// NewMsgSourceFromURL creates a MsgSource by parsing a URL
func NewMsgSourceFromURL(serverURL string) (s *MsgSource, err error) {
	s = &MsgSource{}
	if s.Server, err = url.Parse(serverURL); err != nil {
		return
	}
	s.Prefix = strings.Trim(s.Server.Path, "/")
	if s.Prefix != "" {
		s.Prefix += "/"
	}
	return
}

// Connect connects to MQTT
func (s *MsgSource) Connect() error {
	if s.msgCh == nil {
		s.msgCh = make(chan paho.Message)
	}
	if s.Client == nil {
		opts := paho.NewClientOptions()
		opts.Servers = append(opts.Servers, s.Server)
		if s.Server.User != nil {
			opts.Username = s.Server.User.Username()
			if pwd, present := s.Server.User.Password(); present {
				opts.Password = pwd
			}
		}
		opts.ClientID = s.ClientID
		if opts.ClientID == "" {
			opts.ClientID = xid.New().String()
		}
		s.Client = paho.NewClient(opts)
	}
	if s.Client.IsConnected() {
		return nil
	}
	token := s.Client.Connect()
	token.Wait()
	if err := token.Error(); err != nil {
		return err
	}
	token = s.Client.Subscribe(s.Prefix+MessagesTopic, 0, s.messageHandler)
	token.Wait()
	return token.Error()
}

func (s *MsgSource) messageHandler(_ paho.Client, msg paho.Message) {
	s.msgCh <- msg
}

// RecvMessages implements vis.MessageSink
func (s *MsgSource) RecvMessages(msgs []vis.Msg) {
	if client := s.Client; client != nil && client.IsConnected() {
		client.Publish(s.Prefix+EventsTopic, 0, false,
			[]byte(string(vis.MustEncode(msgs))))
	}
}

// ProcessMessages implements vis.MsgSource
func (s *MsgSource) ProcessMessages(sink vis.MessageSink) error {
	for {
		msg, ok := <-s.msgCh
		if !ok {
			return io.EOF
		}
		decoder := vis.NewMsgDecoder(bytes.NewBuffer(msg.Payload()))
		for {
			if msgs, err := decoder.Decode(); err == nil {
				sink.RecvMessages(msgs)
			} else if err != io.EOF {
				return err
			} else {
				break
			}
		}
	}
}
