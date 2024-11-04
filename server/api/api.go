package api

import (
	"go-product/server/pb"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Server struct {
	pb.ProductServiceServer
	Db      *gorm.DB
	Logger  *logrus.Logger
	AuthServiceClient *grpc.ClientConn
}
