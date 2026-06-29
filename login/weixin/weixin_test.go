package weixin_test

import (
	"menet/login"
	_ "menet/login/weixin"
	"testing"
)

func TestLoginWeiXin(t *testing.T) {
	//defer func() {
	//	if p := recover(); p != nil {
	//		if p != "Login: Register adapter called twice. driver name weixin" {
	//			t.Error(p)
	//		}
	//	}
	//}()
	cfg := map[string]map[string]string{
		"weixin": {
			"appid":  "wx3a73b91cdb463cbb",
			"secret": "6a37d5dabd5eabb70f87bdf4a9c9abf7",
		},
	}
	login.NewLogin(cfg)
	resp := login.CheckResp{}
	login.Validate(&login.CheckReq{
		"weixin", "101", "7788123", nil, "", ""}, &resp)
	if resp.Error != login.INVALID_DOMAIN {
		t.Error("assert result is INVALID_DOMAIN")
	}
	login.Validate(&login.CheckReq{
		"weixin", "", "", nil, "7788123", ""}, &resp)
	if resp.Error != login.FAILED {
		t.Error("assert result is FAILED")
	}
}
