# TrainFS
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-Apache_2.0-green)](https://opensource.org/licenses/Apache-2.0)


一个参考 HDFS 架构设计的分布式文件系统，支持文件读写删除、目录管理等基本操作。元数据存储于 NameNode，文件按块切分后存储在多个 DataNode 中，支持多副本冗余。

## 快速开始

### 1. 启动 NameNode

**方式 A：IDE 调试**
```bash
# GoLand 中 Run Kind 选择 Package
# Program Arguments:
-conf=nameNode/conf/nameNode_config.yml
```

**方式 B：源码运行**
```bash
cd TrainFS/nameNode
go run nameNode_ctrl.go -conf=./conf/nameNode_config.yml
```

**方式 C：编译运行**
```bash
cd TrainFS/nameNode
go build -o ./build/nameNode
./build/nameNode
```

### 2. 启动 DataNode（多节点）

**方式 A：IDE 调试**
```bash
# Program Arguments:
-id=1 -port=9001 -conf=dataNode/conf/dataNode_config.yml
```

**方式 B：源码运行**
```bash
cd TrainFS/dataNode
go run dataNode_ctrl.go -id=1 -port=9001 -conf=./conf/dataNode_config.yml
go run dataNode_ctrl.go -id=2 -port=9002 -conf=./conf/dataNode_config.yml
go run dataNode_ctrl.go -id=3 -port=9003 -conf=./conf/dataNode_config.yml
```

**方式 C：编译运行**
```bash
cd TrainFS/dataNode
go build -o ./build/dataNode
./build/dataNode -id=1 -port=9001
./build/dataNode -id=2 -port=9002
./build/dataNode -id=3 -port=9003
```

### 3. 客户端测试

```bash
cd TrainFS/client
go test -run TestPutFile
go test -run TestGetFile
go test -run TestDelete
```

### 4. 重新生成 Protobuf（可选）

```bash
go get -u google.golang.org/protobuf/proto@latest
go get -u google.golang.org/protobuf/protoc-gen-go@latest
go get -u google.golang.org/grpc/protoc-gen-go-grpc@latest

cd profile
protoc --go_out=. --go-grpc_out=. ./*.proto
```

---

## 系统架构

![TrainFS-Put流程图](doc/pngs/TrainFS-Put.png)

---

## API 说明

### PutFile(localFilePath, remotePath)
上传本地文件到分布式文件系统。

```go
PutFile("/home/user/test.txt", "/app")
```

**流程**：
1. 客户端向 NameNode 请求上传，获取 DataNode 地址链
2. 文件按配置大小切分为多个 Chunk（如 `test.txt_chunk_0`、`test.txt_chunk_1`）
3. 客户端将 Chunk 发送至首个 DataNode，节点间链式转发
4. DataNode 保存后向 NameNode 提交元数据
5. 客户端确认上传完成

### GetFile(remoteFilePath, localPath)
从分布式文件系统下载文件到本地。

```go
GetFile("/app/test.txt", "/home/user")
```

**流程**：向 NameNode 查询文件块位置，从各 DataNode 获取 Chunk 后合并。

### Mkdir(remoteDirPath)
创建远程目录。

```go
Mkdir("/home/user")
```

### DeleteFile(remotePath)
删除文件或空目录。

```go
DeleteFile("/home/user/test.txt")
DeleteFile("/home/user")  // 仅支持空目录
```

**流程**：NameNode 删除元数据，通过心跳下发删除任务给 DataNode。

### ListDir(remoteDirPath)
列出目录内容。

```go
ListDir("/home/user")
```

### ReName(oldPath, newPath)
重命名空目录。

```go
ReName("/home/user", "/home/newUser")
```

---

## 核心组件

### NameNode（元数据中心）

负责管理文件系统的元数据，维护两大核心映射：
- `<FilePath, [Chunk1, Chunk2, ...]>` - 文件与块的映射
- `<ChunkName, [DataNode1, DataNode2, ...]>` - 块与副本位置的映射

**职责**：
1. 接收 DataNode 注册与心跳，管理节点状态
2. 处理文件块提交，更新元数据映射
3. 检测节点故障，触发副本再均衡任务
4. 通过心跳响应下发删除/复制任务

### DataNode（数据节点）

负责实际数据块的存储与传输。

**职责**：
1. 启动时向 NameNode 注册，上报磁盘空间
2. 全量上报已存储的 Chunk 信息
3. 周期性发送心跳，执行下发的任务
4. 接收并存储 Chunk，链式转发至下游节点

---

## 后续规划

1. **优化元数据结构**：细化锁粒度，更好支持 ReName 操作
2. **NameNode 高可用**：升级为分片式分布式集群，基于共识算法保证可用性
3. **异步提交优化**：DataNode 先接收转发数据，再批量提交至 NameNode
