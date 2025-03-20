package service

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/kebukeYi/TrainDB"
	DBCommon "github.com/kebukeYi/TrainDB/common"
	"github.com/kebukeYi/TrainDB/lsm"
	"github.com/kebukeYi/TrainDB/model"
	"github.com/kebukeYi/TrainFS/common"
	"log"
)

type MetaStoreManger struct {
	path string
	db   *TrainDB.TrainKVDB
}

func OpenDataStoreManger(path string) *MetaStoreManger {
	trainKVDB, err, _ := TrainDB.Open(lsm.GetLSMDefaultOpt(path))
	if err != nil {
		log.Fatalln(" DataStoreManger Open trainKVDB fail,", err)
	}
	return &MetaStoreManger{
		db:   trainKVDB,
		path: path,
	}
}

func (m *MetaStoreManger) PutFileMeta(key string, value *FileMeta) error {
	var data2Bytes []byte
	var err error
	if data2Bytes, err = m.data2Bytes(value); err != nil {
		fmt.Printf("PutFileMeta(%s).bytes2FileMeta failed,err:%v \n", key, err)
		return err
	}
	entry := model.Entry{Key: []byte(key), Value: data2Bytes}
	err = m.db.Set(&entry)
	if err != nil {
		fmt.Printf("PutFileMeta(%s).db.Set() failed; err: %v", key, err)
		return err
	}
	return nil
}

func (m *MetaStoreManger) GetFileMeta(key string) (*FileMeta, error) {
	value, err := m.db.Get([]byte(key))
	if err != nil || value == nil {
		if errors.Is(err, DBCommon.ErrKeyNotFound) {
			return nil, common.ErrFileNotFound
		}
		return nil, err
	}
	if value.Value == nil || len(value.Value) == 0 {
		return nil, common.ErrFileNotFound
	}
	fileMeta, err := m.bytes2FileMeta(value.Value)
	if err != nil {
		fmt.Printf("bytes2FileMeta failed, key:%s; len(val):%d; err:%v \n", key, len(value.Value), err)
		return nil, err
	}
	return fileMeta, nil
}

func (m *MetaStoreManger) Delete(key string) error {
	if err := m.db.Del([]byte(key)); err != nil {
		if errors.Is(err, DBCommon.ErrKeyNotFound) {
			return nil
		}
		return err
	}
	return nil
}

func (m *MetaStoreManger) close() error {
	if m.db != nil {
		err := m.db.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MetaStoreManger) data2Bytes(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (m *MetaStoreManger) bytes2FileMeta(data []byte) (*FileMeta, error) {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	var fileMeta *FileMeta
	err := decoder.Decode(&fileMeta)
	if err != nil {
		return nil, err
	}
	return fileMeta, nil
}
