package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kebukeYi/TrainFS/dataNode/service"
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
	dataNodeId := flag.String("id", "", "dataNodeID")
	port := flag.String("port", "", "dataNode local port")
	configFile := flag.String("conf", "../conf/dataNode_config.yml", "Path to conf file")
	flag.Parse()
	dataNode := service.NewDataNode(configFile, port, dataNodeId)
	fmt.Println(dataNode.Config)

	defer func(dataNode *service.DataNode) {
		err := dataNode.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(dataNode)
	ip, err := service.GetOutBoundIP()
	if err != nil {
		fmt.Printf("get ip fail! err:%s", err)
		return
	}
	dataNode.Config.Host = ip + ":" + dataNode.Config.Port
	listen, err := net.Listen("tcp", dataNode.Config.Host)
	if err != nil {
		log.Fatalln(err)
	}
	rpcServer := &RpcServer{dataNode: dataNode}
	server := grpc.NewServer(grpc.MaxRecvMsgSize(dataNode.Config.MaxRecvMsgSize*1024*1024),
		grpc.MaxSendMsgSize(dataNode.Config.MaxSendMsgSize*1024*1024))
	proto.RegisterClientToDataServiceServer(server, rpcServer)
	fmt.Printf("DataNode-%s is running at %s ...\n", dataNode.Config.DataNodeId, dataNode.Config.Host)
	go dataNode.CheckTask() // 每次开机时, 进行任务检查,并送入通道中;
	// 1.注册;
	// 2.注册成功, 上报自身存储的全量chunk信息;
	// 3.上报成功, 启动心跳检测;
	// 4.执行心跳传送回来的trash信息, 进行trash处理;
	// 5.执行心跳传送回来的replicate, 进行复制处理;
	ip = "1.94.19.238"
	dataNode.Config.Host = ip + ":" + dataNode.Config.Port
	register, err := dataNode.Register()
	if register {
		log.Println("register success!")
	} else {
		log.Println("register fail!")
		return
	}
	err = server.Serve(listen)
	if err != nil {
		log.Fatalln(err)
	}
}
