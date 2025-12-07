# TrainFS Core Processes

## File Upload Process (PutFile)

```
Client                    NameNode                   DataNode-1/2/3
  |                          |                            |
  |--- 1. PutFile Request -->|                            |
  |<-- 2. Return DataNode Chain --|                       |
  |                          |                            |
  |--- 3. Send Chunk ---------------------------------->  DN-1
  |                          |                      4. Save Chunk
  |                          |                      5. Forward to DN-2
  |                          |                            |
  |                          |<-- 6. CommitChunk ---------|  
  |                          |      (Each DataNode commits)|
  |                          |                            |
  |--- 7. ConfirmFile ------>|                            |
  |<-- 8. Upload Confirmed --|                            |
```

**Key Steps**:
1. Client requests upload from NameNode with file size, replica count, etc.
2. NameNode selects DataNodes sorted by remaining space, returns address chain
3. Client splits file into Chunks, sends sequentially to the first DataNode
4. DataNode saves data and forwards in chain to downstream nodes
5. Each DataNode commits CommitChunk to NameNode
6. NameNode updates `<Chunk, [DataNode...]>` mapping
7. Client optionally confirms file integrity

---

## File Download Process (GetFile)

```
Client                    NameNode                   DataNode
  |                          |                          |
  |--- 1. GetFile Request -->|                          |
  |<-- 2. Return ChunkInfo --|                          |
  |      (with Chunk locations)|                        |
  |                          |                          |
  |--- 3. Request Chunk_0 -----------------------------> DN-1
  |<-- 4. Return Data ---------------------------------- DN-1
  |                          |                          |
  |--- 5. Request Chunk_1 -----------------------------> DN-2
  |<-- 6. Return Data ---------------------------------- DN-2
  |                          |                          |
  |   7. Merge and write file locally                   |
```

**Key Steps**:
1. Client queries NameNode for file metadata
2. NameNode returns storage locations for each Chunk (multiple replicas)
3. Client retrieves Chunks from DataNodes in order
4. Automatically switches to other replica nodes on failure
5. Verifies file size after local merge

---

## DataNode Registration and Heartbeat

```
DataNode                  NameNode
  |                          |
  |--- 1. RegisterDataNode -->|  (Register on startup, report disk space)
  |<-- 2. Registration Success -|
  |                          |
  |--- 3. ChunkReport ------->|  (Report full Chunk information)
  |<-- 4. Acknowledgment -----|
  |                          |
  |--- 5. HeartBeat --------->|  (Periodic heartbeat)
  |<-- 6. Heartbeat Response -|  (May carry delete/replication tasks)
  |      trashChunkNames      |
  |      replicationTask      |
```

**Tasks Carried in Heartbeat Response**:
- `trashChunkNames`: List of Chunks to delete
- `replicationTask`: Chunks to replicate to other nodes

---

## Failure Recovery and Replica Balancing

```
NameNode                  DataNode-1(Alive)         DataNode-2(Down)
  |                          |                          X
  |                          |                          X
  |<-- Heartbeat Timeout ----|                          X
  |                          |                          |
  |   1. Mark DN-2 Offline   |                          |
  |   2. Count DN-2's Chunks |                          |
  |   3. Select Source DN-1  |                          |
  |   4. Select Target DN-3  |                          |
  |                          |                          |
  |--- 5. Dispatch Replication Task (via heartbeat) -->|
  |                          |--- 6. Send Chunk ------> DN-3
  |                          |                          |
  |<-- 7. CommitChunk -------|  (DN-3 commits)          |
```

**ReplicationBalance Process**:
1. NameNode detects heartbeat timeout, marks node offline
2. Iterates through offline node's Chunk list
3. Finds alive source node for each Chunk
4. Selects target node that doesn't store this Chunk
5. Dispatches replication task via heartbeat response
6. Source node actively pushes data to target node
7. Target node saves and commits to NameNode

---

## File Deletion Process

```
Client                    NameNode                   DataNode
  |                          |                          |
  |--- 1. DeleteFile ------->|                          |
  |                          |                          |
  |   2. Delete Metadata     |                          |
  |   3. Record Trash Task   |                          |
  |                          |                          |
  |<-- 4. Delete Confirmed --|                          |
  |                          |                          |
  |                          |<-- 5. HeartBeat ---------|  
  |                          |--- 6. Response(trashNames) -->|
  |                          |                          |
  |                          |      7. Delete Chunk Files |
  |                          |<-- 8. CommitChunk -------|
  |                          |      (DeleteFileChunk)   |
```

**Async Deletion Design**:
- NameNode deletes metadata first, responds to client immediately
- Delete tasks are persisted to prevent loss
- DataNode gets tasks via heartbeat, commits after execution
- Supports resuming unfinished delete tasks after node restart

---

## Technical Highlights

| Module | Implementation |
|--------|----------------|
| Communication | gRPC + Protobuf |
| Metadata Storage | TrainKV (Custom KV Engine) |
| Data Transfer | Streaming |
| Node Selection | Sorted by Remaining Space |
| Task Persistence | KV Storage (Supports Restart Recovery) |
| Concurrency Control | sync.RWMutex |
