package main

import (
	"context"
	"fmt"
	"go-product/server/api"
	"go-product/server/config"
	"go-product/server/models"
	"go-product/server/pb"
	"net"

	"go-product/server/lib/rabbitmq"
	authPb "go-product/server/lib/stubs/go-auth"

	"github.com/olivere/elastic/v7"
	"github.com/sirupsen/logrus"
	"go.elastic.co/ecslogrus"
	"google.golang.org/grpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const serviceName = "go-product-service"
const defaultPort = 50052
const authServicePort = "50051"

var appConfig *config.Config

func main() {
	appConfig = config.InitConfig(".env")

	esLogger := logrus.New()
	esLogger.SetFormatter(&ecslogrus.Formatter{})
	esLogger.SetLevel(logrus.InfoLevel)

	_, err := elastic.NewClient(elastic.SetURL("http://localhost:9200"), elastic.SetSniff(false))
	if err != nil {
		esLogger.Fatalf("Failed to create elasticsearch client: %v", err)
	}

	dsn := "go_user:go_password@tcp(127.0.0.1:3308)/go_product?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		esLogger.Fatalf("failed to connect database: %v", err)
	}

	db.AutoMigrate(&models.Product{})

	authServiceConn, err := grpc.Dial("localhost:"+authServicePort, grpc.WithInsecure())
	if err != nil {
		esLogger.Fatalf("failed to connect to auth service: %v", err)
	}
	defer authServiceConn.Close()

	esLogger.Info("Connected to auth service")
	authClient := authPb.NewUserServiceClient(authServiceConn)

	esLogger.Info("Getting User")
	ctx := context.Background()
	user, err := authClient.GetUser(ctx, &authPb.GetUserRequest{
		Id: 10,
	})

	esLogger.Info(user)

	rabbitmq.InitRabbitMq()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", defaultPort))
	if err != nil {
		esLogger.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	userService := &api.Server{
		Db:                db,
		Logger:            esLogger,
		AuthServiceClient: authServiceConn,
	}
	pb.RegisterProductServiceServer(grpcServer, userService)

	if err := grpcServer.Serve(lis); err != nil {
		esLogger.Fatalf("failed to serve: %v", err)
	}
}
