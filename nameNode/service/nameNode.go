package service

import (
	"errors"
	"fmt"
	DBcommon "github.com/kebukeYi/TrainDB/common"
	"github.com/kebukeYi/TrainFS/common"
	proto "github.com/kebukeYi/TrainFS/profile"
	"sort"
	"strings"
	"sync"
)

type NameNode struct {
	Config    *NameNodeConfig
	mux       sync.RWMutex
	metaStore *MetaStoreManger
	taskStore *TaskStoreManger

	// <fileName1, [chunk1,chunk2,chunk3]>
	// <fileName2, [chunk4,chunk5,chunk6]>
	// fileChunks map[string][]stringtaskStore

	// <chunk1, [dataNode1,dataNode2,dataNode3]>
	// <chunk2, [dataNode2,dataNode3,dataNode4]>
	chunkLocation map[string][]*ReplicaMeta

	// <dataNode1, [chunk1,chunk2,chunk3]>
	// <dataNode2, [chunk4,chunk5,chunk3]>
	dataNodeChunks map[string][]*ChunkMeta

	// <dataNode1, info>
	dataNodeInfos map[string]*DataNodeInfo
}

func NewNameNode(configFile *string) *NameNode {
	nameNode := &NameNode{}
	nameNode.Config = GetDataNodeConfig(configFile)
	//common.ClearDir(nameNode.Config.Config.DataDir)
	//common.ClearDir(nameNode.Config.Config.TaskDir)
	nameNode.metaStore = OpenDataStoreManger(nameNode.Config.Config.DataDir)
	nameNode.taskStore = OpenTaskStoreManger(nameNode.Config.Config.TaskDir)
	nameNode.chunkLocation = make(map[string][]*ReplicaMeta) // 等待 dataNode 上报的数据;
	nameNode.dataNodeChunks = make(map[string][]*ChunkMeta)  // 等待 dataNode 上报的数据;
	nameNode.dataNodeInfos = make(map[string]*DataNodeInfo)  // dataNode 信息;
	return nameNode
}

func (nn *NameNode) PutFile(arg *proto.FileOperationArg) (*proto.DataNodeChain, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	pathFileName := arg.GetFileName() // fileName: /user/app/example.txt
	// 解析出path: [user,app]
	// fileName: example.txt
	fileName, path := common.SplitFileNamePath(pathFileName)
	// 获得path的最后一个文件夹索引,并且遵循没有就创建原则;
	parentNode, err := nn.checkPathOrCreate(path, true)
	if err != nil {
		return nil, err
	}
	// 改进: 根据参数,是否覆盖原文件; 还是追加文件;
	if _, ok := parentNode.ChildList[fileName]; ok {
		return nil, common.ErrFileAlreadyExists
	}
	// 1. 选择 dataNode节点;
	choseDataNodes, err := nn.choseDataNode(int(arg.ReplicaNum))
	if err != nil {
		return nil, err
	}

	fileMeta := &FileMeta{
		FileName:    fileName,                                           // example.txt
		KeyFileName: CreatKeyFileName(parentNode.KeyFileName, fileName), // /user/app + / + example.txt
		FileSize:    arg.GetFileSize(),                                  // 文件总大小;
		IsDir:       false,
		ChunkNum:    arg.ChunkNum,
	}
	parentNode.ChildList[fileName] = fileMeta
	// 2. 保存 最底层文件块信息;
	err = nn.metaStore.PutFileMeta(fileMeta.KeyFileName, fileMeta)
	if err != nil {
		return nil, err
	}
	// 3. 再保存 父节点信息;
	err = nn.metaStore.PutFileMeta(parentNode.KeyFileName, parentNode)
	if err != nil {
		return nil, err
	}

	dataNode := make([]string, 0)
	for _, node := range choseDataNodes {
		dataNode = append(dataNode, node.Address)
	}
	nodeChain := &proto.DataNodeChain{DataNodeAddress: dataNode}
	fmt.Printf("NameNode rev PutFile(%s), return DataNode:%v;\n", pathFileName, nodeChain)
	return nodeChain, nil
}

func (nn *NameNode) ConfirmFile(arg *proto.ConfirmFileArg) (*proto.ConfirmFileReply, error) {
	chunkInfos := arg.FileLocationInfo.Chunks
	reply := &proto.ConfirmFileReply{Success: true}
	nn.mux.RLock()
	defer nn.mux.RUnlock()
	for _, info := range chunkInfos {
		if _, ok := nn.chunkLocation[info.FilePathChunkName]; !ok {
			if arg.Ack {
				reply.Success = false
				reply.Context += info.FilePathChunkName + " is not exist. "
				return reply, nil
			} else {
				continue
			}
		}
	}
	return reply, nil
}

func (nn *NameNode) GetFile(arg *proto.FileOperationArg) (*proto.FileLocationInfo, error) {
	return nn.GetFileLocation(arg.GetFileName(), arg.ReplicaNum)
}

func (nn *NameNode) GetFileLocation(filePath string, replicaNum int32) (*proto.FileLocationInfo, error) {
	if filePath == "" || filePath == "/" {
		return nil, common.ErrPathFormatError
	}
	nn.mux.RLock()
	defer nn.mux.RUnlock()
	// filePath : /usr/app/example.txt
	fileMeta, err := nn.metaStore.GetFileMeta(filePath)
	if err != nil {
		fmt.Printf("NameNode rev GetFile(%s), err:%v;\n", filePath, err)
		return nil, err
	}
	if fileMeta == nil {
		return nil, common.ErrFileNotFound
	}
	if fileMeta.IsDir {
		return nil, common.ErrFileFormatError
	}
	// 获得文件块信息;
	fmt.Printf("NameNode rev GetFile(%s), dataStore.FileLocationInfo:%v;\n", filePath, fileMeta.Chunks)
	// 文件块信息映射 dataNode 地址;
	//fmt.Printf("NameNode.chunkLocations: %v; \n", nn.chunkLocation)
	fileLocationInfo := &proto.FileLocationInfo{}
	fileLocationInfo.FileSize = fileMeta.FileSize           // 文件总大小;
	fileLocationInfo.ChunkNum = int64(len(fileMeta.Chunks)) // 文件几个块;
	chunkInfos := make([]*proto.ChunkInfo, 0)
	// 开始遍历文件块信息, 来获得对应的dataNode信息;
	for _, chunkMeta := range fileMeta.Chunks { // 可能此时的chunks; 只有1/3,其余的还没上传;
		// 获得文件块信息对应的dataNode信息;同一个块,返回多个dataNode信息;
		if dataNodes, ok := nn.chunkLocation[chunkMeta.ChunkName]; ok {
			addresses := make([]string, 0)
			// 每个文件块都存在于多数dataNode上,需要在线的dataNode;
			for _, dataNode := range dataNodes {
				if nodeInfo, ok := nn.dataNodeInfos[dataNode.DataNodeAddress]; ok {
					if nodeInfo.Status == datanodeUp {
						addresses = append(addresses, dataNode.DataNodeAddress)
					}
				}
			}

			// 找到在线dataNode后, 封装文件chunkInfo;
			chunkInfos = append(chunkInfos, &proto.ChunkInfo{
				ChunkId:           chunkMeta.ChunkId,
				FilePathName:      filePath,
				FilePathChunkName: chunkMeta.ChunkName,
				DataNodeAddress:   &proto.DataNodeChain{DataNodeAddress: addresses},
			})
		} else {
			// 相关chunk没有找到, 说明 dataNode 还没提交相关文件块信息;
			return nil, common.ErrChunkReplicaNotFound
		}
	}
	fileLocationInfo.Chunks = chunkInfos // 可能为空;
	fmt.Printf("NameNode rev GetFile(%s), return FileLocationInfo:%v;\n", filePath, fileLocationInfo)
	return fileLocationInfo, nil
}

func (nn *NameNode) GetFileStoreChain(arg *proto.FileOperationArg) (*proto.DataNodeChain, error) {
	dataNodes, err := nn.choseDataNode(int(arg.ReplicaNum))
	if err != nil {
		return nil, err
	}
	var result []string
	for _, dataNode := range dataNodes {
		result = append(result, dataNode.Address)
	}
	return &proto.DataNodeChain{DataNodeAddress: result}, nil
}

func (nn *NameNode) DeleteFile(arg *proto.FileOperationArg) (*proto.DeleteFileReply, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	pathFileName := arg.FileName
	if pathFileName == "/" {
		return nil, common.ErrCanNotChangeRootDir
	}
	// /root   / root
	// /tt.txt / ttt.txt
	fileName, path := common.SplitFileNamePath(pathFileName)
	parentNode, err := nn.checkPathOrCreate(path, false)
	if err != nil {
		return nil, err
	}
	meta := parentNode.ChildList[fileName]
	if meta == nil {
		return nil, common.ErrFileNotFound
	}
	// 1. 删除的是文件夹类型, 需要进一步判断是否是空目录;
	if meta.IsDir {
		getFile, err := nn.metaStore.GetFileMeta(meta.KeyFileName)
		if err != nil {
			return nil, err
		}
		if len(getFile.ChildList) > 0 {
			return nil, common.ErrCanNotDelNotEmptyDir
		} else {
			err = nn.metaStore.Delete(meta.KeyFileName)
			delete(parentNode.ChildList, fileName)
			nn.metaStore.PutFileMeta(parentNode.KeyFileName, parentNode)
		}
	} else { // 2. 删除的是一个文件, 那么就需要删除文件块信息,并下发给dataNode删除任务;
		meta, err = nn.metaStore.GetFileMeta(meta.KeyFileName) // 尝试获取最新的 dataNode 的上传信息;
		if err != nil {
			return nil, err
		}
		go func() { // 新协程完成删除 文件块信息;
			for _, chunk := range meta.Chunks {
				// 获得这个文件块的所在的几个dataNode信息;
				replicaMetas := nn.chunkLocation[chunk.ChunkName]
				for _, replicaMeta := range replicaMetas { // 当前 chunk 的几个副本;
					nodeInfo := nn.dataNodeInfos[replicaMeta.DataNodeAddress]
					nodeInfo.trashChunkNames = append(nodeInfo.trashChunkNames, chunk.ChunkName)
					nodeTrashKey := GetDataNodeTrashKey(nodeInfo.Address)
					// 持久化dataNode的删除任务; 会不会出现覆盖之前的未执行的任务? 不会, 每次保存的都是全量任务;
					// dataNode 执行完任务后, 提交后, nameNode剔除掉删除任务;
					err = nn.taskStore.PutTrashes(nodeTrashKey, nodeInfo.trashChunkNames)
					if err != nil {
						fmt.Printf("NameNode for PutTrashes delete chunk:%s, replicaIP:%s err:%v ;\n",
							chunk.ChunkName, replicaMeta.DataNodeAddress, err)
						return
					}
					fmt.Printf("NameNode for PutTrashes delete chunk:%s, replica:%s ;\n", chunk.ChunkName, replicaMeta.DataNodeAddress)
				}
				// 删除 <chunk*,dataNode>内存映射;
				delete(nn.chunkLocation, chunk.ChunkName)
				// 删除 <dataNode,chunks> 内存映射;
			}
		}()
		// 主协程完成 删除文件信息;
		err = nn.metaStore.Delete(meta.KeyFileName)
		delete(parentNode.ChildList, fileName)
		nn.metaStore.PutFileMeta(parentNode.KeyFileName, parentNode)
		// fmt.Printf("NameNode dataStore.Delete(%s);\n", meta.KeyFileName)
		if err != nil {
			fmt.Printf("NameNode dataStore.Delete(%s), err:%v ;\n", meta.KeyFileName, err)
		}
	}
	fmt.Printf("NameNode.DeleteFile(%s);\n", pathFileName)
	return nil, err
}

func (nn *NameNode) ListDir(arg *proto.FileOperationArg) (*proto.DirMetaList, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	fileName := arg.GetFileName()
	fileMeta, err := nn.metaStore.GetFileMeta(fileName)
	if err != nil {
		return nil, err
	}
	if fileMeta == nil {
		return nil, common.ErrFileNotFound
	}

	if !fileMeta.IsDir {
		return nil, common.ErrNotDir
	}
	childList := fileMeta.ChildList
	dirList := make([]*proto.FileMeta, 0)
	for _, meta := range childList {
		dirList = append(dirList, &proto.FileMeta{
			KeyFileName: meta.KeyFileName,
			FileName:    meta.FileName,
			FileSize:    meta.FileSize,
			IsDir:       meta.IsDir,
		})
	}
	dirMetaList := &proto.DirMetaList{MetaList: dirList}
	fmt.Printf("NameNode rev ListDir(%s), return DirMetaList:%v;\n", fileName, dirMetaList)
	return dirMetaList, nil
}

func (nn *NameNode) Mkdir(arg *proto.FileOperationArg) (*proto.MkdirReply, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	// arg.GetFileName() : fileName: /user/app/example.txt
	// path: [user,app]
	// fileName: example.txt
	pathFileName := arg.GetFileName()
	fileName, path := common.SplitFileNamePath(pathFileName)
	parentNode, err := nn.checkPathOrCreate(path, true)
	if err != nil {
		return nil, err
	}
	if _, ok := parentNode.ChildList[fileName]; ok {
		return nil, common.ErrDirAlreadyExists
	}
	fileMeta := &FileMeta{
		FileName:    fileName,                                           // example.txt
		KeyFileName: CreatKeyFileName(parentNode.KeyFileName, fileName), // /user/app/example.txt
		FileSize:    arg.GetFileSize(),
		IsDir:       true,
		ChildList:   make(map[string]*FileMeta),
	}
	parentNode.ChildList[fileName] = fileMeta
	err = nn.metaStore.PutFileMeta(fileMeta.KeyFileName, fileMeta)
	if err != nil {
		return nil, err
	}
	err = nn.metaStore.PutFileMeta(parentNode.KeyFileName, parentNode)
	if err != nil {
		return nil, err
	}
	return &proto.MkdirReply{Success: true}, nil
}

func (nn *NameNode) ReName(arg *proto.FileOperationArg) (*proto.ReNameReply, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	reply := &proto.ReNameReply{Success: false}
	oldPathFileName := arg.GetFileName()
	newPathFileName := arg.GetNewFileName()
	if oldPathFileName == "/" {
		return nil, common.ErrCanNotChangeRootDir
	}
	// /root/y.data   /root y.data -> not allowed
	// /root/app   /root app ->  /root/newApp   /root newApp
	oldDirName, oldPath := common.SplitFileNamePath(oldPathFileName)
	newDirName, newPath := common.SplitFileNamePath(newPathFileName)
	// 只允许修改子目录名称, 不允许修改目录路径;
	if oldPath != newPath {
		return reply, common.ErrCanNotDelNotEmptyDir
	}
	parentNode, err := nn.checkPathOrCreate(oldPath, false)
	if err != nil {
		return reply, err
	}
	meta := parentNode.ChildList[oldDirName]
	if meta == nil {
		return reply, common.ErrFileNotFound
	}
	// 1. 删除的是文件夹类型, 需要进一步判断是否是空目录;
	if meta.IsDir {
		getFile, err := nn.metaStore.GetFileMeta(meta.KeyFileName)
		if err != nil {
			return reply, err
		}
		if len(getFile.ChildList) > 0 {
			return reply, common.ErrCanNotDelNotEmptyDir
		} else {
			err = nn.metaStore.Delete(meta.KeyFileName)
			delete(parentNode.ChildList, oldDirName)
			// 增加新文件信息;
			fileMeta := &FileMeta{
				FileName:    newDirName,                                           // newDir
				KeyFileName: CreatKeyFileName(parentNode.KeyFileName, newDirName), // /user/app/newDir
				IsDir:       true,
				ChildList:   make(map[string]*FileMeta),
			}
			parentNode.ChildList[newPathFileName] = fileMeta
			err = nn.metaStore.PutFileMeta(fileMeta.KeyFileName, fileMeta)
			if err != nil {
				return reply, err
			}
			err = nn.metaStore.PutFileMeta(parentNode.KeyFileName, parentNode)
			if err != nil {
				return reply, err
			}
		}
	} else {
		return &proto.ReNameReply{Success: false}, common.ErrNotSupported
	}
	return &proto.ReNameReply{Success: true}, nil
}

// path: /test.txt /user/app/3435543423 /user/app/local/yy  /   /root
// path: /user/test/test.txt /user/app/3435543423 /user/app/local/yy  /   /root
// path: /user/test          /user/app            /user/app/local/yy  /   /root
func (nn *NameNode) checkPathOrCreate(path string, notCreate bool) (*FileMeta, error) {
	if path == "/" {
		fileMeta, err := nn.metaStore.GetFileMeta(path)
		if errors.Is(err, DBcommon.ErrKeyNotFound) && notCreate {
			fileMeta := &FileMeta{
				IsDir:       true,
				ChildList:   make(map[string]*FileMeta),
				KeyFileName: path,
				FileName:    path,
				FileSize:    0,
			}
			err = nn.metaStore.PutFileMeta(path, fileMeta)
			if err != nil {
				return nil, err
			}
		}
		return fileMeta, err
	}
	rootFileMeta, _ := nn.metaStore.GetFileMeta("/")
	if rootFileMeta == nil {
		rootFileMeta = &FileMeta{
			IsDir:       true,
			ChildList:   make(map[string]*FileMeta),
			KeyFileName: "/",
			FileName:    "/",
			FileSize:    0,
		}
		if err := nn.metaStore.PutFileMeta(rootFileMeta.KeyFileName, rootFileMeta); err != nil {
			return nil, err
		}
	}
	fileQueue := make([]*FileMeta, 0)
	fileQueue = append(fileQueue, rootFileMeta)
	split := strings.Split(path, "/")[1:]
	for i := 0; i < len(split); i++ {
		dir := split[i] // user
		file, ok := rootFileMeta.ChildList[dir]
		if !ok {
			if notCreate {
				fileMeta := &FileMeta{
					IsDir:       true,
					ChildList:   make(map[string]*FileMeta),
					KeyFileName: CreatKeyFileName(rootFileMeta.KeyFileName, dir),
					FileName:    dir,
					FileSize:    0,
				}
				// KeyFileName: /app  value: app
				nn.metaStore.PutFileMeta(fileMeta.KeyFileName, fileMeta)
				rootFileMeta.ChildList[dir] = fileMeta
				fileQueue = append(fileQueue, fileMeta)
				nn.metaStore.PutFileMeta(rootFileMeta.KeyFileName, rootFileMeta)
				rootFileMeta = fileMeta
				continue
			}
			return nil, common.ErrNotDir
		}
		fileMeta, err := nn.metaStore.GetFileMeta(file.KeyFileName)
		if err != nil {
			return nil, err
		}
		fileQueue = append(fileQueue, fileMeta)
		rootFileMeta = fileMeta
	}
	//for _, meta := range fileQueue {
	//	fmt.Printf("fileQueue: %v; \n", meta)
	//}
	return rootFileMeta, nil
}

func (nn *NameNode) choseDataNode(num int) ([]*DataNodeInfo, error) {
	if num > len(nn.dataNodeInfos) {
		return nil, common.ErrEnoughReplicaDataNodeServer
	}
	var result []*DataNodeInfo
	for _, nodeMeta := range nn.dataNodeInfos {
		if nodeMeta.Status == datanodeUp {
			result = append(result, nodeMeta)
		}
	}
	if num > len(result) {
		return nil, common.ErrEnoughUpDataNodeServer
	}
	sort.Sort(ByFreeSpace(result))
	return result[:num], nil
}

func (nn *NameNode) Close() error {
	if nn.metaStore != nil {
		err := nn.metaStore.close()
		if err != nil {
			return err
		}
	}
	if nn.taskStore != nil {
		err := nn.taskStore.close()
		if err != nil {
			return err
		}
	}
	return nil
}

func CreatKeyFileName(parentFileName, dir string) string {
	if parentFileName == "/" {
		return parentFileName + dir
	} else {
		return parentFileName + "/" + dir
	}
}
