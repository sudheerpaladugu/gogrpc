package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/gogrpc/blog/blogpb"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("Hello from blog client")

	opts := grpc.WithInsecure()

	//cc, err := grpc.Dial("0.0.0.0:50051", grpc.WithInsecure())
	cc, err := grpc.Dial("localhost:50051", opts)
	if err != nil {
		log.Fatalf("Error - clould not connect to blog server: %v\n", err)
	}
	defer cc.Close()
	c := blogpb.NewBlogServiceClient(cc)

	//create Blog
	blog := &blogpb.Blog{
		AuthorId: "Sudheer",
		Title:    "My first blog",
		Content:  "Share feedback",
	}
	res, err := c.CreateBlog(context.Background(), &blogpb.CreateBlogRequest{Blog: blog})

	if err != nil {
		log.Fatalf("Unexpected error : %v\n", err)
	}
	fmt.Printf("Blog has been created: %v\n", res)
	blogid := res.GetBlog().GetId()

	fmt.Println("Reading the blog...")
	resp, err2 := c.ReadBlog(context.Background(), &blogpb.ReadBlogRequest{BlogId: "ab_5e6581894086248fccee982b"})

	if err2 != nil {
		fmt.Printf("Error happend while reading: %v\n", err2)
	}
	fmt.Printf("Blog response 1 :%v\n", resp)
	readBlogReq := &blogpb.ReadBlogRequest{BlogId: blogid}

	res2, err3 := c.ReadBlog(context.Background(), readBlogReq)

	if err3 != nil {
		fmt.Printf("Error happend while reading: %v\n", err2)
	}
	fmt.Printf("Blog response 2 :%v\n", res2)

	//update blog

	//createing a Blog
	newbBlog := &blogpb.Blog{
		Id:       blogid,
		AuthorId: "Changed Author",
		Title:    "My first blog event",
		Content:  "Share feedback",
	}

	updateRes, updErr := c.UpdateBlog(context.Background(), &blogpb.UpdateBlogRequest{Blog: newbBlog})
	if updErr != nil {
		fmt.Printf("Error happend while updating: %v \n", updErr)
	}
	fmt.Printf("Blog was read: %v\n", updateRes)

	//Delete blog
	deleteRes, delErr := c.DeleteBlog(context.Background(), &blogpb.DeleteBlogRequest{BlogId: blogid})
	if delErr != nil {
		fmt.Printf("Error happend while deleteting: %v \n", delErr)
	}
	fmt.Printf("Blog was deleted: %v\n", deleteRes)

	//List blogs
	stream, strErr := c.ListBlog(context.Background(), &blogpb.ListBlogRequest{})
	if strErr != nil {
		log.Fatalf("Fatal Error while calling ListBlog RPC : %v \n", strErr)
	}
	for {
		res, aerr := stream.Recv()
		if aerr == io.EOF {
			break
		}
		if aerr != nil {
			log.Fatalf("Unexpected Error: %v", aerr)
		}
		fmt.Println("Stream result ", res.GetBlog())
	}

}
