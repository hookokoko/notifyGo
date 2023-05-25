package model

import (
	"context"
	"notifyGo/internal"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
)

type notifyGoDAO struct {
	engine *xorm.Engine
}

// 之所以设计成返回接口的只有一个原因 就是不想暴露结构体
func NewINotifyGoDAO() INotifyGoDAO {
	engine, err := xorm.NewEngine("mysql", "root:@/notify_go?charset=utf8")
	if err != nil {
		return nil
	}
	return &notifyGoDAO{engine}
}

func NewITemplateDAO() ITemplateDAO {
	engine, err := xorm.NewEngine("mysql", "root:@/notify_go?charset=utf8")
	if err != nil {
		return nil
	}
	return &notifyGoDAO{engine}
}

type INotifyGoDAO interface {
	InsertRecord(ctx context.Context, templateId int64, target internal.ITarget, msgContent string) error
}

type ITemplateDAO interface {
	GetContent(templateId int64, country string) (string, error)
}

func (n *notifyGoDAO) InsertRecord(ctx context.Context, templateId int64, target internal.ITarget,
	msgContent string) error {
	sess := n.engine.NewSession()
	defer sess.Close()

	if err := sess.Begin(); err != nil {
		return err
	}

	// 一些字段要封装到参数里了
	delivery := Delivery{
		TemplateId:  templateId,
		Status:      1, // 消息创建状态
		SendChannel: 40,
		MsgType:     20,
		Proposer:    "crm",
		Creator:     "chenhaokun",
		Updator:     "chenhaokun",
		IsDelted:    0,
		Created:     time.Now(),
		Updated:     time.Now(),
	}

	// 这里配置好struct的id自增tag，会自动赋值插入的id
	if _, err := n.engine.Insert(&delivery); err != nil {
		return err
	}

	tgt := Target{
		TargetIdType: target.Type(), // 邮箱，这里封装一个枚举，需要根据 name(target.Email) -> value
		TargetId:     target.Value(),
		DeliveryId:   delivery.Id,
		Status:       1, // 创建状态
		MsgContent:   msgContent,
	}
	if _, err := n.engine.Insert(&tgt); err != nil {
		return err
	}

	return sess.Commit()
}

func (n *notifyGoDAO) GetContent(templateId int64, country string) (string, error) {
	// get 语言 by target
	// mock下查看对应target所在的国家
	tpl := Template{}
	has, err := n.engine.Where("id = ?", templateId).Get(&tpl)
	if err != nil || !has {
		return "", err
	}
	return tpl.ChsContent, nil
}

func (Delivery) TableName() string {
	return "delivery"
}

func (Target) TableName() string {
	return "target"
}

func (Template) TableName() string {
	return "template"
}

type Delivery struct {
	Id          int64     `xorm:"not null pk autoincr"`
	TemplateId  int64     `xorm:"INT"`
	Status      int       `xorm:"INT"`
	SendChannel int       `xorm:"comment('消息发送渠道 10.IM 20.Push 30.短信 40.Email 50.公众号') VARCHAR(255)"`
	MsgType     int       `xorm:"comment('10.通知类消息 20.营销类消息 30.验证码类消息') INT"`
	Proposer    string    `xorm:"comment('业务方') VARCHAR(255)"`
	Creator     string    `xorm:"VARCHAR(255)"`
	Updator     string    `xorm:"VARCHAR(255)"`
	IsDelted    int       `xorm:"INT"`
	Created     time.Time `xorm:"TIMESTAMP"`
	Updated     time.Time `xorm:"TIMESTAMP"`
}

type Target struct {
	Id           int64  `xorm:"not null pk INT"`
	TargetIdType int    `xorm:"comment('接收目标id类型') INT"`
	TargetId     string `xorm:"comment('接收目标id') VARCHAR(255)"`
	DeliveryId   int64  `xorm:"INT"`
	Status       int    `xorm:"INT"`
	MsgContent   string `xorm:"TEXT"`
}

type Template struct {
	Id         int64     `xorm:"not null pk INT"`
	Country    string    `xorm:"VARCHAR(255)"`
	Type       int       `xorm:"comment('sms|email|push') INT"`
	EnContent  string    `xorm:"TEXT"`
	ChsContent string    `xorm:"TEXT"`
	ChtContent string    `xorm:"TEXT"`
	Creator    string    `xorm:"VARCHAR(255)"`
	Updator    string    `xorm:"VARCHAR(255)"`
	IsDelted   int       `xorm:"INT"`
	Created    time.Time `xorm:"TIMESTAMP"`
	Updated    time.Time `xorm:"TIMESTAMP"`
}
