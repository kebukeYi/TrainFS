package service

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"sync"
	"trainfs/src/dataNode-2/config"
	proto "trainfs/src/profile"
)

const (
	TRASHTASK_KEY string = "TRASHTASK_KEY"
	REPLICAT_KEY  string = "REPLICAT_KEY"
	CHUNKINFOSKEY string = "CHUNKINFOSKEY"
)

type DataNode struct {
	mux             sync.Mutex
	Config          *config.DataNode
	name            string
	FileBuffer      *FileBuffer
	dataStoreManger *StoreManger

	taskStoreManger *StoreManger
	metaStoreManger *StoreManger
	allChunkInfos   map[string]*proto.ChunkInfo
	ReplicaTask     []*Replication
	ReplicaChain    chan []*Replication
	TrashTask       []string
	TrashChan       chan []string
}
type Replication struct {
	filePathName      string
	filePathChunkName string
	toAddress         string
}

func NewDataNode() *DataNode {
	dataNode := &DataNode{}
	dataNode.Config = config.GetDataNodeConfig()
	// dataNode.FileBuffer = NewFileBuffer()
	dataNode.dataStoreManger = OpenStoreManager(dataNode.Config.DataDir)
	dataNode.taskStoreManger = OpenStoreManager(dataNode.Config.TaskDir)
	dataNode.metaStoreManger = OpenStoreManager(dataNode.Config.MetaDir)
	dataNode.allChunkInfos = make(map[string]*proto.ChunkInfo)
	dataNode.name = "DataNode-" + dataNode.Config.DataNodeId
	dataNode.ReplicaChain = make(chan []*Replication)
	dataNode.TrashChan = make(chan []string)
	fmt.Println("DataNode-" + dataNode.Config.DataNodeId + " is running...")
	return dataNode
}

func (dataNode *DataNode) GetChunk(arg *proto.FileOperationArg, stream proto.ClientToDataService_GetChunkServer) error {
	dataNode.mux.Lock()
	defer dataNode.mux.Unlock()
	filePathChunkName := arg.GetFileName()
	read, err := dataNode.Read(filePathChunkName)
	if err != nil {
		return err
	}
	err = stream.Send(&proto.FileDataStream{Data: read})
	if err != nil {
		return err
	}
	fmt.Println("DataNode-" + dataNode.Config.DataNodeId + " send file " + filePathChunkName)
	return nil
}

func (dataNode *DataNode) PutChunk(stream proto.ClientToDataService_PutChunkServer) error {
	dataNode.mux.Lock()
	defer dataNode.mux.Unlock()
	buf := make([]byte, 0)
	// fileName: example.txt
	// filePathName: /user/app/example.txt
	// filePathChunkName: /user/app/example.txt_chunk_id
	var filePathChunkName string
	var filePathName string
	var dataNodeChain []string
	fileDataStream, err := stream.Recv()
	chunkId := fileDataStream.ChunkId
	filePathChunkName = fileDataStream.FilePathChunkName
	dataNodeChain = fileDataStream.DataNodeChain
	filePathName = fileDataStream.FilePathName
	if err != nil {
		log.Printf("DataNode stream.Recv() ip:%s, srcName:%s,  error: %s", fileDataStream.Address, fileDataStream.SrcName, err)
		return err
	}
	buf = append(buf, fileDataStream.Data...)
	err = stream.SendAndClose(&proto.FileLocationInfo{})
	if err != nil {
		fmt.Printf("DataNode stream.SendAndClose() ip:%s, srcName:%s,  error: %s", fileDataStream.Address, fileDataStream.SrcName, err)
	}
	go func() {
		chunkInfo := &proto.ChunkInfo{
			FilePathName:      filePathName,
			FilePathChunkName: filePathChunkName,
			DataNodeAddress:   &proto.DataNodeChain{DataNodeAddress: []string{dataNode.Config.Host}},
			ChunkId:           chunkId,
		}
		dataNode.mux.Lock()
		dataNode.allChunkInfos[filePathChunkName] = chunkInfo
		err = dataNode.metaStoreManger.PutChunkInfos(CHUNKINFOSKEY, dataNode.allChunkInfos)
		if err != nil {
			return
		}
		dataNode.mux.Unlock()
		if err = dataNode.Write(filePathChunkName, buf); err != nil {
			log.Println(err)
			log.Printf("DataNode Write filePathChunkName:%s, error: %s", filePathChunkName, err)
			return
		}
		fmt.Printf("DataNode-%s receive file chunk:%s; len: %d ;\n", dataNode.Config.DataNodeId, filePathChunkName, len(buf))
		_, err = dataNode.CommitChunk(&proto.CommitChunkArg{
			FileChunkName: filePathChunkName,
			FilePathName:  filePathName,
			FileSize:      int64(len(buf)),
			Operation:     fileDataStream.Operation,
			ChunkId:       chunkId,
		})
		if err != nil {
			// todo 需要重试
			log.Println(err)
			log.Printf("DataNode Write fileChunkName:%s, error: %s", filePathChunkName, err)
		}
	}()

	if len(dataNodeChain) > 0 {
		nextNodeServer := dataNodeChain[0]
		dataServiceClient, err := dataNode.getGrpcDataServerConn(nextNodeServer)
		putChunkClient, err := dataServiceClient.PutChunk(context.Background())
		if err != nil {
			return err
		}
		err = putChunkClient.Send(&proto.FileDataStream{
			FilePathChunkName: filePathChunkName,
			FilePathName:      filePathName,
			Data:              buf,
			DataNodeChain:     dataNodeChain[1:],
			Address:           dataNode.Config.Host,
			SrcName:           dataNode.name,
			Operation:         proto.ChunkReplicateStatus_NormalToReplicate,
			ChunkId:           chunkId,
		})
		fmt.Println("DataNode-" + dataNode.Config.DataNodeId + " forward file chunk  " + filePathChunkName + " to " + nextNodeServer)
		if err != nil {
			// todo 需要重试发送
			log.Printf("DataNode send to dataNode-1:%s, error: %s", nextNodeServer, err)
			return err
		}
	}
	return nil
}

func (dataNode *DataNode) CommitChunk(arg *proto.CommitChunkArg) (*proto.CommitChunkReply, error) {
	dataNode.mux.Lock()
	defer dataNode.mux.Unlock()
	nameServiceClient, err := dataNode.getGrpcNameNodeServerConn(dataNode.Config.NameNodeHost)
	if err != nil {
		return nil, nil
	}
	reply, err := nameServiceClient.CommitChunk(context.Background(), &proto.CommitChunkArg{
		FileChunkName:   arg.GetFileChunkName(),
		FilePathName:    arg.FilePathName,
		DataNodeAddress: []string{dataNode.Config.Host},
		Operation:       arg.Operation,
	})
	fmt.Println("DataNode-" + dataNode.Config.DataNodeId + " commit file chunk  " + arg.GetFileChunkName() + " Operation " + arg.Operation.String())
	return reply, err
}

func (dataNode *DataNode) GetDataNodeInfo(arg *proto.FileOperationArg) (*proto.FileLocationInfo, error) {
	dataNode.mux.Lock()
	defer dataNode.mux.Unlock()
	chunkInfos := make([]*proto.ChunkInfo, 0)
	for _, chunkInfo := range dataNode.allChunkInfos {
		chunkInfos = append(chunkInfos, chunkInfo)
	}
	chunk := &proto.FileLocationInfo{Chunks: chunkInfos, DataNodeAddress: dataNode.Config.Host}
	return chunk, nil
}

func (dataNode *DataNode) Read(fileName string) ([]byte, error) {
	dataNode.mux.Lock()
	defer dataNode.mux.Unlock()
	return dataNode.dataStoreManger.Get(fileName)
}

func (dataNode *DataNode) Write(fileName string, data []byte) error {
	dataNode.mux.Lock()
	defer dataNode.mux.Unlock()
	return dataNode.dataStoreManger.Put(fileName, data)
}

func (dataNode *DataNode) getGrpcDataServerConn(address string) (proto.ClientToDataServiceClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	toDataServiceClient := proto.NewClientToDataServiceClient(clientConn)
	return toDataServiceClient, nil
}

func (dataNode *DataNode) getGrpcNameNodeServerConn(address string) (proto.DataToNameServiceClient, error) {
	clientConn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	toNameNodeServiceClient := proto.NewDataToNameServiceClient(clientConn)
	return toNameNodeServiceClient, nil
}
