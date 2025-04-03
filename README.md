# TrainFS

本单机文件系统,支持基本读写删除,目录等操作; 元数据存储NameNode中, 文件大小切分成块, 存储在DataNode中.

---

### 启动开始
````
NameNode节点启动
A方式: 本地编辑器源码调试,run kind 选择 package 即可;
注意项目参数输入指定配置信息
program arguments: -conf=nameNode/conf/nameNode_config.yml

B方式: 下载源码后,进入源码目录中直接启动;
cd TrainFS/nameNode
go run nameNode_ctrl.go -conf=./conf/nameNode_config.yml

C: 构建可执行文件方式启动;
cd TrainFS/nameNode
go build -o ./build/nameNode
cd ../build
./nameNode

DataNode-* 多节点启动
A方式: 本地编辑器源码调试,run kind 选择 package 即可:
注意项目参数输入指定配置信息: program arguments: -id=1 -port=9001 -conf=dataNode/conf/dataNode_config.yml

B方式: 下载源码后,进入源码目录中直接启动:
cd TrainFS/dataNode
go run dataNode_ctrl.go -id=1 -port=9001 -conf=./conf/dataNode_config.yml
go run dataNode_ctrl.go -id=2 -port=9002 -conf=./conf/dataNode_config.yml
go run dataNode_ctrl.go -id=3 -port=9003 -conf=./conf/dataNode_config.yml

C方式: 构建可执行文件方式启动:
cd TrainFS/dataNode
go build -o ./build/dataNode
cd ../build
./dataNode -id=1 -port=9001
./dataNode -id=2 -port=9002
./dataNode -id=3 -port=9003

进入client源码中,测试启动
go test -run TestPutFile
go test -run TestGetFile
go test -run TestDelete

额外: 修改.proto文件时,需要提前安装好protoc插件
go get -u google.golang.org/protobuf/proto
go get -u google.golang.org/protobuf/protoc-gen-go
cd profile
protoc --go_out=. --go-grpc_out=. ./*.proto
````
---

举例 "TrainFS-Put流程图":
![TrainFS-Put流程图](docs/TrainFS-Put.png )

## 功能讲解

- PutFile(localFilePath, remotePath)
    - 举例: PutFile("/home/user/test.txt", "/app")
    - 过程: 将本地文件 localFilePath`/home/user/test.txt` 上传到文件系统中,注意文件系统并不会真正的创建 remoteFilePath`/app/test.txt`此目录文件,
      而是基于`/app/test.txt`文件名,切分成`/app/test.txt_chunk_0`, `/app/test.txt_chunk_1`...等文件块名字,构成<remoteFilePath,[chunk_1,chunk_2,chunk_3]>
      kv映射集合元数据,并进行持久化保存; 之后文件系统中NameNode节点返回用户指定副本数量的DataNode节点地址; 客户端仅仅把文件块发送到 primary DataNode单个节点即可;
      文件系统DataNode节点之间会自动转发用户数据,
      DataNode节点将文件块存储到本地文件系统中,也会记录[`/app/test.txt_chunk_2`, `/app/test.txt_chunk_3`....]kv映射元数据;DataNode接收完毕后,
      并提交给NameNode节点,报告给NameNode本次保存了哪些数据;用户上传完毕后,可向NameNode确认文件是否上传完毕; 此时PutFile操作完成, NameNode将返回用户上传成功信息;

- GetFile(remoteFilePath, localPath)
    - 举例: GutFile("/app/test.txt","/home/user")
    - 过程：用户首先向NameNode获知remoteFilePath`/home/user/test.txt`的存储副本的多个DataNode节点地址,然后向DataNode节点请求获得文件块;
- Mkdir(remoteDirPath)
    - 举例: Mkdir("/home/user")
    - 过程：在NameNode的元数据kv存储中,将remoteDirPath`/home/user`添加到元数据中,并设置状态为目录,同时找到此目录的上级目录信息,
      将本节点添加到上层目录的子目录list中, 随后同时持久化保存;
- DeleteFile(remoteDirPath)
    - 举例:DeleteFile("/home/user/test.txt") || DeleteFile("/home/user")
    - 说明:目前仅支持删除末尾文件和末尾目录,不支持删除有 含有子文件或者子文件夹 的目录;
    - 过程:在NameNode的元数据kv存储中,将remoteDirPath`/home/user/test.txt`进行删除,同时找到此目录的上级目录信息,
      将本节点删除，随后针对上层目录节点持久化保存更新;然后在找到对应的DataNode,进行保存Trash删除任务;此时DataNode并不会立即删除,会在随后的心跳响应中,NameNode会下发删除任务;
      DataNode接收任务后,DataNode会删除文件块,并提交给NameNode节点,报告给NameNode本次删除了哪些数据;
- ListDir(remoteDirPath)
    - 举例:ListDir("/home/user")
    - 过程:直接在NameNode的元数据存储中,查询remoteDirPath`/home/user`的节点信息,并返回给用户当前目录节点的子节点的信息列表;
- ReName(oldRemoteDirPath,newRemoteDirPath)
    - 举例:ReName("/home/user","/home/newUser")
    - 过程:直接在NameNode的元数据存储中,查询remoteDirPath`/home/user`的节点信息,并判断是否为目录,并且是空目录,则进行重命名操作,否则报错返回,不支持本次操作;

### NameNode 元数据中心
记录用户每一次的操作,以及文件的副本存储信息;

第一步，针对用户上传文件过程，NameNode可把 remoteFilePath`/app/test.txt`文件,保存在NameNode的元数据kv存储中,此时并没有保存文件的块信息;
第二步，根据DataNode的上传信息，NameNod在内存中做出统计，将`/app/test.txt_chunk_1`, `/app/test.txt_chunk_2`等文件块的存储地址进行持久化保存，
再统计出 <remoteFilePath,[chunk1,chunk2,chunk3]>、<chunk_1,[dataNode1,dataNode2,dataNode3]> 两大基本核心kv映射；

为了基本目标做出的维护:

1. 接受并保存DataNode的注册请求和保存其DataNode上报的文件块信息，并持久化保存；
2. 针对已经注册的DataNode，定期检测其心跳信息，如果发现心跳超时，则进行剔除，并启动文件迁移复制任务;
3. 当用户删除文件时，NameNode会首先在本身的元数据kv存储中，删除此文件块的存储地址信息，并持久化更新; 随后在相关的DataNode发送心跳时，
   将相关的删除文件块信息，返回给DataNode，让其随后删除，当其删除后，会再次上报给NameNode，此时NameNode会更新此DataNode的剩余磁盘空间;
4. 文件迁移复制任务，NameNode统计出下线的DataNode节点存储了哪些文件块，然后针对这些文件块，NameNode会向拥有文件块的DataNode节点下发拷贝转发指令，
    向其他未保存此文件块的DataNode主动发送文件块, 后面的DataNode接收数据并保存后,也会提交给NameNode节点,报告给NameNode本次保存了哪些数据;

### DataNode
记录所有文件块数据,接收用户上传的文件，并使用链式传递文件。


为了基本目标做出的维护：
1. DataNode在启动时，向NameNode注册自己，同时上报自己存储的磁盘空间信息；
2. 上报一次自己的全量存储文件块情况，NameNode需要知道这个；
3. 周期上报心跳，让NameNode知道自己还在线；
4. 保存用户的文件块信息，DataNode会保存这些文件块信息，并记录到本地文件系统，也会持久化到本次的kv映射元数据中,[chunk1,chunk2,chunk3...];


#### 下一步
1. NameNode存储的元数据结构设计,减少锁粒度,也能更好的支持 ReName() 操作;
2. NameNode的单点宕机问题,可升级成 分片式分布式集群,NameNode集群会根据 用户操作 进行分片,将文件负载均衡到到 各个NameNode集群中,
每个NameNode集群会依据共识算法来保证集群可用性,进一步降低单点风险,提高可用性;
3. 允许DataNode可先接收来自客户端的数据,然后转发给其他DataNode节点,实现文件块的快速复制,最后再提及给NameNode节点;
