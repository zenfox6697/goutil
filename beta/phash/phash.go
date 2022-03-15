package phash

import (
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"

	"github.com/corona10/goimagehash"
)

func GetPhash(url string) (hash uint64, err error) {
	retry := 0

	println(1)

	client := &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}

	// 计算 /132 结尾的phash可能有误 这里用计算 /0结尾的
	if strings.HasSuffix(url, "/132") {
		url = url[0:len(url)-3] + "0"
	}

	for {
		println(2)
		resp, err := client.Get(url)
		if err != nil {
			return 0, err
		}
		println(3)
		if resp.Body == nil {
			return 0, errors.New("body is nil")
		}

		img, _, err := image.Decode(resp.Body)
		// img, err := image.Decode(resp.Body)
		println(4)
		_ = resp.Body.Close()
		if err != nil {
			retry++
			if retry > 3 {
				return 0, err
			}
			continue
		}
		println(5)

		imgHash, err := goimagehash.PerceptionHash(img)
		if err != nil {
			return 0, err
		}
		println(6)

		return imgHash.GetHash(), nil
	}
}
