package internal

import "encoding/json"

type TaskInfo struct {
	LogId           int64
	SendChannel     string // 消息渠道，比如是短信、邮件、推送等
	MessageContent  string
	MessageReceiver string
}

type Task struct {
	TaskId      int64      `json:"task_id"`
	SendChannel string     `json:"send_channel"` // 消息渠道，比如是短信、邮件、推送等
	MsgContent  MsgContent `json:"msg_content"`
	MsgReceiver ITarget    `json:"msg_receiver"`
}

type MsgContent struct {
	Type    string `json:"type"` // 消息类型，比如营销、验证码、通知
	Content string `json:"content"`
}

type ITarget interface {
	Type() int
	Value() string
}

const (
	TYPEEmail = iota + 10
	TYPEPhone
	TYPEId
)

type EmailTarget struct {
	Email string `json:"email"`
}

func (e EmailTarget) Type() int {
	return TYPEEmail
}

func (e EmailTarget) Value() string {
	return e.Email
}

type PhoneTarget struct {
	Phone string `json:"phone"`
}

func (e PhoneTarget) Type() int {
	return TYPEPhone
}

func (e PhoneTarget) Value() string {
	return e.Phone
}

type IdTarget struct {
	Id string `json:"id"`
}

func (e IdTarget) Type() int {
	return TYPEId
}

func (e IdTarget) Value() string {
	return e.Id
}

func (t *Task) UnmarshalJSON(data []byte) error {
	var tmp map[string]interface{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	t.TaskId = int64(tmp["task_id"].(float64))
	t.SendChannel = tmp["send_channel"].(string)
	var msgContent MsgContent
	bm, err := json.Marshal(tmp["msg_content"])
	if err != nil {
		return err
	}
	err = json.Unmarshal(bm, &msgContent)
	if err != nil {
		return err
	}
	t.MsgContent = msgContent

	m := tmp["msg_receiver"].(map[string]interface{})
	for k, v := range m {
		switch k {
		case "email":
			var email EmailTarget
			email.Email = v.(string)
			t.MsgReceiver = email
		case "phone":
			var phone PhoneTarget
			phone.Phone = v.(string)
			t.MsgReceiver = phone
		case "id":
			var id IdTarget
			id.Id = v.(string)
			t.MsgReceiver = id
		default:
		}
	}
	return nil
}
