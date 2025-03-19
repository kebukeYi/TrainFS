package service

type datanodeStatus string
type replicaState string

type FileMeta struct {
	KeyFileName string // /usr/app
	FileName    string // app
	FileSize    int64
	IsDir       bool // true or false
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
