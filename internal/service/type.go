package service

type MessageParam struct {
	Receiver  string                 `json:"receiver"`           // 接收者
	Content   string                 `json:"content"`            // 消息内容
	Variables map[string]interface{} `json:"variables,optional"` // 可选 消息内容中的可变部分(占位符替换)
}
