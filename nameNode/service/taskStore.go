package service

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kebukeYi/TrainFS/common"
	"github.com/kebukeYi/TrainKV"
	DBcommon "github.com/kebukeYi/TrainKV/common"
	"github.com/kebukeYi/TrainKV/lsm"
	"github.com/kebukeYi/TrainKV/model"
	"log"
)

type TaskStoreManger struct {
	path string
	db   *TrainKV.TrainKV
}

func OpenTaskStoreManger(path string) *TaskStoreManger {
	trainKVDB, err, _ := TrainKV.Open(lsm.GetLSMDefaultOpt(path))
	if err != nil {
		log.Fatalln(" TaskStoreManger Open trainKVDB fail,", err)
	}
	return &TaskStoreManger{
		db:   trainKVDB,
		path: path,
	}
}

func (m *TaskStoreManger) PutReplications(key string, value []*Replication) error {
	var data []byte
	var err error
	if data, err = replications2bytes(value); err != nil {
		fmt.Printf("PutReplications(%s).replications2bytes encode,error:%s \n", key, err)
		return err
	}
	if data != nil {
		entry := model.NewEntry([]byte(key), data)
		err = m.db.Set(entry)
		if err != nil {
			return err
		}
	}
	return err
}

func (m *TaskStoreManger) GetReplications(key string) ([]*Replication, error) {
	data, err := m.db.Get([]byte(key))
	if err != nil || data.Version == -1 {
		if errors.Is(err, DBcommon.ErrKeyNotFound) {
			return make([]*Replication, 0), nil
		}
		return nil, err
	}
	if data.Value == nil || len(data.Value) == 0 {
		return make([]*Replication, 0), nil
	}
	var m2 []*Replication
	if m2, err = bytes2Replications(data.Value); err != nil {
		fmt.Printf("GetReplications(%s).bytes2Replications decode,error:%s \n", key, err)
		return nil, err
	}
	if m2 != nil {
		return m2, nil
	}
	return nil, err
}

func (m *TaskStoreManger) PutTrashes(key string, value []string) error {
	var data []byte
	var err error
	if data, err = strings2bytes(value); err != nil {
		fmt.Printf("PutTrashes(%s).strings2bytes encode,error:%s \n", key, err)
		return err
	}
	if data != nil {
		entry := model.NewEntry([]byte(key), data)
		err = m.db.Set(entry)
		if err != nil {
			return err
		}
	}
	return err
}

func (m *TaskStoreManger) GetTrashes(key string) ([]string, error) {
	data, err := m.db.Get([]byte(key))
	if err != nil || data == nil {
		if errors.Is(err, DBcommon.ErrKeyNotFound) {
			return make([]string, 0), nil
		}
		return nil, err
	}
	if data.Value == nil || len(data.Value) == 0 {
		return make([]string, 0), nil
	}
	var m2 []string
	if m2, err = bytes2Strings(data.Value); err != nil {
		fmt.Printf("GetTrashes(%s).bytes2Strings decode,error:%s \n", key, err)
		return nil, err
	}
	if m2 != nil {
		return m2, nil
	}
	return nil, err
}

func (m *TaskStoreManger) Delete(key string) error {
	err := m.db.Del([]byte(key))
	if err != nil {
		if errors.Is(err, DBcommon.ErrKeyNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func (m *TaskStoreManger) close() error {
	if m.db != nil {
		err := m.db.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func bytes2Replications(value []byte) ([]*Replication, error) {
	if value == nil {
		return nil, common.ErrInputEmpty
	}
	decoder := gob.NewDecoder(bytes.NewBuffer(value))
	buf := make([]*Replication, 0)
	err := decoder.Decode(&value)
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func replications2bytes(value []*Replication) ([]byte, error) {
	if value == nil {
		return nil, common.ErrInputEmpty
	}
	buff := make([]byte, 0)
	buf := bytes.NewBuffer(buff)
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(value)
	if err != nil {
		return nil, err
	}
	return buff, nil
}

func bytes2Strings(value []byte) ([]string, error) {
	if value == nil {
		return nil, common.ErrInputEmpty
	}
	var strings []string
	err := json.Unmarshal(value, &strings)
	if err != nil {
		return nil, err
	}
	return strings, nil
}

func strings2bytes(value []string) ([]byte, error) {
	if value == nil {
		return nil, common.ErrInputEmpty
	}
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return jsonBytes, nil
}
