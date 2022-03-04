package goutil

import (
	"fmt"
	"log"
	"math/big"
	mathRand "math/rand"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	tmp = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ12345678900"
)

func WriteFile(loc string, content []byte) {
	p, _ := path.Split(loc)
	os.MkdirAll(p, 0644)
	fh, err := os.Create(loc)
	if err != nil {
		log.Println(err)
	}
	fh.Write(content)
}

func Abort(funcname string, err error) {
	panic(funcname + " failed: " + err.Error())
}

func Uuid(length int) string {
	ret := []string{}
	r := mathRand.New(mathRand.NewSource(time.Now().UnixNano()))
	for i := 0; i < length; i++ {
		index := r.Intn(62)
		ret = append(ret, tmp[index:index+1])
	}
	return strings.Join(ret, "")
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func RandString(n int) string {
	mathRand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[mathRand.Intn(len(letterRunes))]
	}
	return string(b)
}

func LongToDate(date interface{}, flag bool) string {
	var int64Date int64
	if l1, ok1 := date.(float64); ok1 {
		int64Date = int64(l1)
	} else if l2, ok2 := date.(int64); ok2 {
		int64Date = l2
	} else if l3, ok3 := date.(int); ok3 {
		int64Date = int64(l3)
	}
	t := time.Unix(int64Date, 0)
	if flag {
		return t.Format("2006-01-02 15:04:05")
	} else {
		return t.Format("2006-01-02")
	}
}

func ObjToString(old interface{}) string {
	if nil == old {
		return ""
	} else {
		r, _ := old.(string)
		return r
	}
}

func ObjArrToStringArr(old []interface{}) []string {
	if old != nil {
		new := make([]string, len(old))
		for i, v := range old {
			new[i], _ = v.(string)
		}
		return new
	} else {
		return nil
	}
}

func ObjArrToMapArr(old []interface{}) []map[string]interface{} {
	if old != nil {
		new := make([]map[string]interface{}, len(old))
		for i, v := range old {
			new[i], _ = v.(map[string]interface{})
		}
		return new
	} else {
		return nil
	}
}

func MapArrToObjArr(old []map[string]interface{}) []interface{} {
	if old != nil {
		new := make([]interface{}, len(old))
		for i, v := range old {
			new[i] = v
		}
		return new
	} else {
		return nil
	}
}

func Catch() {
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

func IntAll(num interface{}) int {
	return IntAllDef(num, 0)
}

func IntAllDef(num interface{}, defaultNum int) int {
	if i, ok := num.(int); ok {
		return int(i)
	} else if i0, ok0 := num.(int32); ok0 {
		return int(i0)
	} else if i1, ok1 := num.(float64); ok1 {
		return int(i1)
	} else if i2, ok2 := num.(int64); ok2 {
		return int(i2)
	} else if i3, ok3 := num.(float32); ok3 {
		return int(i3)
	} else if i4, ok4 := num.(string); ok4 {
		in, _ := strconv.Atoi(i4)
		return int(in)
	} else if i5, ok5 := num.(int16); ok5 {
		return int(i5)
	} else if i6, ok6 := num.(int8); ok6 {
		return int(i6)
	} else if i7, ok7 := num.(*big.Int); ok7 {
		in, _ := strconv.Atoi(fmt.Sprint(i7))
		return int(in)
	} else if i8, ok8 := num.(*big.Float); ok8 {
		in, _ := strconv.Atoi(fmt.Sprint(i8))
		return int(in)
	} else {
		return defaultNum
	}
}

func SimpleCrontab(flag bool, c string, f func()) {
	array := strings.Split(c, ":")
	if len(array) != 2 {
		log.Fatalln("定时任务参数错误!", c)
	}
	if flag {
		go f()
	}
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), IntAll(array[0]), IntAll(array[1]), 0, 0, time.Local)
	if t.Before(now) {
		t = t.AddDate(0, 0, 1)
	}
	sub := t.Sub(now)
	log.Println(c, "run after", sub)
	timer := time.NewTimer(sub)
	for {
		select {
		case <-timer.C:
			go f()
			log.Println(c, "run after", 24*time.Hour)
			timer.Reset(24 * time.Hour)
		}
	}
}
