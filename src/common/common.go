package common

import "errors"

var ErrFileCorruption = errors.New("file Corruption")

var ErrCanNotChangeRootDir = errors.New("can not change root dir")
var ErrPathFormatError = errors.New("path format error")
var ErrFileNotFound = errors.New("file not found")
var ErrNotDir = errors.New("the files is not directory")
var ErrFileAlreadyExists = errors.New("the file already exists")
var ErrCommitChunkType = errors.New("CommitChunk not found")
var ErrChunkReplicaNotFound = errors.New("chunk replica not found")

var ErrNotEnoughStorageSpace = errors.New("not enough storage space ")
var ErrEnoughReplicaDataNodeServer = errors.New("not enough replica num dataNodeServer")
