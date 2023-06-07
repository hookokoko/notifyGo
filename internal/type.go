package internal

type TaskInfo struct {
	LogId           int64
	SendChannel     string // 消息渠道，比如是短信、邮件、推送等
	MessageContent  string
	MessageReceiver string
}

type Task struct {
	TaskId int64
	// 会根据 SendChannel 、MsgType 选择kafka topic
	SendChannel string // 消息渠道，比如是短信、邮件、推送等
	MsgContent  MsgContent
	MsgReceiver ITarget
}

type MsgContent struct {
	Type    string // 消息类型，比如营销、验证码、通知
	Content string
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
	Email string
}

func (e EmailTarget) Type() int {
	return TYPEEmail
}

func (e EmailTarget) Value() string {
	return e.Email
}

type PhoneTarget struct {
	Phone string
}

func (e PhoneTarget) Type() int {
	return TYPEPhone
}

func (e PhoneTarget) Value() string {
	return e.Phone
}

type IdTarget struct {
	Id string
}

func (e IdTarget) Type() int {
	return TYPEId
}

func (e IdTarget) Value() string {
	return e.Id
}
