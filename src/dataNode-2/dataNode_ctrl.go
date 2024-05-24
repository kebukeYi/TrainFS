package main

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"net"
	"trainfs/src/dataNode-2/service"
	proto "trainfs/src/profile"
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
	listen, err := net.Listen("tcp", dataNode.Config.Host)
	if err != nil {
		log.Fatalln(err)
	}
	rpcServer := &RpcServer{dataNode: dataNode}
	server := grpc.NewServer()
	register, err := dataNode.Register()
	if register {
		go dataNode.ChunkReportTask()
		go dataNode.HeartBeatTask()
		go dataNode.DoTrashTask()
		go dataNode.DoReplicaTask()
	}
	proto.RegisterClientToDataServiceServer(server, rpcServer)
	err = server.Serve(listen)
	if err != nil {
		log.Fatalln(err)
	}
}
