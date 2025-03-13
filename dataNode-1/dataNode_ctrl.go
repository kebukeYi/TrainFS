package main

import (
	"context"
	"fmt"
	"github.com/kebukeYi/TrainFS/dataNode-1/service"
	proto "github.com/kebukeYi/TrainFS/profile"
	"google.golang.org/grpc"
	"log"
	"net"
)

type RpcServer struct {
	proto.UnimplementedNameToDataServiceServer
	proto.UnimplementedClientToDataServiceServer
	dataNode *service.DataNode
}

func (s *RpcServer) GetChunk(arg *proto.FileOperationArg, stream proto.ClientToDataService_GetChunkServer) error {
	return s.dataNode.GetChunk(arg, stream)
}

func (s *RpcServer) PutChunk(stream proto.ClientToDataService_PutChunkServer) error {
	return s.dataNode.PutChunk(stream)
}

func (s *RpcServer) GetDataNodeInfo(con context.Context, arg *proto.FileOperationArg) (*proto.FileLocationInfo, error) {
	return s.dataNode.GetDataNodeInfo(arg)
}

func main() {
	dataNode := service.NewDataNode()
	if dataNode == nil {
		return
	}
	listen, err := net.Listen("tcp", dataNode.Config.Host)
	if err != nil {
		log.Fatalln(err)
	}
	rpcServer := &RpcServer{dataNode: dataNode}
	server := grpc.NewServer()
	proto.RegisterClientToDataServiceServer(server, rpcServer)
	fmt.Printf("DataNode-%s is running at %s ...\n", dataNode.Config.DataNodeId, dataNode.Config.Host)
	go dataNode.CheckTask()
	register, err := dataNode.Register()
	if register {
		log.Println("register success!")
	} else {
		log.Println("register fail!")
	}

	err = server.Serve(listen)
	if err != nil {
		log.Fatalln(err)
	}
}
