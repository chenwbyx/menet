package login_test

import (
	"menet/login"
	_ "menet/login/weixin"
	"testing"
)

func TestLogin(t *testing.T) {
	defer func() {
		if p := recover(); p != nil {
			if p != "Login: Register adapter called twice. driver name weixin" {
				t.Error(p)
			}
		}
	}()
	cfg := map[string]map[string]string{
		"weixin": {
			"appid":  "123",
			"secret": "456",
		},
	}
	login.NewLogin(cfg)
	resp := login.CheckResp{}
	login.Validate(&login.CheckReq{
		Domain: "wei", Uin: "101", Ticket: "7788123"}, &resp)
	if resp.Error != login.INVALID_DOMAIN {
		t.Error("what?")
	}
	login.NewLogin(cfg)
}
