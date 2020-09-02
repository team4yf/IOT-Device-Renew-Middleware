//Package message the message data struct defination
package message

//Header the header of the message
type Header struct {
	Version   int    `json:"v"`
	NameSpace string `json:"ns"`
	Name      string `json:"name"`
	AppID     string `json:"appId"`
	ProjID    int64  `json:"projId"`
	Source    string `json:"source"`
}

//Device the device data struct of the payload
type Device struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Name    string                 `json:"name"`
	Brand   string                 `json:"brand"`
	Version string                 `json:"v"`
	Extra   map[string]interface{} `json:"x,omitempty"`
}

//D2SPayload the payload data struct
type D2SPayload struct {
	Device    *Device     `json:"device"`
	Data      interface{} `json:"data"`
	Cgi       string      `json:"cgi"`
	Timestamp int64       `json:"timestamp"`
}

//D2SFeedback the body of the feedback message
type D2SFeedback struct {
	Result    interface{} `json:"result"`
	MsgID     string      `json:"msgId"`
	Cgi       string      `json:"cgi"`
	Timestamp int64       `json:"timestamp"`
}

//RenewPayload the body of the renew message
type RenewPayload struct {
	DeviceID  string `json:"sn"`
	Expire    int64  `json:"expire"`
	Cgi       string `json:"cgi"`
	Timestamp int64  `json:"timestamp"`
}

//MsgPayload the body of the msg message
type MsgPayload struct {
	DeviceID  string `json:"sn"`
	MsgID     string `json:"msgId"`
	Expire    int64  `json:"expire"`
	Cgi       string `json:"cgi"`
	Timestamp int64  `json:"timestamp"`
}

//S2DPayload the payload data struct
type S2DPayload struct {
	Device    *Device     `json:"device"`
	Argument  interface{} `json:"arg"`
	MsgID     string      `json:"msgId"`
	NetID     string      `json:"netId"`
	Cmd       string      `json:"cmd"`
	Cgi       string      `json:"cgi"`
	Feedback  int         `json:"feedback,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

//D2SMessage device to server message
type D2SMessage struct {
	Header  *Header     `json:"header"`
	Payload *D2SPayload `json:"payload"`
}

//RenewMessage device to server message
type RenewMessage struct {
	Header  *Header       `json:"header"`
	Payload *RenewPayload `json:"payload"`
}

//MsgMessage device to server message
type MsgMessage struct {
	Header  *Header     `json:"header"`
	Payload *MsgPayload `json:"payload"`
}

//S2DMessage server to device message
type S2DMessage struct {
	Header  *Header                `json:"header"`
	Bind    map[string]interface{} `json:"bind,omitempty"`
	Payload []*S2DPayload          `json:"payload"`
}

//D2SFeedbackMessage feedback message
type D2SFeedbackMessage struct {
	Header   *Header      `json:"header"`
	Feedback *D2SFeedback `json:"feedback"`
}
