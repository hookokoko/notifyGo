package internal

type TaskInfo struct {
	LogId           int64
	SendChannel     string // 消息渠道，比如是短信、邮件、推送等
	MessageContent  string
	MessageReceiver string
}

type Task struct {
	MsgId       int64
	SendChannel string // 消息渠道，比如是短信、邮件、推送等
	MsgContent  string
	MsgReceiver Target
}

type Target struct {
	UserId int64
	Email  string
	Phone  string
}
