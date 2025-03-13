package service

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/kebukeYi/TrainDB"
	"github.com/kebukeYi/TrainDB/common"
	"github.com/kebukeYi/TrainDB/lsm"
	"github.com/kebukeYi/TrainDB/model"
	"log"
)

type TaskStoreManger struct {
	path string
	db   *TrainDB.TrainKVDB
}

func OpenTaskStoreManger(path string) *TaskStoreManger {
	trainKVDB, err, _ := TrainDB.Open(lsm.GetLSMDefaultOpt(path))
	if err != nil {
		log.Fatalln(" TaskStoreManger Open trainKVDB fail,", err)
	}
	return &TaskStoreManger{
		db:   trainKVDB,
		path: path,
	}
}

func (m *TaskStoreManger) PutReplications(key string, value []*Replication) error {
	data, err := replications2bytes(value)
	if data != nil {
		err := m.db.Set(model.NewEntry([]byte(key), data))
		if err != nil {
			return err
		}
	}
	return err
}

func (m *TaskStoreManger) GetReplications(key string) ([]*Replication, error) {
	data, err := m.db.Get([]byte(key))
	if err != nil {
		if err == common.ErrKeyNotFound {
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

func (m *TaskStoreManger) PutTrashs(key string, value []string) error {
	data, err := strings2bytes(value)
	if data != nil {
		err := m.db.Set(model.NewEntry([]byte(key), data))
		if err != nil {
			return err
		}
	}
	return err
}

func (m *TaskStoreManger) GetTrashs(key string) ([]string, error) {
	value, err := m.db.Get([]byte(key))
	if err != nil {
		if err == common.ErrKeyNotFound {
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

func (m *TaskStoreManger) Delete(key string) error {
	err := m.db.Del([]byte(key))
	if err != nil {
		if err == common.ErrKeyNotFound {
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
