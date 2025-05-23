package service

import (
	"context"
	"fmt"
	"github.com/kebukeYi/TrainFS/common"
	proto "github.com/kebukeYi/TrainFS/profile"
	"github.com/shirou/gopsutil/v3/disk"
	"time"
)

func (dataNode *DataNode) Register() (bool, error) {
	for {
	GETCONN:
		callBack, nameServiceClient, err := dataNode.getGrpcNameNodeServerConn(dataNode.Config.NameNodeHost)
		if err != nil {
			fmt.Printf("DataNode[%s]-%s getGrpcNameNodeServerConn from %s to register failed... err:%v \n",
				dataNode.Config.Host, dataNode.Config.DataNodeId, dataNode.Config.NameNodeHost, err)
			time.Sleep(3 * time.Second)
			goto GETCONN
		}

		freeSpace, err := disk.Usage(dataNode.Config.DataDir)
		if err != nil {
			fmt.Printf("DataNode[%s]-%s get freeSpace failed... err:%v \n",
				dataNode.Config.Host, dataNode.Config.DataNodeId, err)
			return false, err
		}

	REGISTER:
		reply, err := nameServiceClient.RegisterDataNode(context.Background(), &proto.DataNodeRegisterArg{
			DataNodeAddress: dataNode.Config.Host,
			FreeSpace:       freeSpace.Free,
		})
		if err != nil {
			fmt.Printf("DataNode[%s]-%s send register failed... err:%v \n", dataNode.Config.Host, dataNode.Config.DataNodeId, err)
			time.Sleep(2 * time.Second)
			goto REGISTER
		}
		if reply.GetSuccess() {
			fmt.Printf("DataNode[%s]-%s send register success... err:%v \n", dataNode.Config.Host, dataNode.Config.DataNodeId, err)
			go dataNode.ChunkReportTask()
			callBack()
			break
		} else {
			fmt.Printf("DataNode[%s]-%s rev register fail reply:%s... err:%v \n",
				dataNode.Config.Host, dataNode.Config.DataNodeId, reply.Context, err)
			callBack()
			return false, err
		}
	}
	return true, nil
}

// ChunkReportTask 上报自身存储chunk信息任务;
func (dataNode *DataNode) ChunkReportTask() {
	callBack, nameServiceClient, err := dataNode.getGrpcNameNodeServerConn(dataNode.Config.NameNodeHost)
	defer callBack()
	if err != nil {
		fmt.Printf("DataNode[%s]-%s getGrpcNameNodeServerConn fail...err:%v \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId, err)
		return
	}
	chunkInfos := make([]*proto.ChunkInfo, 0)
	for _, chunkInfo := range dataNode.allChunkInfos {
		chunkInfos = append(chunkInfos, chunkInfo)
	}
	if len(chunkInfos) <= 0 {
		fmt.Printf("DataNode[%s]-%s no chunkInfos to report... \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId)
	}
	reply, err := nameServiceClient.ChunkReport(context.Background(),
		&proto.FileLocationInfo{Chunks: chunkInfos, DataNodeAddress: dataNode.Config.Host})
	fmt.Printf("DataNode[%s]-%s report allChunkInfos:%v; \n",
		dataNode.Config.Host, dataNode.Config.DataNodeId, chunkInfos)
	if err != nil {
		fmt.Printf("DataNode[%s]-%s report chunkInfos failed; err:%v \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId, err)
		return
	}
	if reply.GetSuccess() {
		fmt.Printf("DataNode[%s]-%s report chunkInfos success chunkInfosSum:%d; \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId, len(chunkInfos))
		go dataNode.HeartBeatTask() // 开启心跳任务;
		go dataNode.DoTrashTask()   // 开始消费 删除任务;
		go dataNode.DoReplicaTask() // 开始消费 复制转发任务;
	} else {
		fmt.Printf("DataNode[%s]-%s report chunkInfos rev fail context%s; \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId, reply.Context)
	}
}

func (dataNode *DataNode) HeartBeatTask() {
	retry := 0
	interval := dataNode.Config.HeartbeatInterval
	callBack, nameServiceClient, err := dataNode.getGrpcNameNodeServerConn(dataNode.Config.NameNodeHost)
	defer callBack()
	if err != nil {
		fmt.Printf("DataNode[%s]-%s getGrpcNameNodeServerConn to heart fail...err:%v \n", dataNode.Config.Host, dataNode.Config.DataNodeId, err)
		return
	}
	for {
		time.Sleep(time.Duration(interval) * time.Millisecond)
		heartBeatReply, err := nameServiceClient.HeartBeat(context.Background(),
			&proto.HeartBeatArg{DataNodeAddress: dataNode.Config.Host})
		if err != nil {
			retry++
			if retry > dataNode.Config.HeartBeatRetry {
				// 重新走 注册逻辑;
				go dataNode.LiveDetectionTask(dataNode.Config.NameNodeHost)
				return
			}
			// todo toDataNode的心跳连接不上,需要一直重试;达到重试次数后,重新走注册逻辑;
			// 1.nameNode宕机 2.网络原因
			fmt.Printf("DataNode[%s]-%s HeartBeat fail to retry: %d ... err:%v \n", dataNode.Config.Host, dataNode.Config.DataNodeId, retry, err)
			continue
		}
		retry = 0
		// todo 正常心跳 调试打印
		//fmt.Printf("DataNode[%s] send headrbeat;\n", dataNode.Config.Host)
		if heartBeatReply != nil {
			go func() {
				if len(heartBeatReply.TrashFilePathChunkNames) > 0 {
					fmt.Printf("DataNode[%s]-%s rev headrbeatReply delete:%v \n",
						dataNode.Config.Host, dataNode.Config.DataNodeId, heartBeatReply.TrashFilePathChunkNames)
					dataNode.TrashChan <- heartBeatReply.TrashFilePathChunkNames
				}
			}()
			go func() {
				if len(heartBeatReply.NewChunkSevers) > 0 {
					fmt.Printf("DataNode[%s]-%s rev headrbeatReply replicat:%v to:%v \n",
						dataNode.Config.Host, dataNode.Config.DataNodeId, heartBeatReply.FilePathChunkNames, heartBeatReply.NewChunkSevers)
					result := make([]*Replication, 0)
					for i := 0; i < len(heartBeatReply.NewChunkSevers); i++ {
						result = append(result, &Replication{
							FilePathName:      heartBeatReply.FilePathNames[i],
							FilePathChunkName: heartBeatReply.FilePathChunkNames[i],
							ToAddress:         heartBeatReply.NewChunkSevers[i],
						})
					}
					dataNode.ReplicaChain <- result
				}
			}()
		}
	}
}

func (dataNode *DataNode) LiveDetectionTask(address string) {
	interval := dataNode.Config.HeartbeatInterval

	callBack, nameServiceClient, err := dataNode.getGrpcNameNodeServerConn(address)
	defer callBack()
	if err != nil {
		fmt.Printf("DataNode[%s]-%s getGrpcNameNodeServerConn() con failed. err:%v \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId, err)
		return
	}
	for {
		detectionReply, err := nameServiceClient.LiveDetection(context.Background(), &proto.LiveDetectionArg{DataNodeAddress: dataNode.Config.Host})
		if err != nil {
			fmt.Printf("DataNode[%s]-%s send address:%s liveDetection rpc failed. err:%s \n",
				dataNode.Config.Host, dataNode.Config.DataNodeId, address, err)
			time.Sleep(time.Duration(interval) * time.Millisecond)
			continue
		}
		if detectionReply.Success {
			go dataNode.Register()
			return
		}

	}
}

func (dataNode *DataNode) DoTrashTask() {
	for {
		select {
		// 每次心跳,nameNode会下发删除任务;
		case fileChunkNames := <-dataNode.TrashChan:
			dataNode.Trash(fileChunkNames)
		}
	}
}

func (dataNode *DataNode) Trash(fileChunkNames []string) {
	dataNode.mux.Lock()
	defer dataNode.mux.Unlock()
	err := dataNode.taskStoreManger.PutTrashes(trashKey, fileChunkNames)
	if err != nil {
		fmt.Printf("DataNode[%s]-%s taskStoreManger.PutTrashes(%s) fail. err:%s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId, trashKey, err)
		return
	}
	fileChunkNameSize := len(fileChunkNames)
	fmt.Printf("DataNode[%s]-%s rev #Trash fileChunkNames:%v \n",
		dataNode.Config.Host, dataNode.Config.DataNodeId, fileChunkNames)

	for i, fileChunkName := range fileChunkNames {
		if _, ok := dataNode.allChunkInfos[fileChunkName]; !ok {
			fmt.Printf("DataNode[%s]-%s Trash fileChunkName:%s not exist! \n", dataNode.Config.Host, dataNode.Config.DataNodeId, fileChunkName)
			continue
		}
		fmt.Printf("DataNode[%s]-%s for delete the fileChunkName:%v \n", dataNode.Config.Host, dataNode.Config.DataNodeId, fileChunkName)
		tempChunkInfo := dataNode.allChunkInfos[fileChunkName]
		delete(dataNode.allChunkInfos, fileChunkName)
		err := dataNode.dataStoreManger.Delete(fileChunkName)
		if err != nil {
			fmt.Printf("DataNode[%s]-%s dataStoreManger.Delete(%s) fail. err:%v \n",
				dataNode.Config.Host, dataNode.Config.DataNodeId, fileChunkName, err)
			return
		} else {
			_, err = dataNode.CommitChunk(&proto.CommitChunkArg{
				FileChunkName:   fileChunkName,
				FilePathName:    tempChunkInfo.FilePathName,
				FileSize:        tempChunkInfo.ChunkSize,
				Operation:       proto.ChunkReplicateStatus_DeleteFileChunk,
				ChunkId:         tempChunkInfo.ChunkId,
				SrcAddress:      dataNode.Config.Host,
				DataNodeAddress: []string{dataNode.Config.Host},
			}, 4)
			if err != nil {
				// todo 应该重试，重试失败后,应该保存到本地;
				fmt.Printf("DataNode[%s]-%s CommitChunk(%s) type:%s; fail. err:%s \n",
					dataNode.Config.Host, dataNode.Config.DataNodeId, fileChunkName,
					proto.ChunkReplicateStatus_DeleteFileChunk, err)
				continue
			} else {
				fmt.Printf("DataNode[%s]-%s CommitChunk(%s) type:%s; success.\n",
					dataNode.Config.Host, dataNode.Config.DataNodeId, fileChunkName,
					proto.ChunkReplicateStatus_DeleteFileChunk)
			}
			if i+1 >= fileChunkNameSize {
				break
			} else {
				err = dataNode.taskStoreManger.PutTrashes(trashKey, fileChunkNames[i+1:])
				fmt.Printf("DataNode[%s]-%s taskStoreManger.PutTrashes(%s) success.\n",
					dataNode.Config.Host, dataNode.Config.DataNodeId,
					fileChunkNames[i+1:])
			}
		}
	}

	err = dataNode.taskStoreManger.Delete(trashKey)
	if err != nil {
		fmt.Printf("DataNode[%s]-%s taskStoreManger.Delete(%s) fail. err:%v \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId,
			trashKey, err)
		return
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
	err := dataNode.taskStoreManger.PutReplications(replicationKey, replications)
	if err != nil {
		fmt.Printf("DataNode[%s]-%s taskStoreManger.PutReplications(%s) fail. err:%s \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId, replicationKey, err)
		return
	}
	replicationSize := len(replications)
	for i, replication := range replications {
		callBack, dataServiceClient, err := dataNode.getGrpcDataServerConn(replication.ToAddress)
		if err != nil {
			fmt.Printf("DataNode[%s]-%s getGrpcDataServerConn(%s) to Replica fail; err:%v \n",
				dataNode.Config.Host, dataNode.Config.DataNodeId, replication.ToAddress, err)
			callBack()
			continue
		}
		for {
			putChunkClient, err := dataServiceClient.PutChunk(context.Background())
			if err != nil {
				fmt.Printf("DataNode[%s]-%s dataServiceClient.PutChunk() to Replica fail; err:%v \n",
					dataNode.Config.Host, dataNode.Config.DataNodeId, err)
			}

			bytes, err := dataNode.dataStoreManger.Get(replication.FilePathChunkName)
			if err != nil {
				fmt.Printf("DataNode[%s]-%s dataStoreManger.Get(%s) fail. err:%v \n",
					dataNode.Config.Host, dataNode.Config.DataNodeId,
					replication.FilePathChunkName, err)
				return
			}

			err = putChunkClient.Send(&proto.FileDataStream{
				Data:              bytes,
				FilePathName:      replication.FilePathName,
				FilePathChunkName: replication.FilePathChunkName,
				ChunkId:           common.GetChunkIdOfFileChunkName(replication.FilePathChunkName),
				Address:           dataNode.Config.Host,
				SrcName:           dataNode.name,
				Operation:         proto.ChunkReplicateStatus_LostToReplicate,
				DataNodeChain:     nil,
			})
			fmt.Printf("DataNode[%s]-%s putChunkClient.Send(DataLen:%d, FilePathName:%s, FilePathChunkName: %s; type:%s; to:%s;\n",
				dataNode.Config.Host, dataNode.Config.DataNodeId,
				len(bytes),
				replication.FilePathName, replication.FilePathChunkName,
				proto.ChunkReplicateStatus_LostToReplicate,
				replication.ToAddress)
			if err != nil {
				fmt.Printf("DataNode[%s]-%s putChunkClient.Send({Data, FilePathName: %s, FilePathChunkName: %s; type:%s; to:%s; fail. err: %s \n",
					dataNode.Config.Host, dataNode.Config.DataNodeId,
					replication.FilePathName, replication.FilePathName,
					proto.ChunkReplicateStatus_LostToReplicate,
					replication.ToAddress, err)
				time.Sleep(time.Second * 1)
			} else {
				callBack()
				break
			}
		}

		if i+1 >= replicationSize {
			break
		} else {
			err = dataNode.metaStoreManger.PutReplications(replicationKey, replications[i+1:])
			if err != nil {
				fmt.Printf("DataNode[%s]-%s metaStoreManger.PutReplications(%v) fail. err:%s \n",
					dataNode.Config.Host, dataNode.Config.DataNodeId,
					replications[i+1:], err)
			} else {
				fmt.Printf("DataNode[%s]-%s metaStoreManger.PutReplications(%v) success. \n",
					dataNode.Config.Host, dataNode.Config.DataNodeId,
					replications[i+1:])
			}
		}
	}

	err = dataNode.taskStoreManger.Delete(replicationKey)
	if err != nil {
		fmt.Printf("DataNode[%s]-%s taskStoreManger.Delete(%s) fail. err:%v \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId,
			replicationKey, err)
		return
	} else {
		fmt.Printf("DataNode[%s]-%s taskStoreManger.Delete(%s) success. \n",
			dataNode.Config.Host, dataNode.Config.DataNodeId,
			replicationKey)
	}
}
