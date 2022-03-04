package irc

import (
	"crypto/tls"
	"log"
	"strings"

	"gopkg.in/irc.v3"
)

/*

 */
type Bot struct {
	Addr    string
	Channel string
	Nick    string
	Client  *irc.Client
	trigger []func(c *irc.Client, m *irc.Message)
}

func NewBot(addr, channel, nick string) Bot {
	return Bot{
		Addr:    addr,
		Channel: channel,
		Nick:    nick,
	}
}

func (b *Bot) HandleOnCmd(cmd string, fn func(c *irc.Client, m *irc.Message)) {
	b.trigger = append(b.trigger, func(c *irc.Client, m *irc.Message) {
		if m.Command == cmd {
			fn(c, m)
		}
	})
}

func (b *Bot) HandleOnMsg(msg string, fn func(c *irc.Client, m *irc.Message)) {
	b.trigger = append(b.trigger, func(c *irc.Client, m *irc.Message) {
		if m.Command == "PRIVMSG" && strings.HasPrefix(m.Trailing(), msg) {
			fn(c, m)
		}
	})
}

func (b *Bot) HandleOnQuery(msg string, fn func(c *irc.Client, m *irc.Message)) {
	b.trigger = append(b.trigger, func(c *irc.Client, m *irc.Message) {
		if m.Command == "PRIVMSG" && m.Params[0] == b.Nick && OnCommand(m, msg) {
			fn(c, m)
		}
	})
}

func (b *Bot) AddHandler(f func(c *irc.Client, m *irc.Message)) {
	b.trigger = append(b.trigger, f)
}

//
func (b *Bot) MainHandler() irc.HandlerFunc {
	return func(c *irc.Client, m *irc.Message) {
		if m.Command == "001" {
			c.Write("JOIN " + b.Channel)
		}
		for _, v := range b.trigger {
			v(c, m)
		}
	}
}

func (b *Bot) Start() {
	conf := irc.ClientConfig{
		Nick:          b.Nick,
		Pass:          "",
		User:          b.Nick,
		Name:          b.Nick,
		PingFrequency: 0,
		PingTimeout:   0,
		SendLimit:     0,
		SendBurst:     0,
		Handler:       b.MainHandler(),
	}
	conn, err := tls.Dial("tcp", b.Addr, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		log.Fatal(err)
	}
	b.Client = irc.NewClient(conn, conf)
	b.Start()
}
