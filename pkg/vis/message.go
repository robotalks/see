package vis

import (
	"encoding/json"
	"io"
)

// Msg is message with flex field
type Msg map[string]interface{}

// MustEncode encodes message to JSON
func (m Msg) MustEncode() string {
	encoded, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(encoded)
}

// Action returns action property
func (m Msg) Action() string {
	return stringProp(m, PropAction)
}

// Object returns object property
func (m Msg) Object() Object {
	val := m[PropObject]
	if obj, ok := val.(Object); ok {
		return obj
	}
	if objMap, ok := val.(map[string]interface{}); ok {
		return Object(objMap)
	}
	return nil
}

// Value returns the value property
func (m Msg) Value() DataValue {
	val, ok := m[PropValue]
	if !ok {
		return nil
	}
	data, _ := json.Marshal(val)
	return DataValue(data)
}

// ID returns id property
func (m Msg) ID() string {
	return stringProp(m, PropID)
}

// ByKeys retrieves a value following the keys
func (m Msg) ByKeys(keys ...string) (v interface{}) {
	if len(keys) == 0 {
		return m
	}
	if v = m[keys[0]]; v == nil {
		return
	}

	keys = keys[1:]
	for _, key := range keys {
		if dict, ok := v.(map[string]interface{}); ok {
			v = dict[key]
		} else {
			return nil
		}
		if v == nil {
			break
		}
	}
	return
}

// DataValue is the value of a data entry.
type DataValue json.RawMessage

func DataValueMsg(id string, val DataValue) Msg {
	return Msg{
		PropAction: ActionData,
		PropID:     id,
		PropValue:  json.RawMessage(val),
	}
}

// Object is an object containing arbitrary fields
type Object map[string]interface{}

// ID returns object id
func (o Object) ID() string {
	return stringProp(o, PropID)
}

// ObjectMsg creates an object message
func ObjectMsg(obj Object) Msg {
	return Msg{
		PropAction: ActionObject,
		PropObject: obj,
	}
}

// MustEncode encodes data to JSON
func MustEncode(data interface{}) []byte {
	encoded, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return encoded
}

func stringProp(m map[string]interface{}, prop string) string {
	val := m[prop]
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

// MessageSink defines a message receiver
type MessageSink interface {
	RecvMessages([]Msg)
}

type funcMessageSink struct {
	fn func([]Msg)
}

func (s *funcMessageSink) RecvMessages(msgs []Msg) {
	s.fn(msgs)
}

// SinkMessage wraps a func to be a MessageSink
func SinkMessage(sink func([]Msg)) MessageSink {
	return &funcMessageSink{fn: sink}
}

// MsgSource is the source of message, and also accepts messages
type MsgSource interface {
	MessageSink
	// ProcessMessages is a loop to drain messages and send to
	// provided MessageSink. The returned error can be io.EOF to
	// indicate end of message processing, or any other errors to
	// leave for application to determine call ProcessMessages again
	// or simply abort. Returning nil usually indicates there are
	// more messages, and application will call ProcessMessages again
	ProcessMessages(MessageSink) error
}

// Properties and Action names
const (
	PropAction   = "action"
	PropObject   = "object"
	PropValue    = "value"
	PropData     = "data"
	PropID       = "id"
	ActionReset  = "reset"
	ActionObject = "object"
	ActionData   = "data"
	ActionAsset  = "asset"
	ActionRemove = "remove"
)

// MsgDecoder decodes message from a stream
type MsgDecoder struct {
	decoder *json.Decoder
}

// NewMsgDecoder creates a decoder from a stream
func NewMsgDecoder(stream io.Reader) *MsgDecoder {
	return &MsgDecoder{decoder: json.NewDecoder(stream)}
}

// Decode decodes a list of messages
func (d *MsgDecoder) Decode() (msgs []Msg, err error) {
	err = d.decoder.Decode(&msgs)
	return
}
