// beta:leveldb v0.1
package leveldb

import (
	"log"

	"github.com/syndtr/goleveldb/leveldb"
)

type level struct {
	Path string
	Key  string
	DB   *leveldb.DB
}

func NewLevelDB(path, key string) *level {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		log.Println(err)
	}
	return &level{Path: path, DB: db, Key: key}
}

func (l *level) PutString(k, v string) {
	err := l.DB.Put([]byte(l.Key+k), []byte(v), nil)
	if err != nil {
		log.Println(err)
	}
}

func (l *level) GetString(k string) string {
	v, err := l.DB.Get([]byte(k), nil)
	if err != nil {
		log.Println(err)
	}
	return string(v)
}
