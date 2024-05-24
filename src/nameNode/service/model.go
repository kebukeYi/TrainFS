package service

type datanodeStatus string
type replicaState string

type FileMeta struct {
	KeyFileName string
	FileName    string
	FileSize    int64
	IsDir       bool
	ChildList   map[string]*FileMeta
	Chunks      []*ChunkMeta
	ChunkNum    int64
}

type ChunkMeta struct {
	ChunkName string
	ChunkId   int32
	TimeStamp int64
}

type ReplicaMeta struct {
	ChunkId         int32
	ChunkName       string
	DataNodeAddress string
	Status          replicaState
}
