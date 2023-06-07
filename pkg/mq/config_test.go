package mq

import (
	"encoding/json"
	"log"
	"notifyGo/internal"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Config(t *testing.T) {
	cfg := NewConfig("/Users/hooko/GolandProjects/notifyGo/config/kafka_topic.toml")
	log.Printf("%+v\n", cfg)
	var tps map[string]map[string]string
	tB, err := json.Marshal(cfg.Topics)
	assert.Nil(t, err)
	err = json.Unmarshal(tB, &tps)
	assert.Nil(t, err)
	log.Printf("%+v\n", tps)

	match, ok := tps["email1"]["verification"]
	if !ok {
		log.Fatal("找不到对应topic")
	}
	log.Println(match)
}

func Test_Unmarshal(t *testing.T) {
	value := `{"task_id": 123, "send_channel": "email", "msg_content": {"type":"notification", "content": "hello"}, "msg_receiver": {"email": "648646891@qq.com"}}`
	taskInfo := new(Task)
	err := json.Unmarshal([]byte(value), taskInfo)
	log.Println(err)
	log.Printf("%+v\n", taskInfo)
}

type Task struct {
	TaskId      int64               `json:"task_id"`
	SendChannel string              `json:"send_channel"` // 消息渠道，比如是短信、邮件、推送等
	MsgContent  internal.MsgContent `json:"msg_content"`
	MsgReceiver internal.ITarget    `json:"msg_receiver"`
}

func (t *Task) UnmarshalJSON(data []byte) error {
	var tmp map[string]interface{}
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}
	t.TaskId = int64(tmp["task_id"].(float64))
	t.SendChannel = tmp["send_channel"].(string)
	var msgContent internal.MsgContent
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
			var email internal.EmailTarget
			email.Email = v.(string)
			t.MsgReceiver = email
		case "phone":
			var phone internal.PhoneTarget
			phone.Phone = v.(string)
			t.MsgReceiver = phone
		case "id":
			var id internal.IdTarget
			id.Id = v.(string)
			t.MsgReceiver = id
		default:
		}
	}
	return nil
}
