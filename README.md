# TrainFS

本目前是单机文件系统,支持基本读写删除文件或目录,以及目录列表等操作; 系统设计元数据存储NameNode中,用户文件数据切分成块,存储在一定数量的DataNode中.

## 功能

- PutFile(localFilePath, remoteFilePath)
    - 举例: PutFile("/home/user/test.txt", "/app/test.txt")
    - 过程: 将本地目录文件localFilePath`/home/user/test.txt` 上传到文件系统中，注意文件系统并不会真正的创建 remoteFilePath`/app/test.txt`此目录，
      而是基于`/app/test.txt`文件名，切分成`/app/test.txt_chunk_1`, `/app/test.txt_chunk_2`...等文件块名字，构成<
      remoteFilePath,[chunk_1,chunk_2,chunk_3]>
      kv映射集合元数据，并进行持久化保存；之后NameNode返回用户指定副本数量的DataNode节点地址，client会将文件块逐块发送到 primary DataNode 节点，只发送一次即可, 底层 primary DataNode 节点 会向其他DataNode之间自动传递用户文件数据，
      DataNode节点将文件块存储到本地存储引擎中,并没有实际的创建文件，同时也会记录[`/app/test.txt_chunk_2`, `/app/test.txt_chunk_3`....]等文件块名字的元数据；用户上传完毕后，
      DataNode发送给NameNode消息，内容是 此时保存了哪些数据；用户会再次向NameNode确认文件是否上传完毕,可设置ack,是否同步等待DataNode的上传存储信息,还是立即返回,此涉及到一致性问题； 此时PutFile()操作完成，NameNode和DataNode节点都保存了元数据；

- GetFile(remoteFilePath, localFilePath)
    - 举例: GutFile("/home/user/test.txt", "/app/test.txt")
    - 过程：用户首先向NameNode获知remoteFilePath`/home/user/test.txt`的副本数量的DataNode节点地址，然后向DataNode节点逐个请求文件块,注意文件块的请求顺序,以免文件内容拼接混乱；
- Mkdir(remoteDirPath)
    - 举例: Mkdir("/home/user/test.txt") || Mkdir("/home/user")
    - 过程：在NameNode的元数据kv存储中，将remoteDirPath`/home/user/test.txt`添加到元数据中，并设置状态为目录，同时找到此目录的上级目录信息，
      将本节点添加到上层目录的childList[]中，随后同时持久化保存更新；
- DeleteFile(remoteDirPath)
    - 举例：DeleteFile("/home/user/test.txt") || DeleteFile("/home/user")
    - 过程：在NameNode的元数据kv存储中，将remoteDirPath`/home/user/test.txt`进行删除，同时找到此目录的上级目录信息，
      将本节点信息删除，随后针对上层目录节点持久化保存更新; 此时DataNode并不会立即删除相关数据，在随后的NameNode的心跳响应中，NameNode会下发删除指令,这时DataNode才会真正的删除数据； 当NameNode下发删除指令后，当判断到文件副本数量不够时，同样会在心跳响应中下发拷贝指令；
- ListDir(remoteDirPath)
    - 举例：ListDir("/home/user/test.txt") || ListDir("/home/user")
    - 过程：在NameNode的元数据存储中，查询remoteDirPath`/home/user/test.txt`的节点信息，并返回给用户其次节点的子节点childList[]信息列表；

### NameNode

基本目标：维护 文件存储在哪些地方

用户上传文件过程，client把 remoteFilePath`/app/test.txt`文件名，切分成`/app/test.txt_chunk_1`, `/app/test.txt_chunk_2`
...等文件块名字,逐个发送到DataNode中；随后NameNode会根据DataNode的上报信息，在内存中做出统计，不仅记录remoteFilePath`/app/test.txt`文件的实际几个块名字,还将文件块`/app/test.txt_chunk_1`, `/app/test.txt_chunk_2`等,存储地址也进行持久化保存，
从而可构造出 <remoteFilePath,[chunk1,chunk2,chunk3]>、<chunk1,[dataNode1,dataNode2,dataNode3]> 两大基本核心kv映射；

为了基本目标做出的维护：

1. 接受并保存DataNode的注册请求
2. 针对已经注册的DataNode，定期检测其心跳信息，如果发现心跳超时异常，就进行自我剔除，并启动文件迁移任务；
3. 当用户删除文件时，NameNode会首先在本身的元数据kv存储中，删除此文件块的存储地址信息，并持久化保存；随后在相关的DataNode发送心跳时，
   将相关的删除文件块信息，响应返回给DataNode，让其随后删除，当其删除后，会再次上报给NameNode，此时NameNode会更新此DataNode的剩余磁盘空间；
4. 文件迁移任务，NameNode统计出下线的DataNode节点存储了哪些文件，然后针对这些文件块，NameNode会向拥有文件块的DataNode节点下发拷贝指令，向其他没有此数据的 DataNode主动发送文件块；
5. 目前单点宕机问题，后期将会升级成 分片式分布式集群，NameNode集群会根据 用户操作 进行分片，将文件操作 负载均衡到 各个NameNode集群中，DataNode节点也可用共识算法来保证高可用; 每个NameNode集群会依据共识算法来保证集群可用性，进一步降低单点风险；

### DataNode

基本目标：维护 存储了哪些文件

接收用户上传的文件，并使用链式传递文件。

为了基本目标做出的维护：
1. DataNode在每次启动时，都向NameNode注册自己，同时上报自己存储的磁盘剩余空间信息；
2. DataNode在每次启动时, 上报一次自己的全量存储文件块情况,也就是元数据，NameNode需要知道这个；
3. 周期上报心跳，NameNode知道自己还在线存活；
4. 保存用户的文件块信息，DataNode会保存这些文件块信息，并记录到本地文件系统存储引擎中，同时记录元数据,持久化到本次的kv映射元数据中[file1_chunk1,file1_chunk2,file_3_chunk1,file_3_chunk2...];
