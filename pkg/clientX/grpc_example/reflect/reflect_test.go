package reflect

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	"google.golang.org/grpc"
	"log"
	"notifyGo/pkg/clientX/grpc_example/helloworld"
	"testing"
)

func Test_Reflect(t *testing.T) {
	var parser protoparse.Parser
	fileDescriptors, _ := parser.ParseFiles("../helloworld/helloworld.proto")
	fileDescriptor := fileDescriptors[0]
	m := make(map[string]interface{})
	for _, msgDescriptor := range fileDescriptor.GetMessageTypes() {
		m[msgDescriptor.GetName()] = convertMessageToMap(msgDescriptor)
	}
	bs, _ := json.MarshalIndent(m, "", "\t")
	fmt.Println(string(bs))
}

func convertMessageToMap(message *desc.MessageDescriptor) map[string]interface{} {
	m := make(map[string]interface{})
	for _, fieldDescriptor := range message.GetFields() {
		fieldName := fieldDescriptor.GetName()
		if fieldDescriptor.IsRepeated() {
			// 如果是一个数组的话，就返回 nil 吧
			m[fieldName] = nil
			continue
		}
		switch fieldDescriptor.GetType() {
		case descriptor.FieldDescriptorProto_TYPE_MESSAGE:
			m[fieldName] = convertMessageToMap(fieldDescriptor.GetMessageType())
			continue
		}
		m[fieldName] = fieldDescriptor.GetDefaultValue()
	}
	return m
}

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

// makeGRPCRequest 根据指定的服务、方法和JSON主体调用gRPC服务并返回响应
//func makeGRPCRequest(serverAddress, serviceName, methodName, jsonBody string, headers http.Header) ([]byte, error) {
//	// 连接到gRPC服务器
//	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
//	if err != nil {
//		log.Fatalf("did not connect: %v", err)
//	}
//	defer conn.Close()
//
//	// 构建动态客户端
//	descSource := grpcdynamic.NewDiscoveryClient(conn)
//	caller := grpcdynamic.NewCaller(descSource)
//
//	// 找到服务和方法描述符
//	serviceDesc, err := descSource.FindSymbol(fmt.Sprintf("/%s/%s", serviceName, methodName))
//	if err != nil {
//		return nil, err
//	}
//	methodDesc := serviceDesc.(grpc.MethodDescriptor)
//
//	// 解析JSON消息
//	messageType := methodDesc.GetInputType()
//	message := dynamic.NewMessage(messageType)
//	err = jsonpb.UnmarshalString(jsonBody, message)
//	if err != nil {
//		return nil, err
//	}
//
//	// 设置元数据头部
//	md := metadata.New(headers)
//	ctx := metadata.NewOutgoingContext(context.Background(), md)
//
//	// 调用gRPC方法
//	ret, err := caller.CallMethod(ctx, conn, methodDesc, message)
//	if err != nil {
//		st, ok := status.FromError(err)
//		if ok {
//			return nil, fmt.Errorf("%s: %s", st.Code().String(), st.Message())
//		}
//		return nil, err
//	}
//
//	// 构造响应JSON
//	responseMessage := dynamic.NewMessage(methodDesc.GetOutputType())
//	err = proto.Unmarshal(ret, responseMessage)
//	if err != nil {
//		return nil, err
//	}
//	responseJSON, err := (&jsonpb.Marshaler{}).MarshalToString(responseMessage)
//	if err != nil {
//		return nil, err
//	}
//
//	return []byte(responseJSON), nil
//}

//func main() {
//	// 假设您有一个gRPC服务，您可以通过以下方式来创建一个动态客户端存根
//	// 服务的名称是MyService，您的方法名称为MyMethod
//	dialOpt := grpc.WithInsecure()
//	conn, err := grpc.Dial("localhost:50051", dialOpt)
//	if err != nil {
//		log.Fatalf("failed to dial: %v", err)
//	}
//	desc, err := descriptor.LoadMessageDescriptor("package.MyRequestMessage")
//	if err != nil {
//		log.Fatalf("failed to load message descriptor: %v", err)
//	}
//	msg := dynamic.NewMessage(desc)
//	// 假设您已经将请求消息填充为msg
//	stub := grpcdynamic.NewStub(conn)
//	methodDesc := grpc.MethodDesc{
//		MethodName: "MyMethod",
//	}
//	// 注意，这里的请求和响应必须是proto.Message类型的
//	var resp proto.Message
//	resp, err = stub.InvokeRpc(context.Background(), &methodDesc, msg)
//	if err != nil {
//		log.Fatalf("failed to invoke Rpc: %v", err)
//	}
//	// 假设您收到了响应消息并打印它
//	log.Printf("response received: %v", resp)
//}

func main() {
	p := protoparse.Parser{}
	// 使用path的方式解析得到一些列文件描述对象，这里只有一个文件描述对象
	fileDescs, err := p.ParseFiles("../helloworld/helloworld.proto")
	if err != nil {
		log.Printf("parse proto file failed, err = %s", err.Error())
		return
	}

	// 从文件描述对象中根据消息名拿到消息描述对象
	helloReqDesc := fileDescs[0].FindMessage("helloworld.HelloRequest")
	if helloReqDesc == nil {
		log.Printf("no message matched")
		return
	}

	// 根据消息描述对象生成动态消息
	dMsg := dynamic.NewMessage(helloReqDesc)

	bm, _ := json.Marshal(&helloworld.HelloRequest{Name: "haokun"})
	_ = dMsg.UnmarshalJSON(bm)

	// 从service descriptor中拿到method descriptor
	helloMethodDesc := fileDescs[0].FindService("helloworld.Greeter").FindMethodByName("SayHello")

	// 创建grpc连接
	// TODO 这里如果多个实例怎么请求？
	conn, err := grpc.Dial("127.0.0.1:50051", grpc.WithInsecure())
	if err != nil {
		return
	}
	// 使用grpc连接创建动态的client
	client := grpcdynamic.NewStub(conn)

	// 调用方法
	resp, err := client.InvokeRpc(context.Background(), helloMethodDesc, dMsg)
	if err != nil {
		return
	}
	res, ok := resp.(*dynamic.Message)
	if ok {
		ps := &helloworld.HelloReply{}
		err := res.ConvertTo(ps)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(ps)
	}
	fmt.Printf("%+v\n", resp)
}

func Test_R(t *testing.T) {
	main()
}
