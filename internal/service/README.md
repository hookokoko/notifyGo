业务相关的处理逻辑放到这里

整个流程抽象成core，即

```
Send(Target, Content)
```

Target即发送的目标，Content即发送的内容，Send是发送的动作。

分别抽象出3个服务
