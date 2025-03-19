package service

import (
	"context"
	"fmt"
	"github.com/kebukeYi/TrainFS/common"
	"github.com/kebukeYi/TrainFS/dataNode-3/config"
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
	Config          *config.DataNode
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

func NewDataNode() *DataNode {
	dataNode := &DataNode{}
	dataNode.Config = config.GetDataNodeConfig()
	common.ClearDir(dataNode.Config.DataDir)
	common.ClearDir(dataNode.Config.MetaDir)
	common.ClearDir(dataNode.Config.TaskDir)
	dataNode.dataStoreManger = OpenStoreManager(dataNode.Config.DataDir)
	dataNode.taskStoreManger = OpenStoreManager(dataNode.Config.TaskDir)
	dataNode.metaStoreManger = OpenStoreManager(dataNode.Config.MetaDir)
	chunkInfos, err := dataNode.metaStoreManger.GetChunkInfos(AllChunkInfosKey)
	if err != nil {
		fmt.Printf("NewDataNode.dataNode-1.metaStoreManger.GetChunkInfos(%s) err: %v \n", AllChunkInfosKey, err)
		return nil
	}
	dataNode.allChunkInfos = chunkInfos

	trashs, err := dataNode.taskStoreManger.GetTrashes(trashKey)
	if err != nil {
		fmt.Printf("NewDataNode.dataNode-1.metaStoreManger.GetTrashes(%s) err: %v \n", trashKey, err)
		return nil
	}
	dataNode.TrashTask = trashs

	replications, err := dataNode.taskStoreManger.GetReplications(replicationKey)
	if err != nil {
		fmt.Printf("NewDataNode.dataNode-1.metaStoreManger.GetReplications(%s) err: %v \n", replicationKey, err)
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
		fmt.Printf("DataNode[%s]-%s stream.SendAndClose() to ip:%s, srcName:%s ;error: %s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId,
			fileDataStream.Address, fileDataStream.SrcName, err)
		return err
	}

	// 另外开一个协程处理写,主协程继续判断是否继续 转发数据;
	// todo dataNode-1 PutChunk()方法需要改进;
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
		fmt.Printf("DataNode[%s]-%s AllChunkInfosStoreManger.PutChunkInfos(%s) ;error: %s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId, AllChunkInfosKey, err)
		delete(dataNode.allChunkInfos, filePathChunkName)
		dataNode.mux.Unlock()
		panic(err)
	}
	dataNode.mux.Unlock()

	fmt.Printf("DataNode[%s]-%s receive fileChunk(%s), from:%s, len:%d; error:%s \n",
		dataNode.Config.Host, dataNode.Config.DataNodeId,
		filePathChunkName, fileDataStream.SrcName+"-"+fileDataStream.Address, len(buf), err)

	// for 循环提交;
	for {
		// 3.提交记录;
		_, err = dataNode.CommitChunk(&proto.CommitChunkArg{
			FileChunkName:   filePathChunkName,
			FilePathName:    filePathName,
			FileSize:        int64(len(buf)),
			Operation:       fileDataStream.Operation, // NormalToReplicate,
			ChunkId:         chunkId,
			SrcAddress:      dataNode.Config.Host,
			DataNodeAddress: []string{dataNode.Config.Host},
		})
		if err != nil {
			time.Sleep(time.Second * 1)
			// todo 需要重试
			fmt.Printf("DataNode[%s]-%s CommitChunk file chunk(%s), len:%d ; error: %s \n",
				dataNode.Config.Host, dataNode.Config.DataNodeId, filePathChunkName, len(buf), err)
			continue
		} else {
			break
		}
	}

	// 转发逻辑
	// dateNode1: [2,3]
	// dataNode2: [3]
	// dateNode3: []
	if len(dataNodeChain) >= 1 {
		nextNodeServer := dataNodeChain[0]
		dataServiceClient, err := dataNode.getGrpcDataServerConn(nextNodeServer)
		putChunkClient, err := dataServiceClient.PutChunk(context.Background())
		if err != nil {
			fmt.Printf("DataNode[%s]-%s dataNode-1.getGrpcDataServerConn(%s) ;error: %s \n",
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
	nameServiceClient, err := dataNode.getGrpcNameNodeServerConn(dataNode.Config.NameNodeHost)
	if err != nil {
		fmt.Printf("DataNode[%s]-%s dataNode-1.getGrpcNameNodeServerConn(%s) ;error: %s \n",
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
		//go dataNode-1.Trash(dataNode-1.TrashTask)
		dataNode.TrashChan <- dataNode.TrashTask
	}
	// time.Sleep(time.Second * 3)
	if len(dataNode.ReplicaTask) > 0 {
		fmt.Printf("DataNode[%s]-%s ReplicaTask:%v; strat...\n", dataNode.Config.Host, dataNode.Config.DataNodeId, dataNode.ReplicaTask)
		//go dataNode-1.Replica(dataNode-1.ReplicaTask)
		dataNode.ReplicaChain <- dataNode.ReplicaTask
	}
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
