package service

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"sort"
	"strings"
	"sync"
	"trainfs/src/common"
	"trainfs/src/nameNode/config"
	proto "trainfs/src/profile"
)

type NameNode struct {
	Config *config.NameNodeConfig
	mux    sync.RWMutex
	DB     *StateMachine

	// <fileName1, [chunk1,chunk2,chunk3]>
	// <fileName2, [chunk4,chunk5,chunk6]>
	// fileChunks map[string][]string

	// <chunk1, [dataNode1,dataNode2,dataNode3]>
	// <chunk2, [dataNode2,dataNode3,dataNode4]>
	chunkLocation map[string][]*ReplicaMeta

	// <dataNode1, [chunk1,chunk2,chunk3]>
	// <dataNode2, [chunk4,chunk5,chunk3]>
	dataNodeChunks map[string][]*ChunkMeta

	// <dataNode1, info>
	dataNodeInfos map[string]*DataNodeInfo
}

func NewNameNode() *NameNode {
	nameNode := &NameNode{}
	nameNode.Config = config.GetDataNodeConfig()
	db := OpenStateMachine(nameNode.Config.NameNode.DataDir)
	nameNode.DB = db
	nameNode.chunkLocation = make(map[string][]*ReplicaMeta)
	nameNode.dataNodeChunks = make(map[string][]*ChunkMeta)
	nameNode.dataNodeInfos = make(map[string]*DataNodeInfo)
	return nameNode
}

func (nn *NameNode) PutFile(arg *proto.FileOperationArg) (*proto.DataNodeChain, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	// arg.GetFileName() : fileName: /user/app/example.txt
	// path: [user,app]
	// fileName: example.txt
	pathFileName := arg.GetFileName()
	fileName, path := splitFileNamePath(pathFileName)
	parentNode, err := nn.checkPathOrCreate(path, true)
	if err != nil {
		return nil, err
	}
	if _, ok := parentNode.ChildList[fileName]; ok {
		return nil, common.ErrFileAlreadyExists
	}
	fileMeta := &FileMeta{
		FileName:    fileName,                                           // example.txt
		KeyFileName: CreatKeyFileName(parentNode.KeyFileName, fileName), // /user/app/example.txt
		FileSize:    arg.GetFileSize(),
		IsDir:       false,
		ChunkNum:    arg.ChunkNum,
	}
	parentNode.ChildList[fileName] = fileMeta
	// todo 等待DataNode commitChunk 后再持久化
	// todo 不等待DataNode commitChunk 后再持久化,先持久化
	err = nn.DB.PutFileMeta(fileMeta.KeyFileName, fileMeta)
	if err != nil {
		return nil, err
	}
	err = nn.DB.PutFileMeta(parentNode.KeyFileName, parentNode)
	if err != nil {
		return nil, err
	}
	choseDataNodes, err := nn.choseDataNode(int(arg.ReplicaNum))
	if err != nil {
		return nil, err
	}
	//dataNode := make([]string, len(choseDataNodes))
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
	if filePath == "" {
		return nil, common.ErrPathFormatError
	}
	nn.mux.RLock()
	defer nn.mux.RUnlock()
	fileMeta, err := nn.DB.GetFileMeta(filePath)
	if err != nil {
		fmt.Printf("NameNode rev GetFile(%s), err:%v;\n", filePath, err)
		return nil, err
	}
	fmt.Printf("NameNode rev GetFile(%s), db FileLocationInfo:%v;\n", filePath, fileMeta.Chunks)
	fmt.Printf("NameNode  nn.chunkLocation: %v; \n", nn.chunkLocation)
	fileLocationInfo := &proto.FileLocationInfo{}
	fileLocationInfo.FileSize = fileMeta.FileSize
	fileLocationInfo.ChunkNum = int64(len(fileMeta.Chunks))
	chunkInfos := make([]*proto.ChunkInfo, 0)
	for _, chunkMeta := range fileMeta.Chunks {
		if r, ok := nn.chunkLocation[chunkMeta.ChunkName]; ok {
			address := make([]string, 0)
			for _, replica := range r {
				address = append(address, replica.DataNodeAddress)
			}
			chunkInfos = append(chunkInfos, &proto.ChunkInfo{
				ChunkId:           chunkMeta.ChunkId,
				FilePathName:      filePath,
				FilePathChunkName: chunkMeta.ChunkName,
				DataNodeAddress:   &proto.DataNodeChain{DataNodeAddress: address},
			})
		} else {
			return nil, common.ErrChunkReplicaNotFound
		}
		//return nil, common.ErrChunkReplicaNotFound
	}
	fileLocationInfo.Chunks = chunkInfos
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
	fileName := arg.GetFileName()
	if fileName == "/" {
		return nil, common.ErrCanNotChangeRootDir
	}
	// /root   / root
	// /tt.txt / ttt.txt
	fileName, path := splitFileNamePath(arg.FileName)
	parentNode, err := nn.checkPathOrCreate(path, false)
	if err != nil {
		return nil, err
	}
	meta := parentNode.ChildList[fileName]
	if meta == nil {
		return nil, common.ErrFileNotFound
	}
	delete(parentNode.ChildList, fileName)
	nn.DB.PutFileMeta(parentNode.KeyFileName, parentNode)
	fileMeta, err := nn.DB.GetFileMeta(fileName)
	if err != nil {
		return nil, err
	}
	if fileMeta.IsDir {
		err = nn.DB.Delete(fileName)
		return nil, err
	} else {
		for _, chunk := range fileMeta.Chunks {
			replicaMetas := nn.chunkLocation[chunk.ChunkName]
			for _, replicaMeta := range replicaMetas {
				nodeInfo := nn.dataNodeInfos[replicaMeta.DataNodeAddress]
				nodeInfo.trashChunkNames = append(nodeInfo.trashChunkNames, chunk.ChunkName)
				fmt.Printf("NameNode delete chunk:%s, replica:%s ;\n",
					chunk.ChunkName, replicaMeta.DataNodeAddress)
			}
		}
	}
	fmt.Printf("NameNode rev DeleteFile(%s);\n", fileName)
	return nil, nil
}

func (nn *NameNode) ListDir(arg *proto.FileOperationArg) (*proto.DirMetaList, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	fileName := arg.GetFileName()
	fileMeta, err := nn.DB.GetFileMeta(fileName)
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
	dirList := make([]*proto.FileMeta, len(childList))
	for _, meta := range childList {
		dirList = append(dirList, &proto.FileMeta{
			KeyFileName: meta.KeyFileName,
			FileName:    meta.FileName,
			FileSize:    meta.FileSize,
			IsDir:       meta.IsDir,
		})
	}
	dirMetaList := &proto.DirMetaList{MetaList: dirList}
	return dirMetaList, nil
}

func (nn *NameNode) ReName(arg *proto.FileOperationArg) (*proto.ReNameReply, error) {
	nn.mux.Lock()
	defer nn.mux.Unlock()
	oldKeyFilePathName := arg.GetFileName()
	newKeyFilePathName := arg.GetNewFileName()
	fileMeta, err := nn.DB.GetFileMeta(oldKeyFilePathName)
	if err != nil {
		return nil, err
	}
	if fileMeta == nil {
		return nil, common.ErrFileNotFound
	}
	oldFileName, path := splitFileNamePath(oldKeyFilePathName)
	newFileName, path := splitFileNamePath(newKeyFilePathName)
	parentNode, err := nn.checkPathOrCreate(path, false)
	if err != nil {
		return nil, err
	}
	fileMeta.FileName = newFileName
	fileMeta.KeyFileName = newKeyFilePathName
	parentNode.ChildList[newFileName] = fileMeta
	delete(parentNode.ChildList, oldFileName)
	err = nn.DB.PutFileMeta(parentNode.KeyFileName, parentNode)
	if err != nil {
		return nil, err
	}
	err = nn.DB.PutFileMeta(fileMeta.KeyFileName, fileMeta)
	if err != nil {
		return nil, err
	}
	return &proto.ReNameReply{Success: true}, nil
}

// path: /test.txt /user/app/3435543423 /user/app/local/yy  /   /root
// path: /user/test/test.txt /user/app/3435543423 /user/app/local/yy  /   /root
// path: /user/test          /user/app            /user/app/local/yy  /   /root
func (nn *NameNode) checkPathOrCreate(path string, notCreate bool) (*FileMeta, error) {
	if path == "/" {
		fileMeta, err := nn.DB.GetFileMeta(path)
		if err == leveldb.ErrNotFound && notCreate {
			fileMeta := &FileMeta{
				IsDir:       true,
				ChildList:   make(map[string]*FileMeta),
				KeyFileName: path,
				FileName:    path,
				FileSize:    0,
			}
			err = nn.DB.PutFileMeta(path, fileMeta)
			if err != nil {
				return nil, err
			}
		}
		return fileMeta, err
	}
	rootFileMeta, _ := nn.DB.GetFileMeta("/")
	if rootFileMeta == nil {
		rootFileMeta = &FileMeta{
			IsDir:       true,
			ChildList:   make(map[string]*FileMeta),
			KeyFileName: "/",
			FileName:    "/",
			FileSize:    0,
		}
		if err := nn.DB.PutFileMeta(rootFileMeta.KeyFileName, rootFileMeta); err != nil {
			return nil, err
		}
	}
	split := strings.Split(path, "/")[1:]
	for i := 0; i < len(split); i++ {
		dir := split[i] // app
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
				nn.DB.PutFileMeta(fileMeta.KeyFileName, fileMeta)
				rootFileMeta.ChildList[dir] = fileMeta
				nn.DB.PutFileMeta(rootFileMeta.KeyFileName, rootFileMeta)
				rootFileMeta = fileMeta
				continue
			}
			return nil, common.ErrNotDir
		}
		rootFileMeta = file
	}
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
		return nil, common.ErrEnoughReplicaDataNodeServer
	}
	sort.Sort(ByFreeSpace(result))
	return result[:num], nil
}

func CreatKeyFileName(parentFileName, dir string) string {
	if parentFileName == "/" {
		return parentFileName + dir
	} else {
		return parentFileName + "/" + dir
	}
}

func splitFileNamePath(fileNamePath string) (fileName, path string) {
	index := strings.LastIndex(fileNamePath, "/")
	if index < 0 {
		return "", ""
	}
	path = fileNamePath[:index]
	if path == "" {
		path = "/"
		fileName = fileNamePath[index+1:]
		return
	}
	fileName = fileNamePath[index+1:]
	return
}
