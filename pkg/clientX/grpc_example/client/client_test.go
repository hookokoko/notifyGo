package client

import (
	"context"
	"flag"
	"fmt"
	"log"
	"notifyGo/pkg/clientX/grpc_example/stats"
	"testing"
	"time"

	pb "notifyGo/pkg/clientX/grpc_example/helloworld"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "localhost:50052", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func Test_Main(t *testing.T) {
	flag.Parse()
	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Set up a connection to the server.
	t1 := time.Now()
	conn, err := grpc.DialContext(ctx, *addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(stats.New()),
		grpc.WithBlock(),
	)
	t2 := time.Now()

	fmt.Println(t2.Sub(t1).Milliseconds())

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}
