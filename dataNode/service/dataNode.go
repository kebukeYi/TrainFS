package service

import (
	"context"
	"fmt"
	proto "github.com/kebukeYi/TrainFS/profile"
	"google.golang.org/grpc"
	"sync"
	"time"
)

const (
	trashKey         string = "trashKey"
	replicationKey   string = "replicationKey"
	AllChunkInfosKey string = "AllChunkInfosKey"
)

type DataNode struct {
	mux             sync.RWMutex
	Config          *Config
	name            string
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

func NewDataNode(configFile *string, port *string, dataNodeId *string) *DataNode {
	dataNode := &DataNode{}
	dataNode.Config = GetDataNodeConfig(configFile, port, dataNodeId)
	//common.ClearDir(dataNode.Config.DataDir)
	//common.ClearDir(dataNode.Config.MetaDir)
	//common.ClearDir(dataNode.Config.TaskDir)
	dataNode.recoveryData()
	chunkInfos, err := dataNode.metaStoreManger.GetChunkInfos(AllChunkInfosKey)
	if err != nil {
		fmt.Printf("NewDataNode.dataNode.metaStoreManger.GetChunkInfos(%s) err: %v \n", AllChunkInfosKey, err)
		return nil
	}
	dataNode.allChunkInfos = chunkInfos

	trashs, err := dataNode.taskStoreManger.GetTrashes(trashKey)
	if err != nil {
		fmt.Printf("NewDataNode.dataNode.metaStoreManger.GetTrashes(%s) err: %v \n", trashKey, err)
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
	dataNode.ReplicaChain = make(chan []*Replication, 100)
	dataNode.TrashChan = make(chan []string, 100)
	return dataNode
}

func (dataNode *DataNode) GetChunk(arg *proto.FileOperationArg, stream proto.ClientToDataService_GetChunkServer) error {
	dataNode.mux.RLock()
	defer dataNode.mux.RUnlock()
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
	fileDataStream, err := stream.Recv()
	if err != nil {
		fmt.Printf("DataNode[%s]-%s stream Recv from ip:%s, srcName:%s, fileChunkName:%s;error: %s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId,
			fileDataStream.Address, fileDataStream.SrcName, fileDataStream.FilePathChunkName, err)
		return err
	}
	// fileName: example.txt
	// FilePathName: /user/app/example.txt
	// FilePathChunkName: /user/app/example.txt_chunk_0  /user/app/example.txt_chunk_1 ...
	// dataNodeChain: [dataNode1, dataNode2, dataNode3]
	filePathName := fileDataStream.FilePathName
	filePathChunkName := fileDataStream.FilePathChunkName
	dataNodeChain := fileDataStream.DataNodeChain
	chunkId := fileDataStream.ChunkId
	dataNode.mux.RLock()
	if _, ok := dataNode.allChunkInfos[filePathChunkName]; ok {
		fmt.Printf("DataNode[%s]-%s Put chunk %s already exists! \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId, filePathChunkName)
		dataNode.mux.RUnlock()
		return nil
	}
	dataNode.mux.RUnlock()

	buf := make([]byte, 0)
	buf = append(buf, fileDataStream.Data...)
	err = stream.SendAndClose(&proto.FileLocationInfo{})
	if err != nil {
		fmt.Printf("DataNode[%s]-%s stream.SendAndClose() to ip:%s, srcName:%s; err: %s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId,
			fileDataStream.Address, fileDataStream.SrcName, err)
		return err
	}

	// 另外开一个协程处理写,主协程继续判断是否继续 转发数据;
	// todo dataNode PutChunk()方法需要改进;
	chunkInfo := &proto.ChunkInfo{
		ChunkId:           chunkId,
		ChunkSize:         int64(len(buf)),
		FilePathName:      filePathName,
		FilePathChunkName: filePathChunkName,
		DataNodeAddress:   &proto.DataNodeChain{DataNodeAddress: []string{dataNode.Config.Host}},
	}

	// 1.文件实体数据先保存;
	if err = dataNode.Write(filePathChunkName, buf); err != nil {
		fmt.Printf("DataNode[%s]-%s Write FilePathChunkName(%s);error: %s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId,
			filePathChunkName, err)
		// 数据写失败了, 需要重新发送;
		return err
	}

	// 2.更新内存元数据;
	dataNode.mux.Lock()
	dataNode.allChunkInfos[filePathChunkName] = chunkInfo
	err = dataNode.metaStoreManger.PutChunkInfos(AllChunkInfosKey, dataNode.allChunkInfos)
	if err != nil {
		fmt.Printf("DataNode[%s]-%s AllChunkInfos,StoreManger.PutChunkInfos(%s) ;error: %s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId, AllChunkInfosKey, err)
		delete(dataNode.allChunkInfos, filePathChunkName)
		dataNode.mux.Unlock()
		panic(err)
	} else {
		fmt.Printf("DataNode[%s]-%s AllChunkInfos,StoreManger.PutChunkInfos(%s) success.\n",
			dataNode.Config.Host, dataNode.Config.DataNodeId, AllChunkInfosKey)
	}
	dataNode.mux.Unlock()

	fmt.Printf("DataNode[%s]-%s receive fileChunk(%s), from:%s, len:%d; error:%s \n",
		dataNode.Config.Host, dataNode.Config.DataNodeId,
		filePathChunkName, fileDataStream.SrcName+"-"+fileDataStream.Address, len(buf), err)

	// 3.提交记录;
	_, err = dataNode.CommitChunk(&proto.CommitChunkArg{
		FileChunkName:   filePathChunkName,
		FilePathName:    filePathName,
		FileSize:        int64(len(buf)),
		Operation:       fileDataStream.Operation, // NormalToReplicate,
		ChunkId:         chunkId,
		SrcAddress:      dataNode.Config.Host,
		DataNodeAddress: []string{dataNode.Config.Host},
	}, 4)

	if err != nil {
		fmt.Printf("DataNode[%s]-%s CommitChunk file chunk(%s), len:%d, err: %s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId, filePathChunkName, len(buf), err)
		return err
	}

	// 转发逻辑
	// dateNode1: [2,3]
	// dataNode2: [3]
	// dateNode3: []
	if len(dataNodeChain) >= 1 {
		nextNodeServer := dataNodeChain[0]
		callBack, dataServiceClient, err := dataNode.getGrpcDataServerConn(nextNodeServer)
		defer callBack()
		putChunkClient, err := dataServiceClient.PutChunk(context.Background(),
			grpc.MaxCallSendMsgSize(dataNode.Config.MaxSendMsgSize*1024*1024))
		if err != nil {
			fmt.Printf("DataNode[%s]-%s dataNode.getGrpcDataServerConn(%s); err: %s \n",
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
		if err != nil {
			fmt.Printf("DataNode[%s]-%s putChunkClient.Send(%s) to ip:%s, srcName:%s ;error: %s \n",
				dataNode.Config.Host, dataNode.Config.DataNodeId, filePathChunkName,
				fileDataStream.Address, fileDataStream.SrcName, err)
			return err
		}
		defer putChunkClient.CloseAndRecv() // 正确调用
		if err != nil {
			fmt.Printf("DataNode[%s]-%s putChunkClient.CloseAndRecv(%s) to ip:%s, srcName:%s ;error: %s \n",
				dataNode.Config.Host, dataNode.Config.DataNodeId, filePathChunkName, nextNodeServer, fileDataStream.SrcName, err)
		} else {
			fmt.Printf("DataNode[%s]-%s forward fileChunk %s to ip:%s, srcName:%s; error: %s \n",
				dataNode.Config.Host, dataNode.Config.DataNodeId, filePathChunkName,
				nextNodeServer, fileDataStream.SrcName, err)
		}
	}
	return nil
}

func (dataNode *DataNode) CommitChunk(arg *proto.CommitChunkArg, retrySize int8) (*proto.CommitChunkReply, error) {
	var reply *proto.CommitChunkReply
	var err error
	var retryS int8
	callBack, nameServiceClient, err := dataNode.getGrpcNameNodeServerConn(dataNode.Config.NameNodeHost)
	defer callBack()
	if err != nil {
		fmt.Printf("DataNode[%s]-%s dataNode.getGrpcNameNodeServerConn(%s) ;error: %s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId,
			dataNode.Config.NameNodeHost, err)
		return nil, err
	}
	for {
		reply, err = nameServiceClient.CommitChunk(context.Background(), arg)
		if err != nil {
			if retryS == retrySize {
				return nil, err
			}
			retryS++
			fmt.Printf("DataNode[%s]-%s CommitChunk fileChunk(%s),type:%s, len:%d; retry:%d; err:%s \n",
				dataNode.Config.Host, dataNode.Config.DataNodeId,
				arg.FileChunkName, arg.Operation, arg.FileSize,
				retryS, err)
			time.Sleep(time.Second * 2)
		} else {
			break
		}
	}
	return reply, nil
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

func (dataNode *DataNode) getGrpcDataServerConn(address string) (func() error, proto.ClientToDataServiceClient, error) {
	clientConn, err := grpc.NewClient(address, grpc.WithInsecure())
	done := func() error {
		return clientConn.Close()
	}
	if err != nil {
		return done, nil, err
	}
	toDataServiceClient := proto.NewClientToDataServiceClient(clientConn)
	return done, toDataServiceClient, nil
}

func (dataNode *DataNode) getGrpcNameNodeServerConn(address string) (func() error, proto.DataToNameServiceClient, error) {
	clientConn, err := grpc.NewClient(address, grpc.WithInsecure())
	done := func() error {
		return clientConn.Close()
	}
	if err != nil {
		return done, nil, err
	}
	toNameNodeServiceClient := proto.NewDataToNameServiceClient(clientConn)
	return done, toNameNodeServiceClient, nil
}

func (dataNode *DataNode) CheckTask() {
	if len(dataNode.TrashTask) > 0 {
		fmt.Printf("DataNode[%s]-%s TrashTask:%v; strat...\n", dataNode.Config.Host, dataNode.Config.DataNodeId, dataNode.TrashTask)
		//go dataNode.Trash(dataNode.TrashTask)
		dataNode.TrashChan <- dataNode.TrashTask
	}
	// time.Sleep(time.Second * 3)
	if len(dataNode.ReplicaTask) > 0 {
		fmt.Printf("DataNode[%s]-%s ReplicaTask:%v; strat...\n", dataNode.Config.Host, dataNode.Config.DataNodeId, dataNode.ReplicaTask)
		//go dataNode.Replica(dataNode.ReplicaTask)
		dataNode.ReplicaChain <- dataNode.ReplicaTask
	}
}

func (dataNode *DataNode) recoveryData() {
	dataNode.dataStoreManger = OpenStoreManager(dataNode.Config.DataDir)
	dataNode.taskStoreManger = OpenStoreManager(dataNode.Config.TaskDir)
	dataNode.metaStoreManger = OpenStoreManager(dataNode.Config.MetaDir)
}

func (dataNode *DataNode) Close() error {
	if dataNode.dataStoreManger != nil {
		err := dataNode.dataStoreManger.Close()
		if err != nil {
			return err
		}
	}
	if dataNode.metaStoreManger != nil {
		err := dataNode.metaStoreManger.Close()
		if err != nil {
			return err
		}
	}
	if dataNode.taskStoreManger != nil {
		err := dataNode.taskStoreManger.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
