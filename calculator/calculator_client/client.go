package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gogrpc/calculator/calculatorpb"

	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Calculator client")
	cc, err := grpc.Dial("0.0.0.0:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Error - clould not connect: %v", err)
	}
	defer cc.Close()
	c := calculatorpb.NewCalculatorServiceClient(cc)
	//fmt.Printf("Success! Created client: %f", c)
	//doUnary(c)

	//doSteaming(c)

	//doClientSteaming(c)
	doBiDiSteaming(c)
}

func doUnary(c calculatorpb.CalculatorServiceClient) {
	fmt.Printf("Starting calculator client do a Unary RPC...")
	req := &calculatorpb.SumRequest{
		FirstNumber:  5,
		SecondNumber: 10,
	}
	res, err := c.Sum(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling Calculator.Sum RPC %v", err)
	}
	log.Printf("Response from Calculator.Sum: %v", res.SumResult)
}

func doSteaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Printf("Starting PrimeDecompositionServer Streaming RPC...")
	req := &calculatorpb.PrimeNumberDecompositioinRequest{
		Number: 1231234523,
	}

	stream, err := c.PrimeNumberDecompositioin(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling PrimeNumberDecomposition RPC %v", err)
	}
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Something happend :%v", err)
		}
		log.Printf("Response from PrimeNumberDivision: %v\n", res.GetPrimeFactor())
	}

}

func doClientSteaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Printf("Starting ComputeAgerage Streaming RPC...")

	stream, err := c.ComputeAverage(context.Background())
	if err != nil {
		log.Fatalf("Error while calling ComputeAverage RPC %v", err)
	}

	numbers := []int64{3, 5, 9, 20, 120, 2002}

	for _, number := range numbers {
		fmt.Printf("Sending number %v\n", number)
		stream.Send(&calculatorpb.ComputeAverageRequest{
			Number: number,
		})
	}

	fmt.Printf("Closing stream...")
	res, err := stream.CloseAndRecv()
	fmt.Printf("Closed and Received response...")
	if err != nil {
		log.Fatalf("Error while receiving response: %v", err)
	}

	fmt.Printf("The Average is %v\n", res.GetAverage())

}

func doBiDiSteaming(c calculatorpb.CalculatorServiceClient) {
	fmt.Printf("Starting FindMaximum BiDirectioinal Streaming RPC...")

	stream, err := c.FindMaximum(context.Background())
	if err != nil {
		log.Fatalf("Error while opning stream to call FindMaximum RPC %v", err)
	}

	waitc := make(chan struct{})
	//send go routine
	go func() {
		numbers := []int64{3, 5, 9, 2, 120, 3, 11, 5, 2002}
		for _, number := range numbers {
			fmt.Printf("Sending number : %v\n", number)
			stream.Send(&calculatorpb.FindMaximumRequest{
				Number: number,
			})
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()

	//receive go routine
	go func() {
		for {
			res, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while reading response: %v", err)
				break
			}
			maximum := res.GetMaximum()
			fmt.Printf("_____New Maximum is : %v\n", maximum)
		}
		close(waitc)
	}()
	<-waitc
}
