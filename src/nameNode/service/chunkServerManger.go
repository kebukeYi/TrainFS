package service

import (
	"fmt"
	"time"
	"trainfs/src/common"
	proto "trainfs/src/profile"
)

const (
	datanodeDown = datanodeStatus("datanodeDown")
	datanodeUp   = datanodeStatus("datanodeUp")
)

const (
	DATALISTKEY string = "DATALISTKEY"
)

const (
	LostToReplicate   string = "LostToReplicate"
	NormalToReplicate string = "NormalToReplicate"
)

type DataNodeInfo struct {
	Address               string
	FreeSpace             uint64
	HeartBeatTimeStamp    int64
	Status                datanodeStatus
	replicationChunkNames []*Replication
	trashChunkNames       []string
}
type Replication struct {
	filePathName      string
	filePathChunkName string
	toAddress         string
}

// ByFreeSpace 实现sort.Interface接口
type ByFreeSpace []*DataNodeInfo

func (a ByFreeSpace) Len() int           { return len(a) }
func (a ByFreeSpace) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByFreeSpace) Less(i, j int) bool { return a[i].FreeSpace >= a[j].FreeSpace }

func (nn *NameNode) RegisterDataNode(arg *proto.DataNodeRegisterArg) (*proto.DataNodeRegisterReply, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	if _, ok := nn.dataNodeInfos[arg.DataNodeAddress]; !ok {
		dataNodeInfo := &DataNodeInfo{
			Address:               arg.DataNodeAddress,
			FreeSpace:             arg.FreeSpace,
			HeartBeatTimeStamp:    time.Now().UnixMilli(),
			Status:                datanodeUp,
			replicationChunkNames: make([]*Replication, 0),
			trashChunkNames:       make([]string, 0),
		}
		nn.dataNodeInfos[arg.DataNodeAddress] = dataNodeInfo
		nn.DB.PutDataNodeMeta(DATALISTKEY, nn.dataNodeInfos)
		fmt.Println("NameNode rev " + arg.DataNodeAddress + " register success...")
	}
	return &proto.DataNodeRegisterReply{Success: true}, nil
}

func (nn *NameNode) HeartBeat(arg *proto.HeartBeatArg) (*proto.HeartBeatReply, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	heartBeatReply := &proto.HeartBeatReply{}
	if dataNodeInfo, ok := nn.dataNodeInfos[arg.DataNodeAddress]; ok {
		dataNodeInfo.HeartBeatTimeStamp = time.Now().UnixMilli()
		if dataNodeInfo.replicationChunkNames != nil && len(dataNodeInfo.replicationChunkNames) > 0 {
			for i, replication := range dataNodeInfo.replicationChunkNames {
				heartBeatReply.FilePathNames[i] = replication.filePathName
				heartBeatReply.FilePathChunkNames[i] = replication.filePathChunkName
				heartBeatReply.NewChunkSevers[i] = replication.toAddress
			}
			dataNodeInfo.replicationChunkNames = make([]*Replication, 0)
		}
		if dataNodeInfo.trashChunkNames != nil && len(dataNodeInfo.trashChunkNames) > 0 {
			for i, trash := range dataNodeInfo.trashChunkNames {
				heartBeatReply.TrashFilePathChunkNames[i] = trash
			}
			dataNodeInfo.trashChunkNames = make([]string, 0)
		}
		// fmt.Printf("NameNode rev HeartBeat from %s;\n", arg.DataNodeAddress)
		return heartBeatReply, nil
	}
	_, err := nn.RegisterDataNode(&proto.DataNodeRegisterArg{DataNodeAddress: arg.DataNodeAddress, FreeSpace: arg.FreeSpace})
	return nil, err
}

func (nn *NameNode) ChunkReport(arg *proto.FileLocationInfo) (*proto.ChunkReportReply, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	dataNodeAddress := arg.GetDataNodeAddress()
	chunks := arg.GetChunks()
	for _, chunk := range chunks {
		// todo Name 核心，构建文件映射
		nn.chunkLocation[chunk.FilePathChunkName] = append(nn.chunkLocation[chunk.FilePathChunkName],
			&ReplicaMeta{DataNodeAddress: dataNodeAddress})
		nn.dataNodeChunks[dataNodeAddress] = append(nn.dataNodeChunks[dataNodeAddress], &ChunkMeta{
			ChunkName: chunk.FilePathChunkName,
			ChunkId:   chunk.ChunkId,
			TimeStamp: time.Now().UnixMilli(),
		})
	}
	fmt.Printf("NameNode rev ChunkReport: %v, from %s \n", chunks, dataNodeAddress)
	return &proto.ChunkReportReply{Success: true}, nil
}

func (nn *NameNode) CommitChunk(arg *proto.CommitChunkArg) (*proto.CommitChunkReply, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	chunkId := arg.GetChunkId()
	dataNodeAddress := arg.GetDataNodeAddress()[0]
	fileChunkName := arg.GetFileChunkName()
	nn.chunkLocation[fileChunkName] = append(nn.chunkLocation[fileChunkName], &ReplicaMeta{
		ChunkId:         chunkId,
		ChunkName:       fileChunkName,
		DataNodeAddress: dataNodeAddress,
	})
	nn.dataNodeChunks[dataNodeAddress] = append(nn.dataNodeChunks[dataNodeAddress], &ChunkMeta{
		ChunkId:   chunkId,
		ChunkName: fileChunkName,
		TimeStamp: time.Now().UnixMilli(),
	})
	nn.dataNodeInfos[dataNodeAddress].FreeSpace -= uint64(arg.GetFileSize())

	switch arg.Operation {
	case proto.ChunkReplicateStatus_LostToReplicate:
		fmt.Printf("NameNode rev commitChunk LostToReplicate: chunkId:%d; dataNodeAddress:%s; fileChunkName:%s;\n", chunkId, dataNodeAddress, fileChunkName)
		return &proto.CommitChunkReply{Success: true}, nil
	case proto.ChunkReplicateStatus_NormalToReplicate:
		err := nn.HandleNormalToReplicate(arg)
		return &proto.CommitChunkReply{Success: true}, err
	default:
		return &proto.CommitChunkReply{Success: false}, common.ErrCommitChunkType
	}
}

func (nn *NameNode) HandleNormalToReplicate(arg *proto.CommitChunkArg) error {
	chunkId := arg.GetChunkId()
	filePathName := arg.GetFilePathName()
	fileChunkName := arg.GetFileChunkName()
	meta, err := nn.DB.GetFileMeta(filePathName)
	fmt.Printf("NameNode rev commitChunk NormalToReplicate: chunkId:%d; fileChunkName:%s;\n", chunkId, fileChunkName)
	if err != nil {
		return err
	} else {
		chunks := make([]*ChunkMeta, 0)
		chunks = append(chunks, &ChunkMeta{
			ChunkName: fileChunkName,
			ChunkId:   chunkId,
			TimeStamp: time.Now().UnixMilli(),
		})
		if meta != nil {
			meta.Chunks = append(meta.Chunks, chunks...)
		} else {
			meta = &FileMeta{
				FileName:    filePathName,
				KeyFileName: fileChunkName,
				FileSize:    arg.GetFileSize(),
				IsDir:       false,
				Chunks:      chunks,
			}
		}
		err = nn.DB.PutFileMeta(filePathName, meta)
		fmt.Printf("NameNode HandleNormalToReplicate nn.DB.PutFileMeta(%s, %v); \n", filePathName, meta)
		fmt.Printf("NameNode HandleNormalToReplicate.chunks: %v; \n", meta.Chunks)
		if err != nil {
			return err
		}
	}
	return nil
}

func (nn *NameNode) CheckHeartBeat() {
	for {
		nn.mux.Lock()
		for address, dataNodeInfo := range nn.dataNodeInfos {
			if time.Now().UnixMilli()-dataNodeInfo.HeartBeatTimeStamp > int64(nn.Config.NameNode.DataNodeHeartBeatTimeout) && dataNodeInfo.Status == datanodeUp {
				dataNodeInfo.Status = datanodeDown
				fmt.Printf("NameNode CheckHeartBeat() dataNode %s down! \n" + address)
				fmt.Printf("NameNode dataNodeInfos[%v];\n", nn.dataNodeInfos[address])
				// todo 要不要持久化 dataNode-1 的状态,以及NameNode中的内存队列
				// go nn.ReplicationBalance(address)
			}
		}
		nn.mux.Unlock()
		time.Sleep(time.Duration(nn.Config.NameNode.DataNodeHeartBeatInterval) * time.Millisecond)
	}
}

func (nn *NameNode) ReplicationBalance(downDataNodeAddress string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("NameNode ReplicationBalance panic: ", err)
		}
	}()
	nn.mux.Lock()
	defer nn.mux.Unlock()
	needCopyChunkMetas := nn.dataNodeChunks[downDataNodeAddress]
	srcChunkNodes := make([]string, 0)
	desChunkNodes := make([]string, 0)
	// filePathName := make([]string, 0)
	filePathChunkName := make([]string, 0)

	for i, chunkMeta := range needCopyChunkMetas {
		replicaMetas := nn.chunkLocation[chunkMeta.ChunkName]
		for _, replicaMeta := range replicaMetas {
			if replicaMeta.DataNodeAddress != downDataNodeAddress &&
				nn.dataNodeInfos[replicaMeta.DataNodeAddress].Status == datanodeUp {
				srcChunkNodes[i] = replicaMeta.DataNodeAddress
			}
		}
		desChunkNodes[i] = nn.findNoneChunkOfDataNode(chunkMeta.ChunkName)
		filePathChunkName[i] = chunkMeta.ChunkName
		//filePathName[i]= nn.getFilePathName(chunkMeta.ChunkName)
		if srcChunkNodes[i] != "" && desChunkNodes[i] != "" {
			nn.dataNodeInfos[srcChunkNodes[i]].replicationChunkNames =
				append(nn.dataNodeInfos[srcChunkNodes[i]].replicationChunkNames,
					&Replication{
						filePathChunkName: chunkMeta.ChunkName,
						toAddress:         desChunkNodes[i]})
		}
		fmt.Printf("NameNode ReplicationBalance: %v; \n", nn.dataNodeInfos[srcChunkNodes[i]].replicationChunkNames)
	}
}

func (nn *NameNode) findNoneChunkOfDataNode(chunkName string) string {
	for name, info := range nn.dataNodeInfos {
		if info.Status == datanodeUp {
			chunkMetas := nn.dataNodeChunks[name]
			if !chunkMetasHasChunk(chunkMetas, chunkName) {
				return name
			}
		}
	}
	return ""
}

func chunkMetasHasChunk(chunkMetas []*ChunkMeta, chunkName string) bool {
	for _, meta := range chunkMetas {
		if meta.ChunkName == chunkName {
			return true
		}
	}
	return false
}
