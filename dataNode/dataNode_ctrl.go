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
	// A: src code run kind : package
	// program arguments: -id=1 -host=127.0.0.1:9001 -conf=dataNode/conf/dataNode_config.yml
	// B: go run dataNode_ctrl.go -id=1 -host=127.0.0.1:9001 -conf=./conf/dataNode_config.yml
	// C: go build -o ./build/dataNode.exe
	//    cd build dataNode.exe -id=1 -host=127.0.0.1:9001
	dataNodeId := flag.String("id", "", "dataNodeID")
	hostIP := flag.String("host", "", "dataNode local addr ip:port")
	configFile := flag.String("conf", "../conf/dataNode_config.yml", "Path to conf file")
	flag.Parse()
	dataNode := service.NewDataNode(configFile, hostIP, dataNodeId)
	fmt.Println(dataNode.Config)

	defer func(dataNode *service.DataNode) {
		err := dataNode.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}(dataNode)

	listen, err := net.Listen("tcp", dataNode.Config.Host)
	if err != nil {
		log.Fatalln(err)
	}
	rpcServer := &RpcServer{dataNode: dataNode}
	server := grpc.NewServer()
	proto.RegisterClientToDataServiceServer(server, rpcServer)
	fmt.Printf("DataNode-%s is running at %s ...\n", dataNode.Config.DataNodeId, dataNode.Config.Host)
	go dataNode.CheckTask() // 每次开机时, 进行任务检查,并送入通道中;
	// 1.注册;
	// 2.注册成功, 上报自身存储的全量chunk信息;
	// 3.上报成功, 启动心跳检测;
	// 4.执行心跳传送回来的trash信息, 进行trash处理;
	// 5.执行心跳传送回来的replicate, 进行复制处理;
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
