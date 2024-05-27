package service

import (
	"bytes"
	"encoding/gob"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
)

type DataStoreManger struct {
	path string
	db   *leveldb.DB
}

func OpenDataStoreManger(path string) *DataStoreManger {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		log.Fatalln(" NameNode Open level.db file fail,", err)
	}
	return &DataStoreManger{
		db:   db,
		path: path,
	}
}

func (m *DataStoreManger) PutFileMeta(key string, value *FileMeta) error {
	data2Bytes := m.data2Bytes(value)
	err := m.db.Put([]byte(key), data2Bytes, nil)
	if err != nil {
		log.Fatalf("PutFileMeta() failed, %v", err)
		return err
	}
	return nil
}

func (m *DataStoreManger) GetFileMeta(key string) (*FileMeta, error) {
	value, err := m.db.Get([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	fileMeta, err := m.bytes2FileMeta(value)
	if err != nil {
		return nil, err
	}
	return fileMeta, nil
}

func (m *DataStoreManger) Delete(key string) error {
	err := m.db.Delete([]byte(key), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil
		}
		return err
	}
	return nil
}

func (m *DataStoreManger) PutDataNodeMeta(key string, value map[string]*DataNodeInfo) {
	data2Bytes := m.data2Bytes(value)
	err := m.db.Put([]byte(key), data2Bytes, nil)
	if err != nil {
		log.Fatalf("bytes2FileMeta failed, %v", err)
	}
}

func (m *DataStoreManger) GetDataNodeMetas(key string) (map[string]*DataNodeInfo, error) {
	value, err := m.db.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}
	dataNodeMetas, err := m.bytes2DataMeta(value)
	if err != nil {
		return nil, err
	}
	return dataNodeMetas, nil

}

func (m *DataStoreManger) data2Bytes(data interface{}) []byte {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	err := encoder.Encode(data)
	if err != nil {
		log.Fatalf("data2Bytes failed, %v", err)
	}
	return buf.Bytes()
}

func (m *DataStoreManger) bytes2FileMeta(data []byte) (*FileMeta, error) {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	var fileMeta *FileMeta
	err := decoder.Decode(&fileMeta)
	if err != nil {
		log.Fatalf("bytes2FileMeta failed, %v", err)
	}
	return fileMeta, nil
}

func (m *DataStoreManger) bytes2DataMeta(data []byte) (map[string]*DataNodeInfo, error) {
	buffer := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(buffer)
	var dataMeta map[string]*DataNodeInfo
	err := decoder.Decode(&dataMeta)
	if err != nil {
		log.Fatalf("bytes2DataMeta failed, %v", err)
	}
	return dataMeta, nil
}
