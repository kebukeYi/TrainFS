# TrainFS

本单机文件系统,支持基本读写删除,目录等操作; 元数据存储NameNode中,文件切分成块,存储在DataNode中.

## 功能

- PutFile(localFilePath, remoteFilePath)
    - 举例: PutFile("/home/user/test.txt", "/app/test.txt")
    - 过程: 将本地目录文件localFilePath`/home/user/test.txt` 上传到文件系统中，注意文件系统并不会真正的创建 remoteFilePath`/app/test.txt`此目录，
      而是基于`/app/test.txt`文件名，切分成`/app/test.txt_chunk_1`, `/app/test.txt_chunk_2`,...等文件块名字，构成<
      remoteFilePath,[chunk_1,chunk_2,chunk_3]>
      kv映射集合元数据，并进行持久化保存；之后NameNode再返回用户指定副本数量的DataNode节点地址，将文件块逐个发送到 primary DataNode节点，底层DataNode节点之间会自动传递用户数据，
      DataNode节点将文件块存储到本地文件系统，也会记录[`/app/test.txt_chunk_2`, `/app/test.txt_chunk_3`....]元数据；用户上传完毕后，
      DataNode并告诉给NameNode，我此时保存了哪些数据；用户会再次向NameNode确认文件是否上传完毕； 此时PutFile操作完成，NameNode将返回用户上传成功信息；

- GetFile(remoteFilePath, localFilePath)
    - 举例: GutFile("/home/user/test.txt", "/app/test.txt")
    - 过程：用户首先向NameNode获知remoteFilePath`/home/user/test.txt`的副本数量的DataNode节点地址，然后向DataNode节点请求文件块；
- Mkdir(remoteDirPath)
    - 举例: Mkdir("/home/user/test.txt") || Mkdir("/home/user")
    - 过程：在NameNode的元数据kv存储中，将remoteDirPath`/home/user/test.txt`添加到元数据中，并设置状态为目录，同时找到此目录的上级目录信息，
      将本节点添加到上层目录中，随后同时持久化保存更新；
- DeleteFile(remoteDirPath)
    - 举例：DeleteFile("/home/user/test.txt") || DeleteFile("/home/user")
    - 过程：在NameNode的元数据kv存储中，将remoteDirPath`/home/user/test.txt`进行删除，同时找到此目录的上级目录信息，
      将本节点删除，随后针对上层目录节点持久化保存更新，此时DataNode并不会立即删除，会在随后的心跳响应中，NameNode会下发删除指令； 当删除后，文件副本数量不够时，同样会在心跳响应中下发拷贝指令；
- ListDir(remoteDirPath)
    - 举例：ListDir("/home/user/test.txt") || ListDir("/home/user")
    - 过程：直接在NameNode的元数据存储中，查询remoteDirPath`/home/user/test.txt`的节点信息，并返回给用户次节点的子节点信息列表；

### NameNode

基本目标：保证知道 文件存储在哪些地方

第一步，针对用户上传文件过程，NameNode可把 remoteFilePath`/app/test.txt`文件名，切分成`/app/test.txt_chunk_1`, `/app/test.txt_chunk_2`
,...等文件块名字； 第二步，根据DataNode的上传信息，NameNod在内存中做出统计，将`/app/test.txt_chunk_1`, `/app/test.txt_chunk_2`等文件块的存储地址进行持久化保存，
从而可构造出 <remoteFilePath,[chunk1,chunk2,chunk3]>、<chunk1,[dataNode1,dataNode2,dataNode3]> 两大基本核心map映射；

为了基本目标做出的维护：

1. 接受并保存DataNode的注册请求
2. 针对已经注册的DataNode，定期检测其心跳信息，如果发现心跳超时异常，则进行自我剔除，并启动文件迁移任务；
3. 当用户删除文件时，NameNode会首先在本身的元数据kv存储中，删除此文件块的存储地址信息，并持久化保存；随后在相关的DataNode发送心跳时，
   将相关的删除文件块信息，返回给DataNode，让其随后删除，当其删除后，会再次上报给NameNode，此时NameNode会更新此DataNode的剩余磁盘空间；
4. 文件迁移任务，NameNode统计出下线的DataNode节点存储了哪些文件，然后针对这些文件块，NameNode会向拥有文件块的DataNode节点下发拷贝指令，向其他DataNode主动发送文件块；
5. 单点宕机问题，可升级成 分片式分布式集群，NameNode集群会根据 用户操作 进行分片，将文件负载均衡到到 各个NameNode集群中，每个NameNode集群会依据共识算法来保证集群可用性，进一步降低单点风险；

### DataNode

基本目标：保证知道 存储了哪些文件

接收用户上传的文件，并使用链式传递文件。

为了基本目标做出的维护：
1. DataNode在启动时，向NameNode注册自己，同时上报自己存储的磁盘空间信息；
2. 上报一次自己的全量存储文件块情况，NameNode需要知道这个；
3. 周期的上报心跳，让NameNode知道自己还在线；
4. 保存用户的文件块信息，DataNode会保存这些文件块信息，并记录到本地文件系统，也会持久化到本次的kv映射元数据中,[chunk1,chunk2,chunk3...];