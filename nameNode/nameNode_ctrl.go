package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kebukeYi/TrainFS/common"
	"github.com/kebukeYi/TrainFS/nameNode/service"
	proto "github.com/kebukeYi/TrainFS/profile"
	"google.golang.org/grpc"
	"log"
	"net"
)

type RpcServer struct {
	proto.UnimplementedClientToNameServiceServer
	proto.UnimplementedDataToNameServiceServer
	nameNode *service.NameNode
}

func (s RpcServer) PutFile(con context.Context, arg *proto.FileOperationArg) (*proto.DataNodeChain, error) {
	return s.nameNode.PutFile(arg)
}
func (s RpcServer) ConfirmFile(con context.Context, arg *proto.ConfirmFileArg) (*proto.ConfirmFileReply, error) {
	return s.nameNode.ConfirmFile(arg)
}
func (s RpcServer) GetFile(con context.Context, arg *proto.FileOperationArg) (*proto.FileLocationInfo, error) {
	return s.nameNode.GetFile(arg)
}
func (s RpcServer) GetFileLocation(con context.Context, arg *proto.FileOperationArg) (*proto.FileLocationInfo, error) {
	return s.nameNode.GetFileLocation(arg.GetFileName(), arg.GetReplicaNum())
}
func (s RpcServer) GetFileStoreChain(con context.Context, arg *proto.FileOperationArg) (*proto.DataNodeChain, error) {
	return s.nameNode.GetFileStoreChain(arg)
}
func (s RpcServer) DeleteFile(con context.Context, arg *proto.FileOperationArg) (*proto.DeleteFileReply, error) {
	return s.nameNode.DeleteFile(arg)
}
func (s RpcServer) ListDir(con context.Context, arg *proto.FileOperationArg) (*proto.DirMetaList, error) {
	return s.nameNode.ListDir(arg)
}
func (s RpcServer) ReName(con context.Context, arg *proto.FileOperationArg) (*proto.ReNameReply, error) {
	return s.nameNode.ReName(arg)
}
func (s RpcServer) Mkdir(con context.Context, arg *proto.FileOperationArg) (*proto.MkdirReply, error) {
	return s.nameNode.Mkdir(arg)
}
func (s RpcServer) RegisterDataNode(con context.Context, arg *proto.DataNodeRegisterArg) (*proto.DataNodeRegisterReply, error) {
	return s.nameNode.RegisterDataNode(arg)
}
func (s RpcServer) HeartBeat(con context.Context, arg *proto.HeartBeatArg) (*proto.HeartBeatReply, error) {
	return s.nameNode.HeartBeat(arg)
}
func (s RpcServer) ChunkReport(con context.Context, arg *proto.FileLocationInfo) (*proto.ChunkReportReply, error) {
	return s.nameNode.ChunkReport(arg)
}
func (s RpcServer) CommitChunk(con context.Context, arg *proto.CommitChunkArg) (*proto.CommitChunkReply, error) {
	return s.nameNode.CommitChunk(arg)
}
func (s RpcServer) LiveDetection(con context.Context, arg *proto.LiveDetectionArg) (*proto.LiveDetectionReply, error) {
	return s.nameNode.LiveDetection(arg)
}

func main() {
	configFile := flag.String("conf", "../conf/nameNode_config.yml", "Path to conf file")
	flag.Parse()
	newNameNode := service.NewNameNode(configFile)
	defer func(newNameNode *service.NameNode) {
		err := newNameNode.Close()
		if err != nil {
			log.Fatalf("failed to close: %v", err)
		}
	}(newNameNode)
	server1 := &RpcServer{nameNode: newNameNode}
	server2 := &RpcServer{nameNode: newNameNode}
	ip, err := common.GetOutBoundIP()
	if err != nil {
		fmt.Printf("get ip fail! err:%s", err)
		return
	}
	newNameNode.Config.Config.Host = ip + ":" + newNameNode.Config.Config.Port
	listen, err := net.Listen("tcp", newNameNode.Config.Config.Host)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer(grpc.MaxSendMsgSize(newNameNode.Config.Config.MaxSendMsgSize*1024*1024),
		grpc.MaxRecvMsgSize(newNameNode.Config.Config.MaxRecvMsgSize*1024*1024))
	proto.RegisterClientToNameServiceServer(server, server1)
	proto.RegisterDataToNameServiceServer(server, server2)
	fmt.Printf("NameNode server %s start... \n", newNameNode.Config.Config.Host)
	go newNameNode.CheckHeartBeat() // 检测dataNode心跳;
	err = server.Serve(listen)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
