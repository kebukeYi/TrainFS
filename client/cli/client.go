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

func (c *Client) PutFile(localPath string, remotePath string) {
	fileData, err := os.ReadFile(localPath)
	if err != nil {
		log.Fatalf("not found localfile %s", localPath)
	}
	// 400 * 1024 => 400KB
	chunkSize := c.conf.Client.NameNode.ChunkSize * 1024
	var chunkNum int64
	if int64(len(fileData))%chunkSize != 0 {
		chunkNum = int64(len(fileData))/chunkSize + 1
	} else {
		chunkNum = int64(len(fileData)) / chunkSize
	}
	err = c.doWrite(remotePath, int64(len(fileData)), fileData, chunkSize, chunkNum)
	if err != nil {
		fmt.Printf("doWrite file error: %v", err)
	}
}

func (c *Client) doWrite(remotePath string, fileTotalSize int64, fileData []byte, chunkSize int64, chunkNum int64) error {
	fileOperationArg := &proto.FileOperationArg{
		Operation:  proto.FileOperationArg_WRITE,
		FileName:   remotePath, // /user/app/example.txt
		ChunkNum:   chunkNum,
		FileSize:   fileTotalSize,
		ReplicaNum: c.conf.Client.NameNode.ChunkReplicaNum,
	}
	replicaNum := c.conf.Client.NameNode.ChunkReplicaNum
	nameServiceClient := getNameNodeConnection(c.conf.Client.NameNode.Host)
	dataServerChain, err := nameServiceClient.PutFile(context.Background(), fileOperationArg)
	if err != nil {
		return err
	}
	address := dataServerChain.DataNodeAddress
	fmt.Printf("remotePath: %s; len:%d; dataServerChain:%s;\n", remotePath, len(address), address)
	if int32(len(address)) < replicaNum {
		fmt.Printf("len:%d ; dataServerChain:%s ; replicaNum:%d \n; ", len(address), address, replicaNum)
		return common.ErrEnoughReplicaDataNodeServer
	}
	firstNode := address[0]
	lastNode := address[1:]
	//primaryNode := address[len(address)-1]

	fileLocationInfo := &proto.FileLocationInfo{}
	// FilePathChunkName: /user/app/example.txt_chunk_3
	chunkNames := common.GetFileChunkNameOfNum(remotePath, int(chunkNum))
	for i := 0; i < int(chunkNum); i++ {
		var fileDataStream *proto.FileDataStream
		chunkInfo := &proto.ChunkInfo{}
		chunkInfo.FilePathName = remotePath
		chunkInfo.FilePathChunkName = chunkNames[i]
		chunkInfo.ChunkId = int32(i)
		fileLocationInfo.Chunks = append(fileLocationInfo.Chunks, chunkInfo)
		if i == int(chunkNum)-1 {
			fileDataStream = &proto.FileDataStream{
				Data:              fileData[i*int(chunkSize):],
				DataNodeChain:     lastNode,
				FilePathChunkName: chunkNames[i],
				FilePathName:      remotePath,
				ChunkId:           chunkInfo.ChunkId,
				SrcName:           c.name,
				Operation:         proto.ChunkReplicateStatus_NormalToReplicate,
			}
		} else {
			fileDataStream = &proto.FileDataStream{
				Data:              fileData[i*int(chunkSize) : (i+1)*int(chunkSize)],
				DataNodeChain:     lastNode,
				FilePathChunkName: chunkNames[i],
				FilePathName:      remotePath,
				ChunkId:           chunkInfo.ChunkId,
				SrcName:           c.name,
				Operation:         proto.ChunkReplicateStatus_NormalToReplicate,
			}
		}
		err := c.writeToDataNode(firstNode, fileDataStream)
		if err != nil {
			// todo 需要重试发送
			return err
		}
	}

	arg := &proto.ConfirmFileArg{
		FileName:         remotePath,
		FileLocationInfo: fileLocationInfo,
		Ack:              false,
	}

	confirmFileReply := c.ConfirmFile(arg)

	if !confirmFileReply.GetSuccess() {
		// todo 需要重试发送
		fmt.Printf("client confirmFile:%s failed. context: %s \n", remotePath, confirmFileReply.Context)
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
		fmt.Printf("putChunkClient send chunk error: %v", err)
		// todo 需要重试
		return err
	}

	_, err = putChunkClient.CloseAndRecv()

	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}
	return nil
}

func (c *Client) getFileLocation(arg *proto.FileOperationArg) *proto.FileLocationInfo {
	nameServiceClient := getNameNodeConnection(c.conf.Client.NameNode.Host)
	fileLocation, err := nameServiceClient.GetFileLocation(context.Background(), arg)
	if err != nil {
		log.Fatalf("fail to getFileLocation %v \n", err)
	}
	return fileLocation
}

func (c *Client) getFileStoreChain(arg *proto.FileOperationArg) *proto.DataNodeChain {
	nameServiceClient := getNameNodeConnection(c.conf.Client.NameNode.Host)
	fileLocation, err := nameServiceClient.GetFileStoreChain(context.Background(), arg)
	if err != nil {
		log.Fatalf("fail to getFileLocation %v \n", err)
	}
	return fileLocation
}

func (c *Client) ConfirmFile(arg *proto.ConfirmFileArg) *proto.ConfirmFileReply {
	nameServiceClient := getNameNodeConnection(c.conf.Client.NameNode.Host)
	confirmFileReply, err := nameServiceClient.ConfirmFile(context.Background(), arg)
	if err != nil {
		fmt.Printf("client faile to ConfirmFile:%s ack:%v, err:%s; \n", arg.FileName, arg.Ack, err)
	}
	return confirmFileReply
}

func (c *Client) GetFile(localPath string, remotePath string) (*os.File, error) {
	fileOperationArg := &proto.FileOperationArg{
		Operation: proto.FileOperationArg_READ,
		FileName:  remotePath,
	}
	nameServiceClient := getNameNodeConnection(c.conf.Client.NameNode.Host)

	fileLocationInfo, err := nameServiceClient.GetFile(context.Background(), fileOperationArg)
	if err != nil {
		return nil, err
	}
	buf := make([]byte, 0)
	chunkNames := common.GetFileChunkNameOfNum(remotePath, int(fileLocationInfo.ChunkNum))
	sortChunkNames := make([]*proto.ChunkInfo, 0)
	for _, chunkName := range chunkNames {
		for _, chunkInfo := range fileLocationInfo.Chunks {
			if chunkInfo.FilePathChunkName == chunkName {
				sortChunkNames = append(sortChunkNames, chunkInfo)
			}
		}
	}

	for _, chunkInfo := range sortChunkNames {
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
				fmt.Printf("client faile to chunkClient.Recv():%s, err:%s; \n", chunkInfo.FilePathChunkName, err)
				continue
			} else {
				buf = append(buf, fileDataStream.Data...)
				break
			}
		}
	}

	if int64(len(buf)) != fileLocationInfo.FileSize {
		fmt.Printf("client faile to GetFile:%s;len(buf):%d;size:%d; err:%s; \n", remotePath, len(buf), fileLocationInfo.FileSize, err)
		return nil, common.ErrFileCorruption
	}
	openFile, err := os.OpenFile(localPath, os.O_CREATE|os.O_WRONLY, 0644)
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

func (c *Client) DeleteFile(remotePath string) error {
	fileOperationArg := &proto.FileOperationArg{
		Operation: proto.FileOperationArg_DELETE,
		FileName:  remotePath,
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
