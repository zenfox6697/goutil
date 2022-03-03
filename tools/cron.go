package tools

import (
	"log"
	"strings"
	"time"
)

func SimpleCrontab(flag bool, c string, f func()) {
	array := strings.Split(c, ":")
	if len(array) != 2 {
		log.Fatalln("定时任务参数错误!", c)
	}
	if flag {
		go f()
	}
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), IntAllDef(array[0], 0), IntAllDef(array[1], 0), 0, 0, time.Local)
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
