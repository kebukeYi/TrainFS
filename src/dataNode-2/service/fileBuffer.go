package service

import "sync"

type FileBuffer struct {
	mux    sync.Mutex
	buffer map[int][]byte
}

func NewFileBuffer() *FileBuffer {
	return &FileBuffer{buffer: map[int][]byte{}}
}

func (f *FileBuffer) Put(chunkId int, data []byte) {
	f.mux.Lock()
	defer f.mux.Unlock()
	f.buffer[chunkId] = data
}

func (f *FileBuffer) Get(chunkId int) []byte {
	f.mux.Lock()
	defer f.mux.Unlock()
	return f.buffer[chunkId]
}

func (f *FileBuffer) Delete(chunkId int) {
	f.mux.Lock()
	defer f.mux.Unlock()
	delete(f.buffer, chunkId)
}
