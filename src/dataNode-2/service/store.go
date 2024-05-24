package service

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	proto "trainfs/src/profile"
)

type StoreManger struct {
	path string
	db   *leveldb.DB
}

func OpenStoreManager(path string) *StoreManger {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		fmt.Printf(" DataNode Open level.db path %s file fail,err :%s \n", path, err)
	}
	return &StoreManger{
		db:   db,
		path: path,
	}
}

func (m *StoreManger) PutReplications(key string, value []*Replication) error {
	data, err := replications2bytes(value)
	if data != nil {
		err := m.db.Put([]byte(key), data, nil)
		if err != nil {
			return err
		}
	}
	return err
}

func (m *StoreManger) GetReplications(key string) ([]*Replication, error) {
	data, err := m.db.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}
	m2, err := bytes2Replications(data)
	if m2 != nil {
		return m2, nil
	}
	return nil, err
}

func (m *StoreManger) PutChunkInfos(key string, value map[string]*proto.ChunkInfo) error {
	data, err := chunkInfos2bytes(value)
	if data != nil {
		err := m.db.Put([]byte(key), data, nil)
		if err != nil {
			return err
		}
	}
	return err
}

func (m *StoreManger) GetChunkInfos(key string) (map[string]*proto.ChunkInfo, error) {
	value, err := m.db.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}
	m2, err := bytes2ChunkInfos(value)
	if m2 != nil {
		return m2, nil
	}
	return nil, err
}

func (m *StoreManger) PutTrashs(key string, value []string) error {
	data, err := strings2bytes(value)
	if data != nil {
		err := m.db.Put([]byte(key), data, nil)
		if err != nil {
			return err
		}
	}
	return err
}

func (m *StoreManger) GetPutTrashs(key string) ([]string, error) {
	value, err := m.db.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}
	m2, err := bytes2Strings(value)
	if m2 != nil {
		return m2, nil
	}
	return nil, err
}

func (m *StoreManger) Put(key string, value []byte) error {
	err := m.db.Put([]byte(key), value, nil)
	if err != nil {
		return err
	}
	return nil
}

func (m *StoreManger) Get(key string) ([]byte, error) {
	value, err := m.db.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (m *StoreManger) Delete(key string) error {
	err := m.db.Delete([]byte(key), nil)
	if err != nil {
		return err
	}
	return nil
}

func bytes2Replications(value []byte) ([]*Replication, error) {
	if value == nil {
		return nil, nil
	}
	decoder := gob.NewDecoder(bytes.NewBuffer(value))
	buf := make([]*Replication, 0)
	err := decoder.Decode(&value)
	if err != nil {
		fmt.Printf("bytes2Replications[] decode error:%s \n", err)
		return nil, err
	}
	return buf, nil
}

func replications2bytes(value []*Replication) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	buff := make([]byte, 0)
	buf := bytes.NewBuffer(buff)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(value)
	if err != nil {
		fmt.Printf("replications2bytes[%s] encode error:%s \n", value, err)
		return nil, err
	}
	return buff, nil
}

func bytes2ChunkInfos(value []byte) (map[string]*proto.ChunkInfo, error) {
	if value == nil {
		return nil, nil
	}
	decoder := gob.NewDecoder(bytes.NewBuffer(value))
	buf := make(map[string]*proto.ChunkInfo)
	err := decoder.Decode(&value)
	if err != nil {
		fmt.Printf("bytes2ChunkInfos decode error:%s \n", err)
		return nil, err
	}
	return buf, nil
}

func chunkInfos2bytes(value map[string]*proto.ChunkInfo) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	buff := make([]byte, 0)
	buf := bytes.NewBuffer(buff)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(value)
	if err != nil {
		fmt.Printf("chunkInfos2bytes[%v] encode error:%s \n", value, err)
		return nil, err
	}
	return buff, nil
}

func bytes2Strings(value []byte) ([]string, error) {
	if value == nil {
		return nil, nil
	}
	// 反序列化
	var strings []string
	err := json.Unmarshal(value, &strings)
	if err != nil {
		fmt.Printf("bytes2Strings[%v] Unmarshal error:%s \n", value, err)
		return nil, err
	}
	// 输出反序列化后的字符串切片
	return strings, nil
}

func strings2bytes(value []string) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		fmt.Printf("strings2bytes[%v] Marshal error:%s \n", value, err)
		return nil, err
	}
	// 输出序列化后的JSON字符串
	return jsonBytes, nil
}
