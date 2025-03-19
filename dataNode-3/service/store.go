package service

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kebukeYi/TrainDB"
	DBCommon "github.com/kebukeYi/TrainDB/common"
	"github.com/kebukeYi/TrainDB/lsm"
	"github.com/kebukeYi/TrainDB/model"
	"github.com/kebukeYi/TrainFS/common"
	proto "github.com/kebukeYi/TrainFS/profile"
)

type StoreManger struct {
	path string
	db   *TrainDB.TrainKVDB
}

func OpenStoreManager(path string) *StoreManger {
	trainKVDB, err, _ := TrainDB.Open(lsm.GetLSMDefaultOpt(path))
	if err != nil {
		fmt.Printf(" DataNode Open level.db path %s file fail,err :%s \n", path, err)
	}
	return &StoreManger{
		db:   trainKVDB,
		path: path,
	}
}

func (m *StoreManger) PutChunkInfos(key string, value map[string]*proto.ChunkInfo) error {
	data, err := chunkInfos2bytes(value)
	if data != nil {
		m.db.Set(model.Entry{Key: []byte(key), Value: data})
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *StoreManger) GetChunkInfos(key string) (map[string]*proto.ChunkInfo, error) {
	entry, err := m.db.Get([]byte(key))
	if err != nil || entry.Version == -1 {
		if errors.Is(err, DBCommon.ErrKeyNotFound) {
			return make(map[string]*proto.ChunkInfo), nil
		}
		return nil, err
	}
	m2, err := bytes2ChunkInfos(entry.Value)
	if m2 != nil {
		return m2, nil
	}
	return nil, err
}

func (m *StoreManger) Put(key string, value []byte) error {
	err := m.db.Set(model.Entry{Key: []byte(key), Value: value})
	if err != nil {
		return err
	}
	return nil
}

func (m *StoreManger) Get(key string) ([]byte, error) {
	value, err := m.db.Get([]byte(key))
	if err != nil || value.Version == -1 {
		return nil, err
	}
	return value.Value, nil
}

func (m *StoreManger) Delete(key string) error {
	err := m.db.Del([]byte(key))
	if err != nil {
		if errors.Is(err, DBCommon.ErrKeyNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func (m *StoreManger) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

func (m *StoreManger) PutReplications(key string, value []*Replication) error {
	data, err := replications2bytes(value)
	if data != nil {
		err = m.db.Set(model.Entry{Key: []byte(key), Value: data})
		if err != nil {
			return err
		}
	}
	return err
}

func (m *StoreManger) GetReplications(key string) ([]*Replication, error) {
	data, err := m.db.Get([]byte(key))
	if err != nil || data.Version == -1 {
		if errors.Is(err, DBCommon.ErrKeyNotFound) {
			return make([]*Replication, 0), nil
		}
		return nil, err
	}
	m2, err := bytes2Replications(data.Value)
	if m2 != nil {
		return m2, nil
	}
	return nil, err
}

func (m *StoreManger) PutTrashes(key string, value []string) error {
	data, err := strings2bytes(value)
	if data != nil {
		err := m.db.Set(model.Entry{Key: []byte(key), Value: data})
		if err != nil {
			return err
		}
	}
	return err
}

func (m *StoreManger) GetTrashes(key string) ([]string, error) {
	value, err := m.db.Get([]byte(key))
	if err != nil || value.Version == -1 {
		if errors.Is(err, DBCommon.ErrKeyNotFound) {
			return make([]string, 0), nil
		}
		return nil, err
	}
	m2, err := bytes2Strings(value.Value)
	if m2 != nil {
		return m2, nil
	}
	return nil, err
}

func bytes2Replications(value []byte) ([]*Replication, error) {
	if value == nil || len(value) == 0 {
		return nil, common.ErrInputEmpty
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
	if value == nil || len(value) == 0 {
		return nil, common.ErrInputEmpty
	}
	buff := make([]byte, 0)
	buf := bytes.NewBuffer(buff)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(value)
	if err != nil {
		fmt.Printf("replications2bytes[%v] encode error:%s \n", value, err)
		return nil, err
	}
	return buff, nil
}

func bytes2ChunkInfos(value []byte) (map[string]*proto.ChunkInfo, error) {
	if value == nil || len(value) <= 0 {
		return nil, common.ErrInputEmpty
	}
	decoder := gob.NewDecoder(bytes.NewBuffer(value))
	buf := make(map[string]*proto.ChunkInfo)
	err := decoder.Decode(&buf)
	if err != nil {
		fmt.Printf("bytes2ChunkInfos decode error:%s \n", err)
		return nil, err
	}
	return buf, nil
}

func chunkInfos2bytes(value map[string]*proto.ChunkInfo) ([]byte, error) {
	if value == nil || len(value) == 0 {
		return nil, common.ErrInputEmpty
	}
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(value)
	if err != nil {
		fmt.Printf("chunkInfos2bytes[%v] encode error:%s \n", value, err)
		return nil, err
	}
	return buff.Bytes(), nil
}

func bytes2Strings(value []byte) ([]string, error) {
	if value == nil || len(value) == 0 {
		return nil, common.ErrInputEmpty
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
	if value == nil || len(value) == 0 {
		return nil, common.ErrInputEmpty
	}
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		fmt.Printf("strings2bytes[%v] Marshal error:%s \n", value, err)
		return nil, err
	}
	// 输出序列化后的JSON字符串
	return jsonBytes, nil
}
