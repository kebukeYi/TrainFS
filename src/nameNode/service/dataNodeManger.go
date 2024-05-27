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
	DataNodeListKey string = "DataNodeListKey"
	DataNodeKey     string = "DataNodeKey"
	trashKey        string = "trash"
	replicationKey  string = "replication"
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
	fmt.Printf("NameNode rev " + arg.DataNodeAddress + " register ...\n")
	dataNodeInfo := nn.dataNodeInfos[arg.DataNodeAddress]
	if dataNodeInfo == nil {
		dataNodeInfo = &DataNodeInfo{
			Address:               arg.DataNodeAddress,
			replicationChunkNames: nil,
			trashChunkNames:       nil,
		}
	}
	dataNodeInfo.FreeSpace = arg.FreeSpace
	dataNodeInfo.HeartBeatTimeStamp = time.Now().UnixMilli()
	dataNodeInfo.Status = datanodeUp
	// key: DataNodeKey_127.0.0.1_trash  value: [chunk_1, chunk_2]
	trashsKey := GetDataNodeTrashKey(arg.DataNodeAddress)
	trashs, err := nn.taskStore.GetTrashs(trashsKey)
	if err != nil {
		fmt.Printf("NameNode GetTrashs key:%s ; value: %v ; err:%s \n", trashsKey, trashs, err)
		return nil, err
	}
	fmt.Printf("NameNode GetTrashs key:%s ; value: %v; \n", trashsKey, trashs)
	dataNodeInfo.trashChunkNames = trashs
	// key: DataNodeKey_127.0.0.1_trash  value: [chunk_1, chunk_2]
	replicaKey := GetDataNodeReplicaKey(arg.DataNodeAddress)
	replications, err := nn.taskStore.GetReplications(replicaKey)
	if err != nil {
		fmt.Printf("NameNode GetReplications key:%s ; value: %v; err:%s  \n", replicaKey, replications, err)
		return nil, err
	}
	fmt.Printf("NameNode GetReplications key:%s ; value: %v; \n", replicaKey, replications)
	dataNodeInfo.replicationChunkNames = replications

	nn.dataNodeInfos[arg.DataNodeAddress] = dataNodeInfo
	// todo 之后做
	// nn.dataStore.PutDataNodeMeta(DataNodeListKey, nn.dataNodeInfos)
	fmt.Println("NameNode rev " + arg.DataNodeAddress + " register success!")
	return &proto.DataNodeRegisterReply{Success: true}, nil
}

func (nn *NameNode) HeartBeat(arg *proto.HeartBeatArg) (*proto.HeartBeatReply, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	heartBeatReply := &proto.HeartBeatReply{}
	//fmt.Printf("NameNode rev " + arg.DataNodeAddress + " heartBeat...\n")
	if dataNodeInfo, ok := nn.dataNodeInfos[arg.DataNodeAddress]; ok {
		//fmt.Printf("NameNode rev " + arg.DataNodeAddress + " heartBeat success!\n")
		dataNodeInfo.HeartBeatTimeStamp = time.Now().UnixMilli()
		if dataNodeInfo.replicationChunkNames != nil && len(dataNodeInfo.replicationChunkNames) > 0 {
			heartBeatReply.FilePathNames = make([]string, len(dataNodeInfo.replicationChunkNames))
			heartBeatReply.FilePathChunkNames = make([]string, len(dataNodeInfo.replicationChunkNames))
			heartBeatReply.NewChunkSevers = make([]string, len(dataNodeInfo.replicationChunkNames))
			for i, replication := range dataNodeInfo.replicationChunkNames {
				heartBeatReply.FilePathNames[i] = replication.filePathName
				heartBeatReply.FilePathChunkNames[i] = replication.filePathChunkName
				heartBeatReply.NewChunkSevers[i] = replication.toAddress
			}
			dataNodeInfo.replicationChunkNames = make([]*Replication, 0)
		}
		if dataNodeInfo.trashChunkNames != nil && len(dataNodeInfo.trashChunkNames) > 0 {
			heartBeatReply.TrashFilePathChunkNames = make([]string, len(dataNodeInfo.trashChunkNames))
			for i, trash := range dataNodeInfo.trashChunkNames {
				heartBeatReply.TrashFilePathChunkNames[i] = trash
			}
			dataNodeInfo.trashChunkNames = make([]string, 0)
		}
		// todo 打印
		fmt.Printf("NameNode rev HeartBeat from %s; reply: %v \n", arg.DataNodeAddress, heartBeatReply)
		replicaKey := GetDataNodeReplicaKey(arg.DataNodeAddress)
		nn.taskStore.Delete(replicaKey)
		trashsKey := GetDataNodeTrashKey(arg.DataNodeAddress)
		nn.taskStore.Delete(trashsKey)
		return heartBeatReply, nil
	}
	_, err := nn.RegisterDataNode(&proto.DataNodeRegisterArg{DataNodeAddress: arg.DataNodeAddress, FreeSpace: arg.FreeSpace})
	return nil, err
}

func (nn *NameNode) ChunkReport(arg *proto.FileLocationInfo) (*proto.ChunkReportReply, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	dataNodeAddress := arg.GetDataNodeAddress()
	chunks := arg.GetChunks() // 当上传的chunk 在NameNode 中不存在时, 可进行下发删除任务!
	fmt.Printf("NameNode rev ChunkReport: %v, from %s \n", chunks, dataNodeAddress)
	for _, chunk := range chunks {
		filePathName := common.GetFileNameFromChunkName(chunk.FilePathChunkName)
		meta, err := nn.dataStore.GetFileMeta(filePathName)
		if err != nil {
			fmt.Printf("NameNode rev ChunkReport from ip:%s find filePathName:%s , FilePathChunkName:%s not exist at dataStore, so dataNode[%s] need to delete it! \n",
				filePathName, chunk.FilePathChunkName, dataNodeAddress)
			nodeInfo := nn.dataNodeInfos[dataNodeAddress]
			nodeInfo.trashChunkNames = append(nodeInfo.trashChunkNames, chunk.FilePathChunkName)
			return nil, err
		}
		if meta == nil {
			fmt.Printf("NameNode rev chunkReport find filePathName:%s , FilePathChunkName:%s not exist at dataStore, so dataNode[%s] need to delete it! \n",
				filePathName, chunk.FilePathChunkName, dataNodeAddress)
			nodeInfo := nn.dataNodeInfos[dataNodeAddress]
			nodeInfo.trashChunkNames = append(nodeInfo.trashChunkNames, chunk.FilePathChunkName)
			// 刚开机，dataNode应该不会立马宕机，就算重启，还是会重新计算
			//nodeTrashKey := GetDataNodeTrashKey(nodeInfo.Address)
			//err = nn.taskStore.PutTrashs(nodeTrashKey, nodeInfo.trashChunkNames)
		} else {
			// todo Name 核心，构建文件映射
			nn.chunkLocation[chunk.FilePathChunkName] = append(nn.chunkLocation[chunk.FilePathChunkName],
				&ReplicaMeta{DataNodeAddress: dataNodeAddress})
			nn.dataNodeChunks[dataNodeAddress] = append(nn.dataNodeChunks[dataNodeAddress], &ChunkMeta{
				ChunkName: chunk.FilePathChunkName,
				ChunkId:   chunk.ChunkId,
				TimeStamp: time.Now().UnixMilli(),
			})
		}
	}
	return &proto.ChunkReportReply{Success: true}, nil
}

func (nn *NameNode) CommitChunk(arg *proto.CommitChunkArg) (*proto.CommitChunkReply, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	// 仅仅是为了保存 fileName的 几个块名字, 其他不用保存.
	switch arg.Operation {
	// A dataNode sends fileChunk data whose meta already exists in the nameNode to other dataNodes for replication due to a downed dataNode.
	// The dataNode sends fileChunk data for which the meta already exists in the nameNode to other dataNodes to replicate, as a result of a dataNode being down.
	// DataNode sends fileChunk data which meta already exists in nameNode to other dataNode for replication because the down dataNode.
	case proto.ChunkReplicateStatus_LostToReplicate:
		nn.HandleStoreFileChunk(arg)
		fmt.Printf("NameNode rev commitChunk LostToReplicate: chunkId:%d; dataNodeAddress:%s; filePathName:%s ; fileChunkName:%s;\n",
			arg.ChunkId, arg.DataNodeAddress, arg.FilePathName, arg.FileChunkName)
		return &proto.CommitChunkReply{Success: true}, nil
	case proto.ChunkReplicateStatus_NormalToReplicate: // 1. client send //2. dataNode send
		fmt.Printf("NameNode rev commitChunk NormalToReplicate: chunkId:%d; dataNodeAddress:%s; filePathName:%s ; fileChunkName:%s;\n",
			arg.ChunkId, arg.DataNodeAddress, arg.FilePathName, arg.FileChunkName)
		nn.HandleStoreFileChunk(arg)
		err := nn.HandleNormalToReplicate(arg)
		return &proto.CommitChunkReply{Success: true}, err
	case proto.ChunkReplicateStatus_DeleteFileChunk:
		fmt.Printf("NameNode rev commitChunk DeleteFileChunk: chunkId:%d; dataNodeAddress:%s; filePathName:%s ; fileChunkName:%s;\n",
			arg.ChunkId, arg.DataNodeAddress, arg.FilePathName, arg.FileChunkName)
		nn.HandleDeleteFileChunk(arg)
		return &proto.CommitChunkReply{Success: true}, nil
	default:
		return &proto.CommitChunkReply{Success: false}, common.ErrCommitChunkType
	}
}

func (nn *NameNode) HandleStoreFileChunk(arg *proto.CommitChunkArg) {
	chunkId := arg.ChunkId
	fileSize := arg.FileSize
	dataNodeAddress := arg.GetDataNodeAddress()[0]
	//fileChunkName := arg.FilePathName
	fileChunkName := arg.FileChunkName
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
	nn.dataNodeInfos[dataNodeAddress].FreeSpace -= uint64(fileSize)
}

func (nn *NameNode) HandleDeleteFileChunk(arg *proto.CommitChunkArg) {
	fileSize := arg.FileSize
	dataNodeAddress := arg.GetDataNodeAddress()[0]
	fileChunkName := arg.FileChunkName
	delete(nn.chunkLocation, fileChunkName)
	chunkMetas := nn.dataNodeChunks[dataNodeAddress]
	newChunks := make([]*ChunkMeta, 0)
	for _, chunk := range chunkMetas {
		if chunk.ChunkName != fileChunkName {
			newChunks = append(newChunks, chunk)
		}
	}
	nn.dataNodeChunks[dataNodeAddress] = newChunks
	nn.dataNodeInfos[dataNodeAddress].FreeSpace += uint64(fileSize)
}

func (nn *NameNode) HandleNormalToReplicate(arg *proto.CommitChunkArg) error {
	chunkId := arg.GetChunkId()
	filePathName := arg.GetFilePathName()
	fileChunkName := arg.GetFileChunkName()
	meta, err := nn.dataStore.GetFileMeta(filePathName)
	if err != nil {
		fmt.Printf("NameNode HandleNormalToReplicate nn.dataStore.GetFileMeta(%s); err:%s \n", filePathName, err)
		return err
	} else {
		chunks := make([]*ChunkMeta, 0)
		chunks = append(chunks, &ChunkMeta{
			ChunkName: fileChunkName,
			ChunkId:   chunkId,
			TimeStamp: time.Now().UnixMilli(),
		})
		if meta != nil {
			// file : file_chunk_0 file_chunk_1 file_chunk_2
			// dataNode-1 commit file_chunk_0
			// dataNode-2 commit file_chunk_0
			for _, chunkMeta := range meta.Chunks {
				if chunkMeta.ChunkName == fileChunkName {
					fmt.Printf("NameNode HandleNormalToReplicate nn.dataStore.GetFileMeta(%s),fileChunkName:%s, meta:%v ; \n",
						filePathName, fileChunkName, meta.Chunks)
					return nil
				}
			}
			meta.Chunks = append(meta.Chunks, chunks...)
		} else {
			meta = &FileMeta{
				FileName:    filePathName,
				KeyFileName: filePathName,
				FileSize:    arg.GetFileSize(),
				IsDir:       false,
				Chunks:      chunks,
			}
		}
		err = nn.dataStore.PutFileMeta(filePathName, meta)
		fmt.Printf("NameNode HandleNormalToReplicate nn.dataStore.PutFileMeta(%s, %v); \n", filePathName, meta)
		if err != nil {
			fmt.Printf("NameNode HandleNormalToReplicate nn.dataStore.PutFileMeta(%s, %v); err:%s \n", filePathName, meta, err)
			return err
		}
	}
	return nil
}

func (nn *NameNode) LiveDetection(*proto.LiveDetectionArg) (*proto.LiveDetectionReply, error) {
	return &proto.LiveDetectionReply{Success: true}, nil
}

func (nn *NameNode) CheckHeartBeat() {
	for {
		//nn.mux.Lock()
		for address, dataNodeInfo := range nn.dataNodeInfos {
			if time.Now().UnixMilli()-dataNodeInfo.HeartBeatTimeStamp >
				int64(nn.Config.NameNode.DataNodeHeartBeatTimeout) && dataNodeInfo.Status == datanodeUp {
				//dataNodeInfo.Status = datanodeDown
				fmt.Printf("NameNode CheckHeartBeat() dataNode %s down! now dataNodeInfos[%v] \n", address, nn.dataNodeInfos)
				// todo 要不要持久化 dataNode-1 的状态,以及NameNode中的内存队列
				go nn.ReplicationBalance(address)
			}
		}
		//nn.mux.Unlock()
		time.Sleep(time.Duration(nn.Config.NameNode.DataNodeHeartBeatInterval) * time.Millisecond)
	}
}

func (nn *NameNode) ReplicationBalance(downDataNodeAddress string) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	nn.dataNodeInfos[downDataNodeAddress].Status = datanodeDown
	needCopyChunkMetas := nn.dataNodeChunks[downDataNodeAddress]
	srcChunkNodes := make([]string, len(needCopyChunkMetas))
	desChunkNodes := make([]string, len(needCopyChunkMetas))
	filePathChunkNames := make([]string, len(needCopyChunkMetas))
	filePathNames := make([]string, len(needCopyChunkMetas))

	for i, chunkMeta := range needCopyChunkMetas {
		replicaMetas := nn.chunkLocation[chunkMeta.ChunkName]
		for _, replicaMeta := range replicaMetas {
			if replicaMeta.DataNodeAddress != downDataNodeAddress &&
				nn.dataNodeInfos[replicaMeta.DataNodeAddress].Status == datanodeUp {
				srcChunkNodes[i] = replicaMeta.DataNodeAddress
				break
			}
		}
		srcDataNode := srcChunkNodes[i]
		desChunkNodes[i] = nn.findNoneChunkOfDataNode(chunkMeta.ChunkName)
		filePathChunkNames[i] = chunkMeta.ChunkName
		filePathNames[i] = common.GetFileNameFromChunkName(chunkMeta.ChunkName)
		if srcChunkNodes[i] != "" && desChunkNodes[i] != "" {
			nn.dataNodeInfos[srcDataNode].replicationChunkNames =
				append(nn.dataNodeInfos[srcDataNode].replicationChunkNames,
					&Replication{
						filePathName:      filePathNames[i],
						filePathChunkName: chunkMeta.ChunkName,
						toAddress:         desChunkNodes[i]})
			fmt.Printf("NameNode ReplicationBalance success: srcDataNode:%s -> desDataNode:%s; fileChunkName:%s \n",
				srcChunkNodes[i], desChunkNodes[i], chunkMeta.ChunkName)
		} else {
			fmt.Printf("NameNode ReplicationBalance failed: srcDataNode:%s -> desDataNode:%s; fileChunkName:%s \n",
				srcChunkNodes[i], desChunkNodes[i], chunkMeta.ChunkName)
		}
	}
}

func (nn *NameNode) findNoneChunkOfDataNode(chunkName string) string {
	for address, info := range nn.dataNodeInfos {
		if info.Status == datanodeUp {
			chunkMetas := nn.dataNodeChunks[address]
			if !chunkMetasHasChunk(chunkMetas, chunkName) {
				return address
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

func GetDataNodeTrashKey(address string) string {
	return DataNodeKey + "_" + address + "_" + trashKey
}

func GetDataNodeReplicaKey(address string) string {
	return DataNodeKey + "_" + address + "_" + replicationKey
}
