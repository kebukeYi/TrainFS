# TrainFS 核心流程

## 文件上传流程（PutFile）

```
Client                    NameNode                   DataNode-1/2/3
  |                          |                            |
  |--- 1. PutFile请求 ------->|                            |
  |<-- 2. 返回DataNode链 -----|                            |
  |                          |                            |
  |--- 3. 发送Chunk ---------------------------------->  DN-1
  |                          |                      4. 保存Chunk
  |                          |                      5. 转发至DN-2
  |                          |                            |
  |                          |<-- 6. CommitChunk ---------|  
  |                          |      (各DataNode提交)       |
  |                          |                            |
  |--- 7. ConfirmFile ------>|                            |
  |<-- 8. 上传确认 -----------|                            |
```

**关键步骤**：
1. 客户端向 NameNode 请求上传，携带文件大小、副本数等信息
2. NameNode 按剩余空间排序选择 DataNode，返回地址链
3. 客户端将文件切分为 Chunk，依次发送到首个 DataNode
4. DataNode 保存数据后链式转发给下游节点
5. 各 DataNode 向 NameNode 提交 CommitChunk
6. NameNode 更新 `<Chunk, [DataNode...]>` 映射
7. 客户端可选确认文件完整性

---

## 文件下载流程（GetFile）

```
Client                    NameNode                   DataNode
  |                          |                          |
  |--- 1. GetFile请求 ------->|                          |
  |<-- 2. 返回ChunkInfo ------|                          |
  |      (含各Chunk位置)       |                          |
  |                          |                          |
  |--- 3. 请求Chunk_0 --------------------------------> DN-1
  |<-- 4. 返回数据 ------------------------------------ DN-1
  |                          |                          |
  |--- 5. 请求Chunk_1 --------------------------------> DN-2
  |<-- 6. 返回数据 ------------------------------------ DN-2
  |                          |                          |
  |   7. 本地合并写入文件     |                          |
```

**关键步骤**：
1. 客户端向 NameNode 查询文件元数据
2. NameNode 返回各 Chunk 的存储位置（多副本）
3. 客户端按顺序从各 DataNode 获取 Chunk
4. 失败时自动切换到其他副本节点
5. 本地合并后校验文件大小

---

## DataNode 注册与心跳

```
DataNode                  NameNode
  |                          |
  |--- 1. RegisterDataNode -->|  (启动时注册，上报磁盘空间)
  |<-- 2. 注册成功 -----------|
  |                          |
  |--- 3. ChunkReport ------->|  (上报全量Chunk信息)
  |<-- 4. 确认 ---------------|
  |                          |
  |--- 5. HeartBeat --------->|  (周期心跳)
  |<-- 6. 心跳响应 -----------|  (可携带删除/复制任务)
  |      trashChunkNames      |
  |      replicationTask      |
```

**心跳响应携带任务**：
- `trashChunkNames`: 需删除的 Chunk 列表
- `replicationTask`: 需复制到其他节点的 Chunk

---

## 故障恢复与副本均衡

```
NameNode                  DataNode-1(存活)          DataNode-2(宕机)
  |                          |                          X
  |                          |                          X
  |<-- 心跳超时检测 ----------|                          X
  |                          |                          |
  |   1. 标记DN-2下线        |                          |
  |   2. 统计DN-2的Chunk     |                          |
  |   3. 选择源节点DN-1      |                          |
  |   4. 选择目标节点DN-3    |                          |
  |                          |                          |
  |--- 5. 下发复制任务(心跳) ->|                          |
  |                          |--- 6. 发送Chunk -------> DN-3
  |                          |                          |
  |<-- 7. CommitChunk --------|  (DN-3提交)              |
```

**ReplicationBalance 流程**：
1. NameNode 检测心跳超时，标记节点下线
2. 遍历下线节点的 Chunk 列表
3. 为每个 Chunk 找到存活的源节点
4. 选择未存储该 Chunk 的目标节点
5. 通过心跳响应下发复制任务
6. 源节点主动推送数据到目标节点
7. 目标节点保存后提交给 NameNode

---

## 文件删除流程

```
Client                    NameNode                   DataNode
  |                          |                          |
  |--- 1. DeleteFile -------->|                          |
  |                          |                          |
  |   2. 删除元数据           |                          |
  |   3. 记录Trash任务        |                          |
  |                          |                          |
  |<-- 4. 删除确认 -----------|                          |
  |                          |                          |
  |                          |<-- 5. HeartBeat ---------|  
  |                          |--- 6. 响应(trashNames) -->|
  |                          |                          |
  |                          |      7. 删除Chunk文件    |
  |                          |<-- 8. CommitChunk -------|
  |                          |      (DeleteFileChunk)   |
```

**异步删除设计**：
- NameNode 先删除元数据，立即响应客户端
- 删除任务持久化保存，防止丢失
- DataNode 通过心跳获取任务，执行后提交确认
- 支持节点重启后继续执行未完成的删除任务

---

## 技术要点

| 模块 | 技术实现 |
|------|----------|
| 通信框架 | gRPC + Protobuf |
| 元数据存储 | TrainKV（自研KV引擎） |
| 数据传输 | 流式传输（Streaming） |
| 节点选择 | 按剩余空间排序 |
| 任务持久化 | KV存储（支持重启恢复） |
| 并发控制 | sync.RWMutex |
