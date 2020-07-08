package kkk

import (
	"menet/login"
	"crypto/md5"
	"log"
	"fmt"
	"strings"
)

type Login3K struct {
	AppId  string
	Secret string
}

func (loginData *Login3K) Register(data map[string]string) {
	loginData.AppId = data["app_id"]
	loginData.Secret = data["app_key"]
}

// 3k顺序为 [时间戳 ，签名]
func (loginData *Login3K) Validate(jsonReq *login.CheckReq, jsonResp *login.CheckResp) {
	if len(jsonReq.Extra) != 2 {
		jsonResp.Error = login.FAILED
		return
	}
	timestamp, sign := jsonReq.Extra[0], jsonReq.Extra[1]
	sig := fmt.Sprintf("%x", md5.Sum([]byte(timestamp+jsonReq.Uin+loginData.Secret)))

	if strings.Compare(sig, sign) != 0 {
		log.Println(sig, "expect", sign)
		jsonResp.Error = login.INVALID_TOKEN
		return
	}

	jsonResp.Error = login.SUCCESS
}

func NewLogin3K() login.ILogin {
	return &Login3K{}
}

func init() {
	login.RegisterDriver("kkk", NewLogin3K)
}
