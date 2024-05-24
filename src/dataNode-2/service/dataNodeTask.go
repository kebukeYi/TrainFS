package service

import (
	"context"
	"fmt"
	"github.com/shirou/gopsutil/v3/disk"
	"log"
	"time"
	proto "trainfs/src/profile"
)

func (dataNode *DataNode) Register() (bool, error) {
	nameServiceClient, err := dataNode.getGrpcNameNodeServerConn(dataNode.Config.NameNodeHost)
	if err != nil {
		fmt.Printf("DataNode[%s] getGrpcNameNodeServerConn to register failed...err:%v \n", dataNode.Config.Host, err)
		return false, err
	}
	freeSpace, err := disk.Usage(dataNode.Config.DataDir)
	reply, err := nameServiceClient.RegisterDataNode(context.Background(), &proto.DataNodeRegisterArg{
		DataNodeAddress: dataNode.Config.Host,
		FreeSpace:       freeSpace.Free,
	})
	if err != nil {
		fmt.Printf("DataNode[%s] send register failed... err:%v \n", dataNode.Config.Host, err)
		return false, err
	}
	if reply.GetSuccess() {
		fmt.Println("DataNode send " + dataNode.Config.Host + " register success...")
		return true, nil
	} else {
		fmt.Printf("DataNode register failed, err context:%s \n", reply.Context)
		return false, nil
	}
}

func (dataNode *DataNode) ChunkReportTask() {
	nameServiceClient, err := dataNode.getGrpcNameNodeServerConn(dataNode.Config.NameNodeHost)
	if err != nil {
		return
	}
	chunkInfos := make([]*proto.ChunkInfo, 0)
	for _, chunkInfo := range dataNode.allChunkInfos {
		chunkInfos = append(chunkInfos, chunkInfo)
	}
	if len(chunkInfos) <= 0 {
		return
	}
	reply, err := nameServiceClient.ChunkReport(context.Background(),
		&proto.FileLocationInfo{Chunks: chunkInfos, DataNodeAddress: dataNode.Config.Host})
	if err != nil {
		return
	}
	if reply.GetSuccess() {
		fmt.Printf("DataNode report chunk success, chunkNum:%d\n", len(chunkInfos))
	} else {
		fmt.Println(err)
	}
}

func (dataNode *DataNode) HeartBeatTask() {
	interval := dataNode.Config.HeartbeatInterval
	for {
		time.Sleep(time.Duration(interval) * time.Millisecond)
		nameServiceClient, err := dataNode.getGrpcNameNodeServerConn(dataNode.Config.NameNodeHost)
		if err != nil {
			// todo dataNode心跳连接不上,需要一直重试
			fmt.Printf("DataNode get heartbeat con failed, err context:%s \n", err)
			continue
		}
		heartBeatReply, err := nameServiceClient.HeartBeat(context.Background(), &proto.HeartBeatArg{DataNodeAddress: dataNode.Config.Host})
		if err != nil {
			// todo dataNode心跳连接不上,需要一直重试
			// 1.nameNode宕机 2.网络原因
			fmt.Printf("DataNode HeartBeat() failed, err context:%s \n", err)
			continue
		}
		// fmt.Printf("DataNode[%s] send headrbeat;\n", dataNode.Config.Host)
		if heartBeatReply != nil {
			go func() {
				if len(heartBeatReply.TrashFilePathChunkNames) > 0 {
					fmt.Printf("DataNode rev headrbeat delete %v;\n", heartBeatReply.TrashFilePathChunkNames)
					dataNode.TrashChan <- heartBeatReply.TrashFilePathChunkNames
				}
			}()
			go func() {
				if len(heartBeatReply.NewChunkSevers) > 0 {
					fmt.Printf("DataNode rev headrbeat replicat %v;\n", heartBeatReply.NewChunkSevers)
					result := make([]*Replication, 0)
					for i := 0; i < len(heartBeatReply.NewChunkSevers); i++ {
						result = append(result, &Replication{
							filePathChunkName: heartBeatReply.FilePathChunkNames[i],
							toAddress:         heartBeatReply.NewChunkSevers[i],
							filePathName:      heartBeatReply.FilePathNames[i],
						})
					}
					dataNode.ReplicaChain <- result
				}
			}()
		}
	}
}

func (dataNode *DataNode) DoTrashTask() {
	for {
		select {
		case fileChunkNames := <-dataNode.TrashChan:
			dataNode.Trash(fileChunkNames)
		}
	}
}

func (dataNode *DataNode) Trash(fileChunkNames []string) {
	dataNode.mux.Lock()
	defer dataNode.mux.Unlock()
	dataNode.taskStoreManger.PutTrashs(TRASHTASK_KEY, fileChunkNames)
	for _, fileChunkName := range fileChunkNames {
		delete(dataNode.allChunkInfos, fileChunkName)
		err := dataNode.dataStoreManger.Delete(fileChunkName)
		if err != nil {
			log.Printf("dataNode-1.StoreManger.Delete(%s) fail; err:%v \n", fileChunkName, err)
		} else {
			err = dataNode.taskStoreManger.PutTrashs(TRASHTASK_KEY, fileChunkNames[1:])
			if err != nil {
				log.Printf("dataNode-1.StoreManger.PutTrashs(%v) fail; err:%v \n", fileChunkNames[1:], err)
			}
		}
	}
}

func (dataNode *DataNode) DoReplicaTask() {
	for {
		select {
		case replications := <-dataNode.ReplicaChain:
			dataNode.Replica(replications)
		}
	}
}

func (dataNode *DataNode) Replica(replications []*Replication) {
	dataNode.mux.Lock()
	defer dataNode.mux.Unlock()
	// todo DataNode 执行本地持久化
	dataNode.metaStoreManger.PutReplications(REPLICAT_KEY, replications)
	for _, replication := range replications {
		dataServiceClient, err := dataNode.getGrpcDataServerConn(replication.toAddress)
		if err != nil {
			log.Printf("dataNode.getGrpcDataServerConn(%s) fail; err:%v \n", replication.toAddress, err)
		}
		putChunkClient, err := dataServiceClient.PutChunk(context.Background())
		if err != nil {
			log.Printf("dataServiceClient.PutChunk(context.Background()) fail; err:%v \n", err)
		}
		bytes, err := dataNode.dataStoreManger.Get(replication.filePathChunkName)
		if err != nil {
			log.Printf("dataNode.dataStoreManger.Get(%s) fail; err:%v \n", replication.filePathChunkName, err)
		}
		err = putChunkClient.Send(&proto.FileDataStream{
			Data:              bytes,
			FilePathName:      replication.filePathName,
			FilePathChunkName: replication.filePathChunkName,
			Address:           dataNode.Config.Host,
			SrcName:           dataNode.name,
			Operation:         proto.ChunkReplicateStatus_LostToReplicate,
			DataNodeChain:     nil,
		})
		if err != nil {
			log.Printf("putChunkClient.Send({Data, FilePathName: %s, FilePathChunkName: %s; err: %s \n",
				replication.filePathName, replication.filePathName, err)
		} else {
			err = dataNode.metaStoreManger.PutReplications(REPLICAT_KEY, replications[1:])
			if err != nil {
				log.Printf("dataNode.dataStoreManger.PutReplications(%v) fail; err:%v \n", replications[1:], err)
			}
		}
	}
}
