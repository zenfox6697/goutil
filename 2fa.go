package goutil

import (
	"time"

	"github.com/xlzd/gotp"
)

type TOTP struct {
	Secret   string
	Digits   int
	Interval int
}

func NewTOTP(digit, interval int) TOTP {
	return TOTP{
		Secret:   gotp.RandomSecret(64),
		Digits:   digit,
		Interval: interval,
	}
}

func (t TOTP) Verify(a string) bool {
	if t.Digits == 0 || t.Interval == 0 {
		t.Digits = 6
		t.Interval = 60
	}
	return gotp.NewTOTP(t.Secret, t.Digits, t.Interval, nil).Verify(a, int(time.Now().Unix()))
}

func (t TOTP) At(stamp int) string {
	if t.Digits == 0 || t.Interval == 0 {
		t.Digits = 6
		t.Interval = 60
	}
	return gotp.NewTOTP(t.Secret, t.Digits, t.Interval, nil).At(stamp)
}

func (t TOTP) Now() string {
	if t.Digits == 0 || t.Interval == 0 {
		t.Digits = 6
		t.Interval = 60
	}
	return gotp.NewTOTP(t.Secret, t.Digits, t.Interval, nil).Now()
}

func (t TOTP) GetLink(username, appname string) string {
	if t.Digits == 0 || t.Interval == 0 {
		t.Digits = 6
		t.Interval = 60
	}
	return gotp.NewTOTP(t.Secret, t.Digits, t.Interval, nil).ProvisioningUri(username, appname)
}

type HOTP struct {
	Secret string
	Digits int
}

func NewHOTP(digit int) HOTP {
	return HOTP{
		Secret: gotp.RandomSecret(64),
		Digits: digit,
	}
}

func (h HOTP) Verify(a string, c int) bool {
	if h.Digits == 0 {
		h.Digits = 6
	}
	return gotp.NewHOTP(h.Secret, h.Digits, nil).Verify(a, c)
}

func (h HOTP) At(c int) string {
	if h.Digits == 0 {
		h.Digits = 6
	}
	return gotp.NewHOTP(h.Secret, h.Digits, nil).At(c)
}

func (h HOTP) GetLink(username, appname string, initcount int) string {
	if h.Digits == 0 {
		h.Digits = 6
	}
	return gotp.NewHOTP(h.Secret, h.Digits, nil).ProvisioningUri(username, appname, initcount)
}
