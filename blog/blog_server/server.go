package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/gogrpc/blog/blogpb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var collection *mongo.Collection

type server struct{}

type blogItem struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	AuthorID string             `bson:"author_id"`
	Content  string             `bson:"content"`
	Title    string             `bson:"title"`
}

func (*server) CreateBlog(ctx context.Context, req *blogpb.CreateBlogRequest) (*blogpb.CreateBlogResponse, error) {
	fmt.Println("creating blog request....")
	blog := req.GetBlog()
	data := blogItem{
		AuthorID: blog.GetAuthorId(),
		Content:  blog.GetContent(),
		Title:    blog.GetTitle(),
	}

	res, err := collection.InsertOne(context.Background(), data)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Internal error :%v", err),
		)
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("connot convert to OID :%v", err),
		)
	}
	return &blogpb.CreateBlogResponse{
		Blog: &blogpb.Blog{
			Id:       oid.Hex(),
			AuthorId: blog.GetAuthorId(),
			Title:    blog.GetTitle(),
			Content:  blog.GetContent(),
		},
	}, nil
}

func (*server) ReadBlog(ctx context.Context, req *blogpb.ReadBlogRequest) (*blogpb.ReadBlogResponse, error) {
	fmt.Println("ReadBlog request received")
	blogid := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogid)
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse id:%v\n", err),
		)
	}

	//create an empty struct
	data := &blogItem{}
	filter := bson.D{primitive.E{Key: "_id", Value: oid}} //as per new version

	res := collection.FindOne(context.Background(), filter)
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID:%v", err),
		)
	}

	return &blogpb.ReadBlogResponse{
		Blog: &blogpb.Blog{
			Id:       data.ID.Hex(),
			AuthorId: data.AuthorID,
			Title:    data.Title,
			Content:  data.Content,
		},
	}, nil

}

func (*server) UpdateBlog(ctx context.Context, req *blogpb.UpdateBlogRequest) (*blogpb.UpdateBlogResponse, error) {
	fmt.Println("UpdateBlog request received")
	//Getting blog from the request
	blog := req.GetBlog()
	//  Getting object ID from hex
	oid, err := primitive.ObjectIDFromHex(blog.GetId())
	if err != nil {
		return nil, status.Errorf(
			codes.InvalidArgument,
			fmt.Sprintf("Cannot parse id:%v\n", err),
		)
	}

	// create an empty struct
	data := &blogItem{}
	// creating a filder to query db as where clause
	filter := bson.D{primitive.E{Key: "_id", Value: oid}} //as per new version
	// db query
	res := collection.FindOne(context.Background(), filter)
	// decoding results into data object and capturing error
	if err := res.Decode(data); err != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot find blog with specified ID:%v", err),
		)
	}

	// updating data struct with request data
	data.AuthorID = blog.GetAuthorId()
	data.Content = blog.GetContent()
	data.Title = blog.GetTitle()
	// updating database with the modified data struct
	_, updErr := collection.ReplaceOne(context.Background(), filter, data)
	if updErr != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot update object in MongoDB: %v", updErr),
		)
	}
	// returning data
	return &blogpb.UpdateBlogResponse{
		Blog: dataToBlogDB(data),
	}, nil
}

func (*server) DeleteBlog(ctx context.Context, req *blogpb.DeleteBlogRequest) (*blogpb.DeleteBlogResponse, error) {
	fmt.Println("DeleteBlog request received")
	blogid := req.GetBlogId()
	oid, err := primitive.ObjectIDFromHex(blogid)
	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Cannot parse id:%v\n", err),
		)
	}

	// filter to delete record
	filter := bson.D{primitive.E{Key: "_id", Value: oid}}

	res, delErr := collection.DeleteOne(context.Background(), filter)
	if delErr != nil {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot delete blog from MongoDB: %v\n", delErr),
		)
	}

	if res.DeletedCount == 0 {
		return nil, status.Errorf(
			codes.NotFound,
			fmt.Sprintf("Cannot delete blog from MongoDB. Internal Error occured..."),
		)
	}

	return &blogpb.DeleteBlogResponse{BlogId: blogid}, nil
}

func (*server) ListBlog(req *blogpb.ListBlogRequest, stream blogpb.BlogService_ListBlogServer) error {
	fmt.Println("List Blog request received")

	// Using cursor. Pass empty bson.D instead of nil for Find().
	// It will return 'document is nil' to the client if nil filter passed to Find()
	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknow Internal Error %v\n", err),
		)
	}
	fmt.Println("cur received")
	//closing cursor
	defer cur.Close(context.Background())

	//streaming
	for cur.Next(context.Background()) {
		data := &blogItem{}
		err := cur.Decode(data)
		if err != nil {
			return status.Errorf(
				codes.Internal,
				fmt.Sprintf("Error while decoding data from MongoDB: %v", err),
			)
		}
		fmt.Println("so far so good")
		//returning stream to the client
		stream.Send(&blogpb.ListBlogResponse{Blog: dataToBlogDB(data)})
		fmt.Println("sent first stream")
	}
	//checking for error in stream
	if aerr := cur.Err(); aerr != nil {
		return status.Errorf(
			codes.Internal,
			fmt.Sprintf("Unknown internal error: %v", aerr),
		)
	}
	// close the stream back to client
	return nil
}

func dataToBlogDB(data *blogItem) *blogpb.Blog {
	return &blogpb.Blog{
		Id:       data.ID.Hex(),
		AuthorId: data.AuthorID,
		Title:    data.Title,
		Content:  data.Content,
	}
}

func main() {
	//To print file name and line number when error occurred
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Blog Server started")

	//connecting Mongo db
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Mongo db client create error :%v", err)
		return
	}
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatalf("Mongo db connection error :%v", err)
		return
	}
	fmt.Println("connecting to mango db")
	//connecting from client
	collection = client.Database("mydb").Collection("blog")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to Listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	s := grpc.NewServer(opts...)
	blogpb.RegisterBlogServiceServer(s, &server{})

	go func() {
		fmt.Println("Starting server...")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to start server: %v", err)
		} else {
			fmt.Println("Server is ready to proces requests...")
		}
	}()
	//fmt.Println("Server is ready to proces requests...")
	//waiting for contrl + C to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	//block until a signal is received
	<-ch

	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Closing listener")
	lis.Close()
	fmt.Println("closing mongo db connection")
	client.Disconnect(context.TODO())
	fmt.Println("Listener closed")

}
