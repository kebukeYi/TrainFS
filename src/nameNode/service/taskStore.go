package service

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
)

type TaskStoreManger struct {
	path string
	db   *leveldb.DB
}

func OpenTaskStoreManger(path string) *TaskStoreManger {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		log.Fatalln(" NameNode Open level.db file fail,", err)
	}
	return &TaskStoreManger{
		db:   db,
		path: path,
	}
}

func (m *TaskStoreManger) PutReplications(key string, value []*Replication) error {
	data, err := replications2bytes(value)
	if data != nil {
		err := m.db.Put([]byte(key), data, nil)
		if err != nil {
			return err
		}
	}
	return err
}

func (m *TaskStoreManger) GetReplications(key string) ([]*Replication, error) {
	data, err := m.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return make([]*Replication, 0), nil
		}
		return nil, err
	}
	m2, err := bytes2Replications(data)
	if m2 != nil {
		return m2, nil
	}
	return nil, err
}

func (m *TaskStoreManger) PutTrashs(key string, value []string) error {
	data, err := strings2bytes(value)
	if data != nil {
		err := m.db.Put([]byte(key), data, nil)
		if err != nil {
			return err
		}
	}
	return err
}

func (m *TaskStoreManger) GetTrashs(key string) ([]string, error) {
	value, err := m.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return make([]string, 0), nil
		}
		return nil, err
	}
	m2, err := bytes2Strings(value)
	if m2 != nil {
		return m2, nil
	}
	return nil, err
}

func (m *TaskStoreManger) Delete(key string) error {
	err := m.db.Delete([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil
		}
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
		fmt.Printf("replications2bytes[] encode error:%s \n", err)
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
		fmt.Printf("bytes2Strings[] Unmarshal error:%s \n", err)
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
		fmt.Printf("strings2bytes[] Marshal error:%s \n", err)
		return nil, err
	}
	// 输出序列化后的JSON字符串
	return jsonBytes, nil
}
