package service

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"sync"
	"time"
	"trainfs/src/dataNode-1/config"
	proto "trainfs/src/profile"
)

const (
	trashKey         string = "trashKey"
	replicationKey   string = "replicationKey"
	AllChunkInfosKey string = "AllChunkInfosKey"
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

	ReplicaTask  []*Replication
	ReplicaChain chan []*Replication

	TrashTask []string
	TrashChan chan []string
}
type Replication struct {
	FilePathName      string
	FilePathChunkName string
	ToAddress         string
}

func NewDataNode() *DataNode {
	dataNode := &DataNode{}
	dataNode.Config = config.GetDataNodeConfig()
	dataNode.dataStoreManger = OpenStoreManager(dataNode.Config.DataDir)
	dataNode.taskStoreManger = OpenStoreManager(dataNode.Config.TaskDir)
	dataNode.metaStoreManger = OpenStoreManager(dataNode.Config.MetaDir)
	chunkInfos, err := dataNode.metaStoreManger.GetChunkInfos(AllChunkInfosKey)
	if err != nil {
		fmt.Printf("NewDataNode.dataNode.metaStoreManger.GetChunkInfos(%s) err: %v \n", AllChunkInfosKey, err)
		return nil
	}
	dataNode.allChunkInfos = chunkInfos

	trashs, err := dataNode.taskStoreManger.GetTrashs(trashKey)
	if err != nil {
		fmt.Printf("NewDataNode.dataNode.metaStoreManger.GetTrashs(%s) err: %v \n", trashKey, err)
		return nil
	}
	dataNode.TrashTask = trashs

	replications, err := dataNode.taskStoreManger.GetReplications(replicationKey)
	if err != nil {
		fmt.Printf("NewDataNode.dataNode.metaStoreManger.GetReplications(%s) err: %v \n", replicationKey, err)
		return nil
	}
	dataNode.ReplicaTask = replications

	dataNode.name = "DataNode-" + dataNode.Config.DataNodeId
	dataNode.ReplicaChain = make(chan []*Replication)
	dataNode.TrashChan = make(chan []string)
	return dataNode
}

func (dataNode *DataNode) GetChunk(arg *proto.FileOperationArg, stream proto.ClientToDataService_GetChunkServer) error {
	dataNode.mux.Lock()
	defer dataNode.mux.Unlock()
	filePathChunkName := arg.GetFileName()
	read, err := dataNode.Read(filePathChunkName)
	if err != nil {
		fmt.Printf("DataNode[%s]-%s read chunk %s fail. err:%s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId, filePathChunkName, err)
		return err
	}
	err = stream.Send(&proto.FileDataStream{Data: read})
	if err != nil {
		fmt.Printf("DataNode[%s]-%s send chunk %s fail. err:%s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId, filePathChunkName, err)
		return err
	}
	fmt.Printf("DataNode[%s]-%s send chunk %s success.\n ",
		dataNode.Config.Host, dataNode.Config.DataNodeId, filePathChunkName)
	return nil
}

func (dataNode *DataNode) PutChunk(stream proto.ClientToDataService_PutChunkServer) error {
	dataNode.mux.Lock()
	defer dataNode.mux.Unlock()
	// fileName: example.txt
	// FilePathName: /user/app/example.txt
	// FilePathChunkName: /user/app/example.txt_chunk_0  /user/app/example.txt_chunk_1
	var filePathChunkName string
	var filePathName string
	var dataNodeChain []string
	fileDataStream, err := stream.Recv()
	if err != nil {
		fmt.Printf("DataNode[%s]-%s stream Recv from ip:%s, srcName:%s, fileChunkName:%s;error: %s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId,
			fileDataStream.Address, fileDataStream.SrcName, fileDataStream.FilePathChunkName, err)
		return err
	}
	filePathChunkName = fileDataStream.FilePathChunkName
	dataNodeChain = fileDataStream.DataNodeChain
	filePathName = fileDataStream.FilePathName
	chunkId := fileDataStream.ChunkId
	if _, ok := dataNode.allChunkInfos[filePathChunkName]; ok {
		fmt.Printf("DataNode[%s]-%s Put chunk %s already exists! \n", dataNode.Config.Host, dataNode.Config.DataNodeId, filePathChunkName)
		return nil
	}
	buf := make([]byte, 0)
	buf = append(buf, fileDataStream.Data...)
	err = stream.SendAndClose(&proto.FileLocationInfo{})
	if err != nil {
		fmt.Printf("DataNode[%s]-%s stream.SendAndClose() to ip:%s, srcName:%s ;error: %s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId,
			fileDataStream.Address, fileDataStream.SrcName, err)
		return err
	}
	go func() {
		chunkInfo := &proto.ChunkInfo{
			ChunkId:           chunkId,
			ChunkSize:         int64(len(buf)),
			FilePathName:      filePathName,
			FilePathChunkName: filePathChunkName,
			DataNodeAddress:   &proto.DataNodeChain{DataNodeAddress: []string{dataNode.Config.Host}},
		}
		dataNode.mux.Lock()
		dataNode.allChunkInfos[filePathChunkName] = chunkInfo
		err = dataNode.metaStoreManger.PutChunkInfos(AllChunkInfosKey, dataNode.allChunkInfos)
		if err != nil {
			fmt.Printf("DataNode[%s]-%s metaStoreManger.PutChunkInfos(%s) ;error: %s \n",
				dataNode.Config.Host, dataNode.Config.DataNodeId,
				AllChunkInfosKey, err)
			return
		}
		dataNode.mux.Unlock()

		if err = dataNode.Write(filePathChunkName, buf); err != nil {
			fmt.Printf("DataNode[%s]-%s Write FilePathChunkName(%s);error: %s \n",
				dataNode.Config.Host, dataNode.Config.DataNodeId,
				filePathChunkName, err)
			return
		}
		fmt.Printf("DataNode[%s]-%s receive fileChunk(%s), from:%s, len:%d; error:%s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId,
			filePathChunkName, fileDataStream.SrcName+"-"+fileDataStream.Address, len(buf), err)

		_, err = dataNode.CommitChunk(&proto.CommitChunkArg{
			FileChunkName:   filePathChunkName,
			FilePathName:    filePathName,
			FileSize:        int64(len(buf)),
			Operation:       fileDataStream.Operation,
			ChunkId:         chunkId,
			SrcAddress:      dataNode.Config.Host,
			DataNodeAddress: []string{dataNode.Config.Host},
		})
		if err != nil {
			// todo 需要重试
			fmt.Printf("DataNode[%s]-%s CommitChunk file chunk(%s), len:%d ; error: %s \n",
				dataNode.Config.Host, dataNode.Config.DataNodeId,
				filePathChunkName, len(buf), err)
		}
	}()

	if len(dataNodeChain) > 0 {
		nextNodeServer := dataNodeChain[0]
		dataServiceClient, err := dataNode.getGrpcDataServerConn(nextNodeServer)
		putChunkClient, err := dataServiceClient.PutChunk(context.Background())
		if err != nil {
			fmt.Printf("DataNode[%s]-%s dataNode.getGrpcDataServerConn(%s) ;error: %s \n",
				dataNode.Config.Host, dataNode.Config.DataNodeId,
				nextNodeServer, err)
			return err
		}
		err = putChunkClient.Send(&proto.FileDataStream{
			FilePathChunkName: filePathChunkName,
			FilePathName:      filePathName,
			ChunkId:           chunkId,
			Data:              buf,
			DataNodeChain:     dataNodeChain[1:],
			Address:           dataNode.Config.Host,
			SrcName:           dataNode.name,
			Operation:         proto.ChunkReplicateStatus_NormalToReplicate,
		})
		defer putChunkClient.CloseSend()
		if err != nil {
			fmt.Printf("DataNode[%s]-%s putChunkClient.Send(%s) to ip:%s, srcName:%s ;error: %s \n",
				dataNode.Config.Host, dataNode.Config.DataNodeId, filePathChunkName,
				fileDataStream.Address, fileDataStream.SrcName, err)
			return err
		}
		fmt.Printf("DataNode[%s]-%s forward fileChunk %s to ip:%s, srcName:%s; error: %s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId, filePathChunkName,
			nextNodeServer, fileDataStream.SrcName, err)
	}
	return nil
}

func (dataNode *DataNode) CommitChunk(arg *proto.CommitChunkArg) (*proto.CommitChunkReply, error) {
	//dataNode.mux.Lock()
	//defer dataNode.mux.Unlock()
	nameServiceClient, err := dataNode.getGrpcNameNodeServerConn(dataNode.Config.NameNodeHost)
	if err != nil {
		fmt.Printf("DataNode[%s]-%s dataNode.getGrpcNameNodeServerConn(%s) ;error: %s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId,
			dataNode.Config.NameNodeHost, err)
		return nil, err
	}
	reply, err := nameServiceClient.CommitChunk(context.Background(), arg)
	fmt.Printf("DataNode[%s]-%s CommitChunk fileChunk(%s),type:%s, len:%d ; error: %s \n",
		dataNode.Config.Host, dataNode.Config.DataNodeId,
		arg.Operation,
		arg.FileChunkName, arg.FileSize, err)
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
	return dataNode.dataStoreManger.Get(fileName)
}

func (dataNode *DataNode) Write(fileName string, data []byte) error {
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
	withTimeout, _ := context.WithTimeout(context.Background(), time.Duration(5000*10)*time.Millisecond)
	clientConn, err := grpc.DialContext(withTimeout, address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	toNameNodeServiceClient := proto.NewDataToNameServiceClient(clientConn)
	return toNameNodeServiceClient, nil
}

func (dataNode *DataNode) CheckTask() {
	if len(dataNode.TrashTask) > 0 {
		fmt.Printf("DataNode[%s]-%s TrashTask:%v; strat...\n", dataNode.Config.Host, dataNode.Config.DataNodeId, dataNode.TrashTask)
		go dataNode.Trash(dataNode.TrashTask)
	}
	time.Sleep(time.Second * 3)
	if len(dataNode.ReplicaTask) > 0 {
		fmt.Printf("DataNode[%s]-%s ReplicaTask:%v; strat...\n", dataNode.Config.Host, dataNode.Config.DataNodeId, dataNode.ReplicaTask)
		go dataNode.Replica(dataNode.ReplicaTask)
	}
}
