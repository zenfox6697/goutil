package redis

import (
	"encoding/json"
	"errors"
	"log"
	"runtime"
	"strings"
	"time"

	redigo "github.com/garyburd/redigo/redis"
)

var RedisPool map[string]*redigo.Pool

//初始化redis 1为多个连接池，2为共用一个连接池
func InitRedis(addrs string) {
	InitRedisBySize(addrs, 80, 10, 240)
}

func InitRedisBySize(addrs string, maxSize, maxIdle, timeout int) {
	RedisPool = map[string]*redigo.Pool{}
	addr := strings.Split(addrs, ",")
	for _, v := range addr {
		saddr := strings.Split(v, "=")
		RedisPool[saddr[0]] = &redigo.Pool{MaxActive: maxSize, MaxIdle: maxIdle,
			IdleTimeout: time.Duration(timeout) * time.Second, Dial: func() (redigo.Conn, error) {
				c, err := redigo.Dial("tcp", saddr[1])
				if err != nil {
					return nil, err
				}
				return c, nil
			}}
	}
}

//分流redis
//并存入字符串缓存
func PutKV(key string, obj interface{}) bool {
	return Put("other", key, obj, -1)
}
func PutCKV(code, key string, obj interface{}) bool {
	return Put(code, key, obj, -1)
}
func Put(code, key string, obj interface{}, timeout int) bool {
	b := false
	defer catch()
	conn := RedisPool[code].Get()
	defer conn.Close()
	var err error
	_obj, _err := json.Marshal(obj)
	if _err != nil {
		log.Println("redisutil-SET-序列化出错Error", _err)
		return b
	}
	if timeout < 1 {
		_, err = conn.Do("SET", key, _obj)
	} else {
		_, err = conn.Do("SET", key, _obj, "EX", timeout)
	}
	if nil != err {
		log.Println("redisutil-SETError-put", err)
	} else {
		b = true
	}
	return b
}

func BulkPut(code string, timeout int, obj ...interface{}) bool {
	b := false
	defer catch()
	conn := RedisPool[code].Get()
	defer conn.Close()
	var err error
	for _, _tmp := range obj {
		tmp, ok := _tmp.([]interface{})
		if ok && len(tmp) == 2 {
			key, kok := tmp[0].(string)
			if kok && key != "" {
				_obj, _err := json.Marshal(tmp[1])
				if _err != nil {
					log.Println("redisutil-SET-序列化出错Error", _err)
					return b
				}
				if timeout < 1 {
					_, err = conn.Do("SET", key, _obj)
				} else {
					_, err = conn.Do("SET", key, _obj, "EX", timeout)
				}
			}
		}
	}
	if nil != err {
		b = false
		log.Println("redisutil-SETError-put", err)
	} else {
		b = b && true
	}
	return b
}

//直接存字节流
func PutBytes(code, key string, data *[]byte, timeout int) (err error) {
	defer catch()

	conn := RedisPool[code].Get()
	defer conn.Close()

	if timeout < 1 {
		_, err = conn.Do("SET", key, *data)
	} else {
		_, err = conn.Do("SET", key, *data, "EX", timeout)
	}
	if nil != err {
		log.Println("redisutil-SETError", err)
	}
	return
}

//设置超时时间,单位秒
func SetExpire(code, key string, expire int) error {
	defer catch()

	conn := RedisPool[code].Get()
	defer conn.Close()
	_, err := conn.Do("expire", key, expire)
	return err
}

//判断一个key是否存在
func Exists(code, key string) (bool, error) {
	defer catch()

	conn := RedisPool[code].Get()
	defer conn.Close()
	repl, err := conn.Do("exists", key)
	ret, _ := redigo.Int(repl, err)
	return ret == 1, err
}

//获取string
func GetStr(code, key string) string {
	res := Get(code, key)
	str, _ := res.(string)
	return str
}

//获取int
func GetInt(code, key string) int {
	result, _ := GetNewInt(code, key)
	return result
}

func GetNewInt(code, key string) (int, error) {
	var res interface{}
	err := GetNewInterface(code, key, &res)
	var result int
	if str, ok := res.(float64); ok {
		result = int(str)
	}
	return result, err
}

//取得字符串,支持变参，2个 (key,code)，返回后自己断言
func Get(code, key string) (result interface{}) {
	GetInterface(code, key, &result)
	return
}

func GetInterface(code, key string, result interface{}) {
	GetNewInterface(code, key, result)
}

func GetNewInterface(code, key string, result interface{}) error {
	defer catch()
	conn := RedisPool[code].Get()
	defer conn.Close()
	ret, err := conn.Do("GET", key)
	if nil != err {
		log.Println("redisutil-GetError", err)
	} else {
		var ok bool
		var res []byte
		if res, ok = ret.([]byte); ok {
			err = json.Unmarshal(res, result)
			if err != nil {
				log.Println("Get ERROR:", err.Error())
			}
		}
	}
	return err
}

//直接返回字节流
func GetBytes(code, key string) (ret *[]byte, err error) {
	defer catch()

	conn := RedisPool[code].Get()
	defer conn.Close()
	var r interface{}
	r, err = conn.Do("GET", key)
	if err != nil {
		log.Println("redisutil-GetBytesError", err)
	} else {
		if tmp, ok := r.([]byte); ok {
			ret = &tmp
		} else {
			err = errors.New("redis返回数据格式不对")
		}
	}
	return
}
func GetNewBytes(code, key string) (ret *[]byte, err error) {
	defer catch()
	redisPool := RedisPool[code]
	if redisPool == nil {
		err = errors.New("redis code " + code + " is nil")
		log.Println("redisutil-GetNewBytesError", err)
		return
	}
	conn := redisPool.Get()
	defer conn.Close()
	var r interface{}
	r, err = conn.Do("GET", key)
	if err != nil {
		log.Println("redisutil-GetNewBytesError", err)
	} else if r != nil {
		if tmp, ok := r.([]byte); ok {
			ret = &tmp
		}
	}
	return
}

//删所有key
func FlushDB(code string) bool {
	b := false
	defer catch()

	conn := RedisPool[code].Get()
	defer conn.Close()

	var err error
	_, err = conn.Do("FLUSHDB")
	if nil != err {
		log.Println("redisutil-FLUSHDBError", err)
	} else {
		b = true
	}
	return b
}

//支持删除多个key
func Del(code string, key ...interface{}) bool {
	defer catch()
	b := false
	conn := RedisPool[code].Get()
	defer conn.Close()

	var err error
	_, err = conn.Do("DEL", key...)
	if nil != err {
		log.Println("redisutil-DELError", err)
	} else {
		b = true
	}
	return b
}

/**
func DelKey(key ...interface{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("[E]", r)
			for skip := 1; ; skip++ {
				_, file, line, ok := runtime.Caller(skip)
				if !ok {
					break
				}
				go log.Printf("%v,%v\n", file, line)
			}
		}
	}()
	for i := 0; i < len(RedisPool); i++ {
		delByNum(i, key...)
	}

}
**/

/**
func delByNum(n int, key ...interface{}) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("[E]", r)
			for skip := 1; ; skip++ {
				_, file, line, ok := runtime.Caller(skip)
				if !ok {
					break
				}
				go log.Printf("%v,%v\n", file, line)
			}
		}
	}()
	i := 0
	for _, v := range RedisPool {
		if i == n {
			conn := v.Get()
			defer conn.Close()
			conn.Do("DEL", key...)
			break
		}
		i++
	}
}
**/

//根据代码和前辍key删除多个
func DelByCodePattern(code, key string) {
	defer catch()

	conn := RedisPool[code].Get()
	defer conn.Close()
	ret, err := conn.Do("KEYS", key)
	var result []interface{}
	if nil != err {
		log.Println("redisutil-GetError", err)
	} else {
		result = ret.([]interface{})
		for k := 0; k < len(result); k++ {
			conn.Do("DEL", string(result[k].([]uint8)))
		}
	}
}

/**
func DelByPattern(key string) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("[E]", r)
			for skip := 1; ; skip++ {
				_, file, line, ok := runtime.Caller(skip)
				if !ok {
					break
				}
				go log.Printf("%v,%v\n", file, line)
			}
		}
	}()
	i := 0
	for _, v := range RedisPool {
		conn := v.Get()
		defer conn.Close()
		ret, err := conn.Do("KEYS", key)
		var result []interface{}
		if nil != err {
			log.Println("redisutil-GetError", err)
		} else {
			result = ret.([]interface{})
			for k := 0; k < len(result); k++ {
				delByNum(i, string(result[k].([]uint8)))
			}
		}
		i++
	}

}
**/
//自增计数器
func Incr(code, key string) int64 {
	ret, err := IncrByErr(code, key)
	if nil != err {
		log.Println("redisutil-INCR-Error", err)
	}
	return ret
}

//自增计数器
func IncrByErr(code, key string) (int64, error) {
	defer catch()
	conn := RedisPool[code].Get()
	defer conn.Close()
	ret, err := conn.Do("INCR", key)
	if nil != err {
		return 0, err
	}
	if res, ok := ret.(int64); ok {
		return res, nil
	} else {
		return 0, nil
	}
}

//自减
func Decrby(code, key string, val int) int64 {
	defer catch()
	conn := RedisPool[code].Get()
	defer conn.Close()
	ret, err := conn.Do("DECRBY", key, val)
	if nil != err {
		log.Println("redisutil-DECR-Error", err)
	} else {
		if res, ok := ret.(int64); ok {
			return res
		} else {
			return 0
		}
	}
	return 0
}

//根据正则去取
func GetKeysByPattern(code, key string) []interface{} {
	defer catch()
	conn := RedisPool[code].Get()
	defer conn.Close()
	ret, err := conn.Do("KEYS", key)
	if nil != err {
		log.Println("redisutil-GetKeysError", err)
		return nil
	} else {
		res, _ := ret.([]interface{})
		return res
	}
}

//批量取多个key
func Mget(code string, key []string) []interface{} {
	defer catch()
	conn := RedisPool[code].Get()
	defer conn.Close()
	interfaceKeys := make([]interface{}, len(key))
	for n, k := range key {
		interfaceKeys[n] = k
	}
	ret, err := conn.Do("MGET", interfaceKeys...)
	if nil != err {
		log.Println("redisutil-MgetError", err)
		return nil
	} else {
		res, _ := ret.([]interface{})
		return res
	}
}

//取出并删除Key
func Pop(code string, key string) (result interface{}) {
	defer catch()
	conn := RedisPool[code].Get()
	defer conn.Close()
	ret, err := conn.Do("GET", key)
	if nil != err {
		log.Println("redisutil-PopError", err)
	} else {
		var ok bool
		var res []byte
		if res, ok = ret.([]byte); ok {
			err = json.Unmarshal(res, &result)
			if err != nil {
				log.Println("Poperr", err)
			}
		}
		conn.Do("DEL", key)
	}
	return
}

//list操作
func LPOP(code, list string) (result interface{}) {
	defer catch()
	conn := RedisPool[code].Get()
	defer conn.Close()
	ret, err := conn.Do("LPOP", list)
	if nil != err {
		log.Println("redisutil-LPopError", err)
	} else {
		if res, ok := ret.([]byte); ok {
			err = json.Unmarshal(res, &result)
			log.Println(err)
		}
	}
	return
}

func RPUSH(code, list string, val interface{}) bool {
	defer catch()
	conn := RedisPool[code].Get()
	defer conn.Close()
	_obj, _ := json.Marshal(val)
	_, err := conn.Do("RPUSH", list, _obj)
	if nil != err {
		log.Println("redisutil-RPUSHError", err)
		return false
	}
	return true
}

func LLEN(code, list string) int64 {
	defer catch()
	conn := RedisPool[code].Get()
	defer conn.Close()
	ret, err := conn.Do("LLEN", list)
	if nil != err {
		log.Println("redisutil-LLENError", err)
		return 0
	}
	if res, ok := ret.(int64); ok {
		return res
	} else {
		return 0
	}
}

func catch() {
	if r := recover(); r != nil {
		log.Println(r)
		for skip := 0; ; skip++ {
			_, file, line, ok := runtime.Caller(skip)
			if !ok {
				break
			}
			go log.Printf("%v,%v\n", file, line)
		}
	}
}
