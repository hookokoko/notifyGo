package internal

type TaskInfo struct {
	LogId           uint64
	SendChannel     string // 消息渠道，比如是短信、邮件、推送等
	MessageContent  string
	MessageReceiver string
}

type Task struct {
	MsgId       uint64
	SendChannel string // 消息渠道，比如是短信、邮件、推送等
	MsgContent  string
	MsgReceiver []Target
}

type Target struct {
	UserId uint64
	Email  string
	Phone  string
}
