package irc

import (
	"fmt"
	"os/exec"
	"strings"

	"gopkg.in/irc.v3"
)

// send PRIVMSG
func PRIVMSG(c *irc.Client, to string, txt ...interface{}) {
	c.WriteMessage(&irc.Message{
		Command: "PRIVMSG",
		Params: []string{
			to,
			fmt.Sprintln(txt...),
		},
	})
}

// return PRIVMSG input as slice
func ParsePrivmsg(m *irc.Message) []string {
	msg := m.Params[1]
	return strings.Split(msg, " ")
}

// return PRIVMSG input as subcommand string and params slice
func ParsePrivmsg2(m *irc.Message) (string, []string) {
	msg := m.Params[1]
	cmd := strings.Split(msg, " ")
	var sc string
	var sv []string
	for k, v := range cmd {
		if k == 1 {
			sc = v
			continue
		}
		sv = append(sv, v)
	}
	return sc, sv
}

func OnCommand(m *irc.Message, cmd string) bool {
	mcmd, _ := ParsePrivmsg2(m)
	return mcmd == cmd
}

func ParseMultiLine(res string) []string {
	return strings.Split(res, "\n")
}

// DANGER!!!
// this action adds your bot ability to execute shell commands
// and may cause bot stop respond.
func AddShellExec(trigger string) func(c *irc.Client, m *irc.Message) {
	return func(c *irc.Client, m *irc.Message) {
		msg := m.Params[1]
		if strings.HasPrefix(msg, trigger) {
			cmd := strings.Split(msg, " ")[1:]
			var cc string
			var cp []string
			for k, v := range cmd {
				if k == 0 {
					cc = v
					continue
				}
				cp = append(cp, v)
			}
			ret, err := exec.Command(cc, cp...).Output()
			if err != nil {
				PRIVMSG(c, m.Name, fmt.Sprint("EXEC error ", err))
			}
			br := strings.Split(string(ret), "\n")
			PRIVMSG(c, m.Name, "EXEC "+strings.Join(cmd, " "))
			for _, v := range br {
				PRIVMSG(c, m.Name, v)
			}
			PRIVMSG(c, m.Name, "END EXEC")
			return
		}
	}
}
