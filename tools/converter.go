package tools

import (
	"fmt"
	"log"
	"math/big"
	"runtime"
	"strconv"
	"time"
)

func IntAll(num interface{}) int {
	return IntAllDef(num, 0)
}

func Int64All(num interface{}) int64 {
	if i, ok := num.(int64); ok {
		return int64(i)
	} else if i0, ok0 := num.(int32); ok0 {
		return int64(i0)
	} else if i1, ok1 := num.(float64); ok1 {
		return int64(i1)
	} else if i2, ok2 := num.(int); ok2 {
		return int64(i2)
	} else if i3, ok3 := num.(float32); ok3 {
		return int64(i3)
	} else if i4, ok4 := num.(string); ok4 {
		i64, _ := strconv.ParseInt(i4, 10, 64)
		//in, _ := strconv.Atoi(i4)
		return i64
	} else if i5, ok5 := num.(int16); ok5 {
		return int64(i5)
	} else if i6, ok6 := num.(int8); ok6 {
		return int64(i6)
	} else if i7, ok7 := num.(*big.Int); ok7 {
		in, _ := strconv.ParseInt(fmt.Sprint(i7), 10, 64)
		return int64(in)
	} else if i8, ok8 := num.(*big.Float); ok8 {
		in, _ := strconv.ParseInt(fmt.Sprint(i8), 10, 64)
		return int64(in)
	} else {
		return 0
	}
}

func Float64All(num interface{}) float64 {
	if i, ok := num.(float64); ok {
		return float64(i)
	} else if i0, ok0 := num.(int32); ok0 {
		return float64(i0)
	} else if i1, ok1 := num.(int64); ok1 {
		return float64(i1)
	} else if i2, ok2 := num.(int); ok2 {
		return float64(i2)
	} else if i3, ok3 := num.(float32); ok3 {
		return float64(i3)
	} else if i4, ok4 := num.(string); ok4 {
		in, _ := strconv.ParseFloat(i4, 64)
		return in
	} else if i5, ok5 := num.(int16); ok5 {
		return float64(i5)
	} else if i6, ok6 := num.(int8); ok6 {
		return float64(i6)
	} else if i6, ok6 := num.(uint); ok6 {
		return float64(i6)
	} else if i6, ok6 := num.(uint8); ok6 {
		return float64(i6)
	} else if i6, ok6 := num.(uint16); ok6 {
		return float64(i6)
	} else if i6, ok6 := num.(uint32); ok6 {
		return float64(i6)
	} else if i6, ok6 := num.(uint64); ok6 {
		return float64(i6)
	} else if i7, ok7 := num.(*big.Float); ok7 {
		in, _ := strconv.ParseFloat(fmt.Sprint(i7), 64)
		return float64(in)
	} else if i8, ok8 := num.(*big.Int); ok8 {
		in, _ := strconv.ParseFloat(fmt.Sprint(i8), 64)
		return float64(in)
	} else {
		return 0
	}
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

func ObjToString(old interface{}) string {
	if nil == old {
		return ""
	} else {
		r, _ := old.(string)
		return r
	}
}

func ObjToStringDef(old interface{}, defaultstr string) string {
	if nil == old {
		return defaultstr
	} else {
		r, _ := old.(string)
		if r == "" {
			return defaultstr
		}
		return r
	}
}

//对象数组转成string数组
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

//对象数组转成map数组
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

//map数组转成对象数组
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

func SubstrByByte(str string, length int) string {
	bs := []byte(str)[:length]
	bl := 0
	for i := len(bs) - 1; i >= 0; i-- {
		switch {
		case bs[i] >= 0 && bs[i] <= 127:
			return string(bs[:i+1])
		case bs[i] >= 128 && bs[i] <= 191:
			bl++
		case bs[i] >= 192 && bs[i] <= 253:
			cl := 0
			switch {
			case bs[i]&252 == 252:
				cl = 6
			case bs[i]&248 == 248:
				cl = 5
			case bs[i]&240 == 240:
				cl = 4
			case bs[i]&224 == 224:
				cl = 3
			default:
				cl = 2
			}
			if bl+1 == cl {
				return string(bs[:i+cl])
			}
			return string(bs[:i])
		}
	}
	return ""
}

func SubString(str string, begin, length int) (substr string) {
	// 将字符串的转换成[]rune
	rs := []rune(str)
	lth := len(rs)
	// 简单的越界判断
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length
	if end > lth {
		end = lth
	}

	// 返回子串
	return string(rs[begin:end])
}

//捕获异常
func Try(fun func(), handler func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			for skip := 1; ; skip++ {
				_, file, line, ok := runtime.Caller(skip)
				if !ok {
					break
				}
				go log.Printf("%v,%v\n", file, line)
			}
			handler(err)
		}
	}()
	fun()
}

//3目运算
func If(b bool, to, fo interface{}) interface{} {
	if b {
		return to
	} else {
		return fo
	}
}

//HashCode值
func HashCode(uid string) int {
	var h uint32 = 0
	rs := []rune(uid)
	for i := 0; i < len(rs); i++ {
		h = 31*h + uint32(rs[i])
	}
	return int(h)
}

//获取离n天的秒差
func GetDayStartSecond(n int) int64 {
	now := time.Now()
	tom := time.Date(now.Year(), now.Month(), now.Day()+n, 0, 0, 0, 0, time.Local)
	return tom.Unix()
}

func InterfaceArrTointArr(arr []interface{}) []int {
	tmp := make([]int, 0)
	for _, v := range arr {
		tmp = append(tmp, int(v.(float64)))
	}
	return tmp
}
func InterfaceArrToint64Arr(arr []interface{}) []int64 {
	tmp := make([]int64, 0)
	for _, v := range arr {
		tmp = append(tmp, int64(v.(float64)))
	}
	return tmp
}
