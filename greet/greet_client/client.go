package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gogrpc/greet/greetpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Hello I am a client")
	certFile := "ssl/ca.crt" //certificate Authority Trust certificate
	creds, sslErr := credentials.NewClientTLSFromFile(certFile, "")
	if sslErr != nil {
		log.Fatalf("Error while loading CA certificate: %v", sslErr)
		return
	}
	opts := grpc.WithTransportCredentials(creds)

	//cc, err := grpc.Dial("0.0.0.0:50051", grpc.WithInsecure())
	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("Error - clould not connect: %v", err)
	}
	defer cc.Close()
	c := greetpb.NewGreetServiceClient(cc)
	//fmt.Printf("Success! Created client: %f", c)
	doUnary(c)

	//doServerStreaming(c)

	//doClientStreaming(c)

	//doBiDiStreaming(c)

	//doUnaryWithDeadline(c, 5*time.Second) //should complete
	//doUnaryWithDeadline(c, 1*time.Second) //should timeout
}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Printf("Starting to do a Unary RPC...\n")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Sudheer",
			LastName:  "Paladugu",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling Greet RPC %v", err)
	}
	log.Printf("Response from Greet: %v", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Sending to do a Server streaming request RPC..")

	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Sudheer",
			LastName:  "Paladugu",
		},
	}

	resStream, err := c.GreetManyTimes(context.Background(), req)

	if err != nil {
		log.Fatalf("Error when calling GreetManyTimesRequest: %v", err)
	}

	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			//we've reached end of stream
			break
		}
		if err != nil {
			fmt.Printf("error whiel reading stream%+v\n", err)
		}

		log.Printf("response from greet many times: %v", msg.GetResult())
	}
}

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Sending to do a Client Stream request  ...")

	requests := []*greetpb.LongGreetRequest{
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Sudheer",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Sree",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Sadhvik",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Satya",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Subbarao",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Prashant",
			},
		},
	}

	stream, err := c.LongGreet(context.Background())

	if err != nil {
		log.Fatalf("error while calling StreamGreat: %v", err)
	}

	for _, req := range requests {
		fmt.Printf("sending stream request %v\n", req)
		stream.Send(req)
		time.Sleep(1000 * time.Millisecond)
	}

	res, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("error whiel receiving response from LongGreet: %v", err)
	}

	fmt.Printf("Long greet response: %v", res)

}

func doBiDiStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Sending to do a BiDirectional Stream request  ...")

	stream, err := c.GreetEveryone(context.Background())

	//create a stream by invoking the client
	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
		return
	}
	requests := []*greetpb.GreetEveryoneRequest{
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Sudheer",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Sree",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Sadhvik",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Satya",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Subbarao",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName: "Prashant",
			},
		},
	}

	waitc := make(chan struct{})
	//we send a bunch of messages to the client (go routine)
	go func() {
		//to send messages
		for _, req := range requests {
			fmt.Printf("Sending request %v\n", req)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	//we receive a bunch of messages from the client (go routine)
	go func() {
		// to receive messages
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatalf("Error while receiving: %v", err)
				break
			}
			fmt.Printf("Received :%v\n", res.GetResult())
		}
		close(waitc)
	}()
	//block until everything is done
	<-waitc
}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, timeout time.Duration) {
	fmt.Printf("Starting to do a UnaryWithDeadline RPC...\n")
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Sudheer",
			LastName:  "Paladugu",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := c.GreetWithDeadline(ctx, req)
	if err != nil {
		//Checking for type of error. ok == true means error form server, else unexpected (unhandled) error
		statusErr, ok := status.FromError(err)
		//if it is valid error
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout was hit! Deadline was exceeded")
			} else {
				fmt.Printf("Unexpected error: %v", statusErr)
			}
		} else {
			log.Fatalf("Error while calling Greet RPC %v", err)
		}
		return
	}
	log.Printf("Response from Greet: %v", res.Result)
}
