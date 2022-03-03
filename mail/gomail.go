package mail

import (
	"crypto/tls"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/go-gomail/gomail"
)

type GmailAuth struct {
	SmtpHost string //邮箱服务器
	SmtpPort int    //邮箱端口
	User     string //用户
	Pwd      string //密码
	PoolChan chan *gomail.Dialer
	PoolSize int
	ReTry    int
}

var locker = &sync.Mutex{}

func getDialer(flag bool, auth *GmailAuth, to string) *gomail.Dialer {
	if auth.PoolChan == nil {
		locker.Lock()
		defer locker.Unlock()
		if auth.PoolChan == nil {
			if auth.PoolSize == 0 {
				auth.PoolSize = 1
			}
			auth.PoolChan = make(chan *gomail.Dialer, auth.PoolSize)
			for i := 0; i < auth.PoolSize; i++ {
				// dialer := gomail.NewPlainDialer(auth.SmtpHost, auth.SmtpPort, auth.User, auth.Pwd) // 发送邮件服务器、端口、发件人账号、发件人密码
				dialer := gomail.NewDialer(auth.SmtpHost, auth.SmtpPort, auth.User, auth.Pwd) // 发送邮件服务器、端口、发件人账号、发件人密码
				dialer.TLSConfig = &tls.Config{ServerName: auth.SmtpHost, InsecureSkipVerify: true}
				auth.PoolChan <- dialer
			}
		}
	}
	if flag {
		log.Println(auth.User, to, "发送邮件getDialer：get new gomail Dialer")
		// dialer := gomail.NewPlainDialer(auth.SmtpHost, auth.SmtpPort, auth.User, auth.Pwd) // 发送邮件服务器、端口、发件人账号、发件人密码
		dialer := gomail.NewDialer(auth.SmtpHost, auth.SmtpPort, auth.User, auth.Pwd) // 发送邮件服务器、端口、发件人账号、发件人密码
		dialer.TLSConfig = &tls.Config{ServerName: auth.SmtpHost, InsecureSkipVerify: true}
		auth.PoolChan <- dialer
	}
	return <-auth.PoolChan
}

//发送普通
func GSendMail(from, to, cc, bcc, subject, body, fname, rename string, auth *GmailAuth) bool {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", auth.User, from) // 发件人
	m.SetHeader("To",
		m.FormatAddress(to, "收件人")) // 收件人
	if cc != "" {
		m.SetHeader("Cc", m.FormatAddress(cc, "收件人")) //抄送
	}
	if bcc != "" {
		m.SetHeader("Bcc", m.FormatAddress(bcc, "收件人")) // 暗送
	}
	m.SetHeader("Subject", subject) // 主题
	m.SetBody("text/html", body)    // 正文
	if fname != "" {
		h := map[string][]string{"Content-Type": {"text/plain; charset=UTF-8"}}
		m.Attach(fname, gomail.Rename(rename), gomail.SetHeader(h)) //添加附件
		//m.Attach(fname) //添加附件
	}
	reTry := auth.ReTry
	if reTry == 0 {
		reTry = 3
	}
	return gSend(reTry, auth, m, to)
}

//如果附件是byte，用这个
func GSendMail_B(from, to, cc, bcc, subject, body, fname string, fb []byte, auth *GmailAuth) bool {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", auth.User, from) // 发件人
	m.SetHeader("To",
		m.FormatAddress(to, "收件人")) // 收件人
	if cc != "" {
		m.SetHeader("Cc", m.FormatAddress(cc, "收件人")) //抄送
	}
	if bcc != "" {
		m.SetHeader("Bcc", m.FormatAddress(bcc, "收件人")) // 暗送
	}
	m.SetHeader("Subject", subject) // 主题
	m.SetBody("text/html", body)    // 正文
	if fname != "" {
		// h := map[string][]string{"Content-Type": {"text/plain; charset=UTF-8"}}
		// m.Attach_new(fb, gomail.Rename(fname), gomail.SetHeader(h)) //添加附件
		m.Attach(fname) //添加附件
	}
	reTry := auth.ReTry
	if reTry == 0 {
		reTry = 1
	}
	return gSend(reTry, auth, m, to)
}

func GSendMail_Bq(from, to, cc, bcc, subject, body, fname string, fb []byte, auth *GmailAuth) bool {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", auth.User, from) // 发件人
	tos := strings.Split(to, "|")
	if len(tos) > 0 {
		tos1 := strings.Split(tos[0], ",")
		m.SetHeader("To", tos1...) // 收件人
	}
	if len(tos) > 1 {
		tos2 := strings.Split(tos[1], ",")
		if cc != "" {
			tos2 = append(tos2, cc)
		}
		m.SetHeader("Cc", tos2...) // 收件人
	} else {
		if cc != "" {
			m.SetHeader("Cc", m.FormatAddress(cc, "收件人")) //抄送
		}
	}
	if len(tos) > 2 {
		tos3 := strings.Split(tos[2], ",")
		if bcc != "" {
			tos3 = append(tos3, cc)
		}
		m.SetHeader("Bcc", tos3...) // 收件人
	} else {
		if bcc != "" {
			m.SetHeader("Bcc", m.FormatAddress(bcc, "收件人")) // 暗送
		}
	}
	m.SetHeader("Subject", subject) // 主题
	m.SetBody("text/html", body)    // 正文
	if fname != "" {
		// h := map[string][]string{"Content-Type": {"text/plain; charset=UTF-8"}}
		// m.Attach_new(fb, gomail.Rename(fname), gomail.SetHeader(h)) //添加附件
		m.Attach(fname) //添加附件
	}
	reTry := auth.ReTry
	if reTry == 0 {
		reTry = 1
	}
	return gSend(reTry, auth, m, to)
}

func GSendMail_q(from, to, cc, bcc, subject, body, fname, rename string, auth *GmailAuth) bool {
	m := gomail.NewMessage()
	m.SetAddressHeader("From", auth.User, from) // 发件人
	tos := strings.Split(to, "|")
	if len(tos) > 0 {
		tos1 := strings.Split(tos[0], ",")
		m.SetHeader("To", tos1...) // 收件人
	}
	if len(tos) > 1 {
		tos2 := strings.Split(tos[1], ",")
		if cc != "" {
			tos2 = append(tos2, cc)
		}
		m.SetHeader("Cc", tos2...) // 收件人
	} else {
		if cc != "" {
			m.SetHeader("Cc", m.FormatAddress(cc, "收件人")) //抄送
		}
	}
	if len(tos) > 2 {
		tos3 := strings.Split(tos[2], ",")
		if bcc != "" {
			tos3 = append(tos3, cc)
		}
		m.SetHeader("Bcc", tos3...) // 收件人
	} else {
		if bcc != "" {
			m.SetHeader("Bcc", m.FormatAddress(bcc, "收件人")) // 暗送
		}
	}
	m.SetHeader("Subject", subject) // 主题
	m.SetBody("text/html", body)    // 正文
	if fname != "" {
		h := map[string][]string{"Content-Type": {"text/plain; charset=UTF-8"}}
		m.Attach(fname, gomail.Rename(rename), gomail.SetHeader(h)) //添加附件
		//m.Attach(fname) //添加附件
	}
	reTry := auth.ReTry
	if reTry == 0 {
		reTry = 3
	}
	return gSend(reTry, auth, m, to)
}

//
func gSend(retry int, auth *GmailAuth, m *gomail.Message, to string) bool {
	defer Catch()
	dialer := getDialer(false, auth, to)
	defer func() {
		auth.PoolChan <- dialer
	}()
	status := false
	for i := 0; i < retry; i++ {
		if err := dialer.DialAndSend(m); err != nil {
			dialer = getDialer(true, auth, to)
			if retry > 0 {
				log.Println(auth.User, to, "第", i+1, "次发送邮件gSend error：", err)
				time.Sleep(200 * time.Millisecond)
			}
		} else {
			status = true
			break
		}
	}
	return status
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
