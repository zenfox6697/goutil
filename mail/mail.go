/*
 邮件发送操作包装，<br/>
 暂支持邮件发送功能。<br/>
 邮件内容若想模板化，请在应用程序中处理。<br/>
 暂不支持附件
*/
package mail

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"
)

const (
	SPLIT = "\r\n"
)

//邮箱认证信息
type MailAuth struct {
	SmtpHost string //邮箱服务器
	SmtpPort int    //邮箱端口
	User     string //用户
	Pwd      string //密码
}

//邮件消息
type Message struct {
	Subject string   //主题
	From    string   //来自
	To      []string //接收人,可以是多个接收人
	Body    string   //html邮件内容
}

//发送邮件
func SendMail(auth *MailAuth, mes *Message) {
	go sendMailJob(auth, mes)
}

//发送邮件
func sendMailJob(auth *MailAuth, mes *Message) error {
	buf := bytes.NewBuffer(nil)
	buf.WriteString("Subject: " + mes.Subject + SPLIT)
	buf.WriteString("MIME-Version: 1.0" + SPLIT)
	buf.WriteString("From: " + mes.From + SPLIT)
	buf.WriteString("To: " + strings.Join(mes.To, ";") + SPLIT)
	buf.WriteString("Content-Type: " + "text/html; charset=UTF-8" + SPLIT)
	buf.WriteString(SPLIT + mes.Body)
	smtpauth := smtp.PlainAuth(
		"",
		auth.User,
		auth.Pwd,
		auth.SmtpHost,
	)
	err := sendMailUsingTLS(
		fmt.Sprintf("%s:%d", auth.SmtpHost, auth.SmtpPort),
		smtpauth,
		auth.User,
		mes.To,
		buf.Bytes(),
	)
	if err != nil {
		fmt.Println(err.Error())
	}
	return err
}

func dial(addr string) (*smtp.Client, error) {
	config := &tls.Config{ServerName: addr, InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", addr, config)
	if err != nil {
		log.Println("Dialing Error:", err)
		return nil, err
	}
	//分解主机端口字符串
	host, _, _ := net.SplitHostPort(addr)
	return smtp.NewClient(conn, host)
}

//安全协议发送邮件
func sendMailUsingTLS(addr string, auth smtp.Auth, from string,
	to []string, msg []byte) (err error) {

	//create smtp client
	c, err := dial(addr)
	if err != nil {
		log.Println("Create smpt client error:", err)
		return err
	}
	defer c.Close()

	if auth != nil {
		if ok, _ := c.Extension("AUTH"); ok {
			if err = c.Auth(auth); err != nil {
				log.Println("Error during AUTH", err)
				return err
			}
		}
	}

	if err = c.Mail(from); err != nil {
		return err
	}

	for _, addr := range to {
		if err = c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return c.Quit()
}
