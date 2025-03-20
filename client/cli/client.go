package cli

import (
	"context"
	"fmt"
	"github.com/kebukeYi/TrainFS/client/config"
	"github.com/kebukeYi/TrainFS/common"
	proto "github.com/kebukeYi/TrainFS/profile"
	"google.golang.org/grpc"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

type Client struct {
	conf     *config.ClientConfig
	name     string
	clientId int
	SeqId    int
}

func NewClient() *Client {
	client := &Client{}
	client.conf = config.GetClientConfig()
	client.clientId = client.conf.Client.ClientId
	client.name = "client-" + strconv.Itoa(client.clientId)
	client.SeqId = 0
	return client
}

// PutFile 上传文件;
// localFilePath: /local/file/example.txt
// remotePath: /user/app
// remoteFilePath: /user/app/example.txt
func (c *Client) PutFile(localFilePath string, remotePath string) {
	fileData, err := os.ReadFile(localFilePath)
	if err != nil {
		log.Fatalf("not found localfile: %s", localFilePath)
	}
	// 400 * 1024 => 400KB
	chunkSize := c.conf.Client.NameNode.ChunkSize * 1024
	var chunkNum int64
	if int64(len(fileData))%chunkSize != 0 {
		chunkNum = int64(len(fileData))/chunkSize + 1
	} else {
		chunkNum = int64(len(fileData)) / chunkSize
	}
	fileName, _ := common.SplitClientFileNamePath(localFilePath)
	remoteFilePath := path.Join(remotePath, fileName)
	err = c.doWrite(remoteFilePath, int64(len(fileData)), fileData, chunkSize, chunkNum)
	if err != nil {
		fmt.Printf("doWrite file error: %v\n", err)
	}
}

func (c *Client) doWrite(remoteFilePath string, fileTotalSize int64, fileData []byte, chunkSize int64, chunkNum int64) error {
	fileOperationArg := &proto.FileOperationArg{
		Operation:  proto.FileOperationArg_WRITE,
		FileName:   remoteFilePath, // /user/app/example.txt
		ChunkNum:   chunkNum,
		FileSize:   fileTotalSize,
		ReplicaNum: c.conf.Client.NameNode.ChunkReplicaNum,
	}
	replicaNum := fileOperationArg.ReplicaNum
	nameServiceClient := getNameNodeConnection(c.conf.Client.NameNode.Host)
	// client 向 nameNode 发送文件信息, 来获得datanode信息;
	dataServerChain, err := nameServiceClient.PutFile(context.Background(), fileOperationArg)
	if err != nil {
		return err
	}
	address := dataServerChain.DataNodeAddress

	fmt.Printf("remotePath: %s; len:%d; dataServerChain:%s;\n", remoteFilePath, len(address), address)
	if int32(len(address)) < replicaNum {
		fmt.Printf("DataNodeAddress.lengh:%d ; dataServerChain:%s ; replicaNum:%d \n; ",
			len(address), address, replicaNum)
		return common.ErrEnoughReplicaDataNodeServer
	}

	firstNode := address[0]
	nextNode := address[1:]
	//primaryNode := address[len(address)-1]

	fileLocationInfo := &proto.FileLocationInfo{} // 保存单个文件的 每个块信息以及所在的dataNode信息;
	// FilePathChunkName: /user/app/example.txt_chunk_0 1 2
	chunkNames := common.GetFileChunkNameOfNum(remoteFilePath, int(chunkNum))
	for i := 0; i < int(chunkNum); i++ {
		chunkInfo := &proto.ChunkInfo{}             // 单个文件块信息;
		chunkInfo.FilePathName = remoteFilePath     // /user/app/example.txt
		chunkInfo.FilePathChunkName = chunkNames[i] // /user/app/example.txt_chunk_0
		chunkInfo.ChunkId = int32(i)                // 0 1 2

		fileDataStream := &proto.FileDataStream{
			DataNodeChain:     nextNode, // 每个文件块的下个dataNode节点;
			FilePathName:      remoteFilePath,
			FilePathChunkName: chunkNames[i],
			ChunkId:           chunkInfo.ChunkId,
			SrcName:           c.name, // 发源地点;
			Operation:         proto.ChunkReplicateStatus_NormalToReplicate,
		}
		fileLocationInfo.Chunks = append(fileLocationInfo.Chunks, chunkInfo)
		if i == int(chunkNum)-1 {
			fileDataStream.Data = fileData[i*int(chunkSize):]
		} else {
			fileDataStream.Data = fileData[i*int(chunkSize) : (i+1)*int(chunkSize)]
		}
		// 把文件的每一块,按序向首个dataNode发送; 当dataNode收到后,其会向nextNode发送数据;
		err = c.writeToDataNode(firstNode, fileDataStream)
		if err != nil {
			// 重试都出现错误, 返回错误;
			return err
		}
	}

	arg := &proto.ConfirmFileArg{
		FileName:         remoteFilePath,
		FileLocationInfo: fileLocationInfo,
		Ack:              false, // 是否检测全部文件 chunk 数据;
	}

	// 客户端确认可能需要时间, 因为 dataNode-1 存储文件块以及转发文件,以及提交, 需要时间;
	confirmFileReply := c.ConfirmFile(arg)

	if !confirmFileReply.GetSuccess() {
		// todo 根据判断结果,client是否需要重试发送
		fmt.Printf("client confirmFile:%s failed. context: %s \n", remoteFilePath, confirmFileReply.Context)
	}
	return nil
}

func (c *Client) writeToDataNode(dataNodeAddress string, file *proto.FileDataStream) error {
	dataServiceClient := getDataNodeConnection(dataNodeAddress)
	putChunkClient, err := dataServiceClient.PutChunk(context.Background())
	if err != nil {
		return err
	}
	err = putChunkClient.Send(file)
	if err != nil {
		// todo 需要重试发送
		fmt.Printf("putChunkClient send chunk error: %v", err)
		return err
	}

	// dataName 返回的是空值;
	_, err = putChunkClient.CloseAndRecv()

	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	return nil
}

func (c *Client) ConfirmFile(arg *proto.ConfirmFileArg) *proto.ConfirmFileReply {
	nameServiceClient := getNameNodeConnection(c.conf.Client.NameNode.Host)
	confirmFileReply, err := nameServiceClient.ConfirmFile(context.Background(), arg)
	if err != nil {
		fmt.Printf("client faile to ConfirmFile:%s ack:%v, err:%s; \n", arg.FileName, arg.Ack, err)
	}
	return confirmFileReply
}

func (c *Client) GetFile(localPath string, remoteFilePath string) (*os.File, error) {
	fileOperationArg := &proto.FileOperationArg{
		Operation: proto.FileOperationArg_READ,
		FileName:  remoteFilePath,
	}
	nameServiceClient := getNameNodeConnection(c.conf.Client.NameNode.Host)

	fileLocationInfo, err := nameServiceClient.GetFile(context.Background(), fileOperationArg)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 0)
	chunkNames := common.GetFileChunkNameOfNum(remoteFilePath, int(fileLocationInfo.ChunkNum))
	sortChunkNames := make([]*proto.ChunkInfo, 0)

	// sortChunkName: chunkName1, chunkName2, chunkName3;
	for _, chunkName := range chunkNames {
		for _, chunkInfo := range fileLocationInfo.Chunks {
			if chunkInfo.FilePathChunkName == chunkName {
				sortChunkNames = append(sortChunkNames, chunkInfo)
				break
			}
		}
	}

	// getChunkData
	for _, chunkInfo := range sortChunkNames { // chunkNameInfo1, chunkNameInfo2, chunkNameInfo3
		dataNodeAddress := chunkInfo.DataNodeAddress.DataNodeAddress
		for _, nodeAddress := range dataNodeAddress {
			toDataServiceClient := getDataNodeConnection(nodeAddress)
			chunkClient, err := toDataServiceClient.GetChunk(context.Background(),
				&proto.FileOperationArg{FileName: chunkInfo.FilePathChunkName})
			if err != nil {
				fmt.Printf("client faile to GetChunk:%s, err:%s; \n", chunkInfo.FilePathChunkName, err)
				continue
			}
			fileDataStream, err := chunkClient.Recv()
			if err != nil {
				fmt.Printf("client faile to chunkClient.Recv():%s,from:%s, err:%s; \n",
					chunkInfo.FilePathChunkName, nodeAddress, err)
				// 同一个 chunkName, 向下一个 dataNode-1 请求数据;
				continue
			} else {
				fmt.Printf("client chunkClient.Recv():%s,from:%s,success; \n",
					chunkInfo.FilePathChunkName, nodeAddress)
				buf = append(buf, fileDataStream.Data...)
				break
			}
		}
	}

	// valid data size
	if int64(len(buf)) != fileLocationInfo.FileSize {
		fmt.Printf("client faile to GetFile:%s;len(buf):%d;size:%d; err:%s; \n", remoteFilePath, len(buf), fileLocationInfo.FileSize, err)
		return nil, common.ErrFileCorruption
	}

	// name : yy.png
	name, _ := common.SplitFileNamePath(remoteFilePath)
	localFilePath := filepath.Join(localPath, name)

	// openFile
	openFile, err := os.OpenFile(localFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	} else {
		if _, err = openFile.Write(buf); err != nil {
			return nil, err
		} else {
			return openFile, nil
		}
	}
}

func (c *Client) DeleteFile(remoteFilePath string) error {
	fileOperationArg := &proto.FileOperationArg{
		Operation: proto.FileOperationArg_DELETE,
		FileName:  remoteFilePath,
	}
	nameServiceClient := getNameNodeConnection(c.conf.Client.NameNode.Host)
	_, err := nameServiceClient.DeleteFile(context.Background(), fileOperationArg)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) ListDir(remotePath string) (*proto.DirMetaList, error) {
	fileOperationArg := &proto.FileOperationArg{
		Operation: proto.FileOperationArg_LISTDIR,
		FileName:  remotePath,
	}
	nameServiceClient := getNameNodeConnection(c.conf.Client.NameNode.Host)
	metaList, err := nameServiceClient.ListDir(context.Background(), fileOperationArg)
	if err != nil {
		return nil, err
	}
	return metaList, nil
}

func (c *Client) Mkdir(remotePath string) error {
	fileOperationArg := &proto.FileOperationArg{
		Operation: proto.FileOperationArg_MKDIR,
		FileName:  remotePath,
	}
	nameServiceClient := getNameNodeConnection(c.conf.Client.NameNode.Host)
	_, err := nameServiceClient.Mkdir(context.Background(), fileOperationArg)
	if err != nil {
		return err
	}
	return nil
}

// ReName 目前仅支持,更改目录为空的目录名字, 不能更改文件名字和中间目录名字;
func (c *Client) ReName(oldPath string, newPath string) (*proto.ReNameReply, error) {
	fileOperationArg := &proto.FileOperationArg{
		Operation:   proto.FileOperationArg_RENAME,
		FileName:    oldPath,
		NewFileName: newPath,
	}
	nameServiceClient := getNameNodeConnection(c.conf.Client.NameNode.Host)
	nameReply, err := nameServiceClient.ReName(context.Background(), fileOperationArg)
	if err != nil {
		return nil, err
	}
	return nameReply, nil
}

func getNameNodeConnection(nameNodeAddress string) proto.ClientToNameServiceClient {
	conn, err := grpc.DialContext(context.Background(), nameNodeAddress, grpc.WithInsecure())
	if err != nil {
		log.Printf("Did not connect to nameNodeAddress %v error %v ;\n", nameNodeAddress, err)
	}
	client := proto.NewClientToNameServiceClient(conn)
	return client
}

func getDataNodeConnection(dataNodeAddress string) proto.ClientToDataServiceClient {
	conn, err := grpc.DialContext(context.Background(), dataNodeAddress, grpc.WithInsecure())
	if err != nil {
		log.Printf("Did not connect to dataNodeAddress %v error %v ;\n", dataNodeAddress, err)
	}
	client := proto.NewClientToDataServiceClient(conn)
	return client
}

// Deprecated
func (c *Client) getFileLocation(arg *proto.FileOperationArg) *proto.FileLocationInfo {
	nameServiceClient := getNameNodeConnection(c.conf.Client.NameNode.Host)
	fileLocation, err := nameServiceClient.GetFileLocation(context.Background(), arg)
	if err != nil {
		log.Fatalf("fail to getFileLocation %v \n", err)
	}
	return fileLocation
}

// Deprecated
func (c *Client) getFileStoreChain(arg *proto.FileOperationArg) *proto.DataNodeChain {
	nameServiceClient := getNameNodeConnection(c.conf.Client.NameNode.Host)
	fileLocation, err := nameServiceClient.GetFileStoreChain(context.Background(), arg)
	if err != nil {
		log.Fatalf("fail to getFileLocation %v \n", err)
	}
	return fileLocation
}
