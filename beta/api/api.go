package api

// import (
// 	"bytes"
// 	"log"
// 	"net/http"
// )

// type Caller struct {
// 	BaseUrl string
// }

// func (c *Caller) GET(suburl string, header http.Header, body []byte) *http.Response {
// 	req, err := http.NewRequest("GET", c.BaseUrl+suburl, bytes.NewReader(body))
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	return resp
// }
