package service

import (
	"errors"
	"fmt"
	"github.com/kebukeYi/TrainFS/common"
	proto "github.com/kebukeYi/TrainFS/profile"
	"time"
)

const (
	datanodeDown = datanodeStatus("datanodeDown")
	datanodeUp   = datanodeStatus("datanodeUp")
)

const (
	DataNodeKey    string = "DataNodeKey"
	trashKey       string = "trash"
	replicationKey string = "replication"
)

type DataNodeInfo struct {
	Address               string
	FreeSpace             uint64
	HeartBeatTimeStamp    int64
	Status                datanodeStatus
	replicationChunkNames []*Replication // chunk复制任务
	trashChunkNames       []string       // chunk删除任务
}

type Replication struct {
	// 字段需要首字母大写,否则:gob: type service.Replication has no exported fields
	FilePathName      string
	FilePathChunkName string
	ToAddress         string
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
	trashs, err := nn.taskStore.GetTrashes(trashsKey)
	if err != nil {
		fmt.Printf("NameNode GetTrashes key:%s ; value: %v ; err:%s \n", trashsKey, trashs, err)
		return nil, err
	}
	fmt.Printf("NameNode GetTrashes key:%s ; value: %v; \n", trashsKey, trashs)
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
	//nn.metaStore.PutDataNodeMeta(DataNodeListKey, nn.dataNodeInfos)
	fmt.Println("NameNode rev " + arg.DataNodeAddress + " register success!")
	return &proto.DataNodeRegisterReply{Success: true}, nil
}

func (nn *NameNode) HeartBeat(arg *proto.HeartBeatArg) (*proto.HeartBeatReply, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	heartBeatReply := &proto.HeartBeatReply{}
	if dataNodeInfo, ok := nn.dataNodeInfos[arg.DataNodeAddress]; ok {
		dataNodeInfo.HeartBeatTimeStamp = time.Now().UnixMilli()
		if dataNodeInfo.replicationChunkNames != nil && len(dataNodeInfo.replicationChunkNames) > 0 {
			heartBeatReply.FilePathNames = make([]string, len(dataNodeInfo.replicationChunkNames))
			heartBeatReply.FilePathChunkNames = make([]string, len(dataNodeInfo.replicationChunkNames))
			heartBeatReply.NewChunkSevers = make([]string, len(dataNodeInfo.replicationChunkNames))
			for i, replication := range dataNodeInfo.replicationChunkNames {
				heartBeatReply.FilePathNames[i] = replication.FilePathName
				heartBeatReply.FilePathChunkNames[i] = replication.FilePathChunkName
				heartBeatReply.NewChunkSevers[i] = replication.ToAddress
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
		// todo 打印心跳日志
		//fmt.Printf("NameNode rev HeartBeat from %s; reply: %v \n", arg.DataNodeAddress, heartBeatReply)
		//replicaKey := GetDataNodeReplicaKey(arg.DataNodeAddress)
		//nn.taskStore.Delete(replicaKey)
		//trashsKey := GetDataNodeTrashKey(arg.DataNodeAddress)
		//nn.taskStore.Delete(trashsKey)
		return heartBeatReply, nil
	}
	// dataNode 不存在, 需要走注册流程;
	return nil, common.ErrHeartBeatNotExist
}

func (nn *NameNode) ChunkReport(arg *proto.FileLocationInfo) (*proto.ChunkReportReply, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	dataNodeAddress := arg.GetDataNodeAddress()
	chunks := arg.GetChunks() // 当上传的chunk 在NameNode 中不存在时, 可进行下发删除任务!
	fmt.Printf("NameNode rev ChunkReport: %v, from %s \n", chunks, dataNodeAddress)
	for _, chunk := range chunks {
		filePathName := common.GetFileNameFromChunkName(chunk.FilePathChunkName)
		_, err := nn.metaStore.GetFileMeta(filePathName)
		if errors.Is(err, common.ErrFileNotFound) {
			fmt.Printf("NameNode rev ChunkReport from ip:%s,find filePathName:%s, FilePathChunkName:%s not exist at dataStore, so dataNode[%s] need to delete it after! \n",
				dataNodeAddress, filePathName, chunk.FilePathChunkName, dataNodeAddress)
			// dataNode 上传的信息, nameNode 没有此信息, 下发删除任务;
			// 什么情境下会发生这样的事情? 用户删除文件, 但是此dataNode下线了,没有及时执行删除任务;
			nodeInfo := nn.dataNodeInfos[dataNodeAddress]
			nodeInfo.trashChunkNames = append(nodeInfo.trashChunkNames, chunk.FilePathChunkName)
		} else if err != nil {
			fmt.Printf("NameNode rev ChunkReport from ip:%s,find filePathName:%s, FilePathChunkName:%s err:%s! \n",
				dataNodeAddress, filePathName, chunk.FilePathChunkName, err)
			return nil, err
		}
		// todo nameNode 核心, 根据dataNode上传的chunk信息, 构建内存中的文件映射, 需要去重;
		nn.updateChunkLocation(chunk.FilePathChunkName, dataNodeAddress)
		nn.updateDataNodeChunks(chunk, dataNodeAddress)
	}
	return &proto.ChunkReportReply{Success: true}, nil
}

func (nn *NameNode) updateChunkLocation(FilePathChunkName, dataNodeAddress string) {
	replicaMetas := nn.chunkLocation[FilePathChunkName]
	isAdd := true
	for _, meta := range replicaMetas {
		if meta.DataNodeAddress == dataNodeAddress {
			isAdd = false
		}
	}
	if isAdd {
		nn.chunkLocation[FilePathChunkName] = append(nn.chunkLocation[FilePathChunkName],
			&ReplicaMeta{DataNodeAddress: dataNodeAddress})
	}
}

func (nn *NameNode) updateDataNodeChunks(chunk *proto.ChunkInfo, dataNodeAddress string) {
	chunkMetas := nn.dataNodeChunks[dataNodeAddress]
	isAdd := true
	for _, meta := range chunkMetas {
		if meta.ChunkName == chunk.FilePathChunkName {
			isAdd = false
		}
	}
	if isAdd {
		nn.dataNodeChunks[dataNodeAddress] = append(nn.dataNodeChunks[dataNodeAddress], &ChunkMeta{
			ChunkName: chunk.FilePathChunkName,
			ChunkId:   chunk.ChunkId,
			TimeStamp: time.Now().UnixMilli(),
		})
	}
}

// CommitChunk 接收来自DataNode成功保存 chunk 后的提交信息,以便更新内存映射;
func (nn *NameNode) CommitChunk(arg *proto.CommitChunkArg) (*proto.CommitChunkReply, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	// 仅仅是为了保存 fileName 被分成的几个chunk块名字;当存在chunkName时,就说明至少被一个dataNode所存储过了,
	// 其他需要更新 副本复制, 副本删除任务进度; dataNode每次重启都会上报自己所存储的信息;
	switch arg.Operation {
	case proto.ChunkReplicateStatus_LostToReplicate: // 2.dataNode send;
		//  dataNode宕机下线导致数据丢失,触发复制后的提交;
		nn.HandleStoreFileChunk(arg) // 更新内存映射; 因为就算丢失了,file的chunkName不会变少,还是固定的,不用持久化更新;
		// 需要删除对应的复制任务;
		err := nn.HandleFileChunkReplicateTask(arg)
		fmt.Printf("NameNode rev commitChunk LostToReplicate: chunkId:%d; dataNodeAddress:%s; filePathName:%s ; fileChunkName:%s;\n", arg.ChunkId, arg.DataNodeAddress, arg.FilePathName, arg.FileChunkName)
		if err != nil {
			return &proto.CommitChunkReply{Success: false}, err
		}
		return &proto.CommitChunkReply{Success: true}, nil
	case proto.ChunkReplicateStatus_NormalToReplicate: // 1.client send to dataNode; 2.dataNode send to dataNode;
		fmt.Printf("NameNode rev commitChunk NormalToReplicate: chunkId:%d; dataNodeAddress:%s; filePathName:%s ; fileChunkName:%s;\n",
			arg.ChunkId, arg.DataNodeAddress, arg.FilePathName, arg.FileChunkName)
		nn.HandleStoreFileChunk(arg)           // 更新内存映射;
		err := nn.HandleNormalToReplicate(arg) // 持久化更新file被分成的chunkName;存在则证明至少被一个dataNode所存储了;
		return &proto.CommitChunkReply{Success: true}, err
	case proto.ChunkReplicateStatus_DeleteFileChunk:
		fmt.Printf("NameNode rev commitChunk DeleteFileChunk: chunkId:%d; dataNodeAddress:%s; filePathName:%s ; fileChunkName:%s;\n",
			arg.ChunkId, arg.DataNodeAddress, arg.FilePathName, arg.FileChunkName)
		nn.HandleDeleteFileChunk(arg)
		// 需要删除对应的 '删除任务';
		err := nn.HandleFileChunkDeleteTask(arg)
		if err != nil {
			return &proto.CommitChunkReply{Success: false}, err
		}
		return &proto.CommitChunkReply{Success: true}, nil
	default:
		return &proto.CommitChunkReply{Success: false}, common.ErrCommitChunkType
	}
}

// HandleStoreFileChunk 根据dataNode上传的信息, 仅仅更新内存映射;
func (nn *NameNode) HandleStoreFileChunk(arg *proto.CommitChunkArg) {
	chunkId := arg.ChunkId
	fileSize := arg.FileSize
	dataNodeAddress := arg.GetDataNodeAddress()[0]
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

// HandleDeleteFileChunk 根据dataNode上传的删除文件块信息, 仅仅更新内存映射;
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
	meta, err := nn.metaStore.GetFileMeta(filePathName)
	if err != nil {
		// 最开始 client 向 nameNode 发送过的文件名,因此正常情况下是可以找到的;
		fmt.Printf("NameNode HandleNormalToReplicate nn.dataStore.GetFileMeta(%s); err:%s \n",
			filePathName, err)
		return err
	} else {
		chunks := make([]*ChunkMeta, 0)
		chunks = append(chunks, &ChunkMeta{
			ChunkName: fileChunkName,
			ChunkId:   chunkId,
			TimeStamp: time.Now().UnixMilli(),
		})
		if meta.Chunks != nil || len(meta.Chunks) == 0 {
			// file : file_chunk_0  file_chunk_1  file_chunk_2
			// dataNode-1 commit file_chunk_0
			// dataNode-2 commit file_chunk_0
			for _, chunkMeta := range meta.Chunks {
				// 本地已经有了这个文件块信息;直接返回;
				if chunkMeta.ChunkName == fileChunkName {
					fmt.Printf("NameNode HandleNormalToReplicate nn.dataStore.GetFileMeta(%s),fileChunkName:%s, meta:%v ; \n",
						filePathName, fileChunkName, meta.Chunks)
					return nil
				}
			}
			meta.Chunks = append(meta.Chunks, chunks...)
		} else {
			meta.Chunks = chunks
		}
		// chunk 更新完毕;
		err = nn.metaStore.PutFileMeta(filePathName, meta)
		fmt.Printf("NameNode HandleNormalToReplicate nn.dataStore.PutFileMeta(%s, %v); \n", filePathName, meta)
		if err != nil {
			fmt.Printf("NameNode HandleNormalToReplicate nn.dataStore.PutFileMeta(%s, %v); err:%s \n", filePathName, meta, err)
			return err
		}
	}
	return nil
}

func (nn *NameNode) HandleFileChunkReplicateTask(arg *proto.CommitChunkArg) error {
	dataNodeAddress := arg.GetDataNodeAddress()[0]
	dataNodeInfo := nn.dataNodeInfos[dataNodeAddress]
	replications := dataNodeInfo.replicationChunkNames
	newReplications := make([]*Replication, 0)
	for _, task := range replications {
		if task.FilePathChunkName != arg.FileChunkName {
			newReplications = append(newReplications, task)
		}
	}
	dataNodeInfo.replicationChunkNames = newReplications
	dataNodeReplicaKey := GetDataNodeReplicaKey(dataNodeAddress)
	err := nn.taskStore.PutReplications(dataNodeReplicaKey, newReplications)
	if err != nil {
		return err
	}
	return nil
}

func (nn *NameNode) HandleFileChunkDeleteTask(arg *proto.CommitChunkArg) error {
	dataNodeAddress := arg.GetDataNodeAddress()[0]
	dataNodeInfo := nn.dataNodeInfos[dataNodeAddress]
	trashChunkNames := dataNodeInfo.trashChunkNames
	newTrashNames := make([]string, 0)
	for _, trashName := range trashChunkNames {
		if trashName != arg.FileChunkName {
			newTrashNames = append(newTrashNames, trashName)
		}
	}
	dataNodeInfo.trashChunkNames = newTrashNames
	dataNodeTrashKey := GetDataNodeTrashKey(dataNodeAddress)
	err := nn.taskStore.PutTrashes(dataNodeTrashKey, newTrashNames)
	if err != nil {
		return err
	}
	return nil
}

func (nn *NameNode) LiveDetection(*proto.LiveDetectionArg) (*proto.LiveDetectionReply, error) {
	return &proto.LiveDetectionReply{Success: true}, nil
}

func (nn *NameNode) CheckHeartBeat() {
	for {
		for address, dataNodeInfo := range nn.dataNodeInfos {
			// 距离上一次的上线时间超过心跳检测时间,则认为该dataNode已经宕机;
			if time.Now().UnixMilli()-dataNodeInfo.HeartBeatTimeStamp >
				int64(nn.Config.Config.DataNodeHeartBeatTimeout) && dataNodeInfo.Status == datanodeUp {
				fmt.Printf("NameNode CheckHeartBeat() dataNode:%s down! now dataNodeInfos[%v] \n",
					address, nn.dataNodeInfos)
				// todo 要不要持久化 dataNode 的状态,以及NameNode中的内存队列
				// 执行dataNode下线后的资源清理工作;
				go nn.ReplicationBalance(address)
			}
		}
		time.Sleep(time.Duration(nn.Config.Config.DataNodeHeartBeatInterval) * time.Millisecond)
	}
}

func (nn *NameNode) ReplicationBalance(downDataNodeAddress string) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	nn.dataNodeInfos[downDataNodeAddress].Status = datanodeDown
	// 下线dataNode节点所掌握的chunk信息;
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
		if srcChunkNodes[i] == "" {
			// 该chunk没有可用的dataNode; 需记录在案;
			continue
		}
		srcDataNode := srcChunkNodes[i]
		desChunkNodes[i] = nn.findNoneChunkOfDataNode(chunkMeta.ChunkName)
		filePathChunkNames[i] = chunkMeta.ChunkName
		filePathNames[i] = common.GetFileNameFromChunkName(chunkMeta.ChunkName)
		if srcChunkNodes[i] != "" && desChunkNodes[i] != "" {
			nn.dataNodeInfos[srcDataNode].replicationChunkNames =
				append(nn.dataNodeInfos[srcDataNode].replicationChunkNames,
					&Replication{
						FilePathName:      filePathNames[i],
						FilePathChunkName: chunkMeta.ChunkName,
						ToAddress:         desChunkNodes[i]})
			fmt.Printf("NameNode ReplicationBalance success: srcDataNode:%s -> desDataNode:%s; fileChunkName:%s \n",
				srcChunkNodes[i], desChunkNodes[i], chunkMeta.ChunkName)
		} else {
			// 该chunk没有可用的dataNode; 需记录在案;
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
