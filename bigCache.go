package BigCache

import (
	"bytes"
	"encoding/gob"
	"github.com/allegro/bigcache"
	"log"
	"time"
)

type Cache interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, error)
}

type BigCache struct {
	bc *bigcache.BigCache
}

var c *BigCache

func NewInstance() *BigCache {
	if c == nil {
		cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(10 * time.Minute))
		if err != nil {
			log.Fatal("new bigCache err : ", err)
		}
		c = &BigCache{
			bc: cache,
		}
	}
	return c
}

func (c *BigCache) Set(key string, value interface{}) error {
	// 将 value 序列化为 bytes
	valueBytes, err := serialize(value)
	if err != nil {
		return err
	}
	return c.bc.Set(key, valueBytes)
}

func (c *BigCache) Get(key string) (interface{}, error) {
	valueBytes, err := c.bc.Get(key)
	if err != nil {
		return nil, err
	}

	// 反序列化 valueBytes
	value, err := deserialize(valueBytes)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func serialize(value interface{}) ([]byte, error) {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)

	gob.Register(value)

	err := enc.Encode(&value)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func deserialize(valueBytes []byte) (interface{}, error) {
	var value interface{}

	buf := bytes.NewBuffer(valueBytes)
	dec := gob.NewDecoder(buf)

	err := dec.Decode(&value)
	if err != nil {
		return nil, err
	}
	return value, nil
}
