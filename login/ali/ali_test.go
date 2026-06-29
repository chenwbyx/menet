//go:build integration

package ali_test

import (
	"menet/login"
	_ "menet/login/ali"
	"testing"
)

func TestLoginAli(t *testing.T) {
	//defer func() {
	//	if p := recover(); p != nil {
	//		if p != "Login: Register adapter called twice. driver name weixin" {
	//			t.Error(p)
	//		}
	//	}
	//}()
	cfg := map[string]map[string]string{
		"ali": {
			"app_id":  "1046937",
			"app_key": "d97df50ac900f6b94b2822f32098cbb8",
		},
	}
	login.NewLogin(cfg)
	resp := login.CheckResp{}
	login.Validate(&login.CheckReq{
		Domain: "aliali", Ticket: "7788123"}, &resp)
	if resp.Error != login.INVALID_DOMAIN {
		t.Error("assert result is INVALID_DOMAIN")
	}
	login.Validate(&login.CheckReq{
		Domain: "ali", Ticket: "7788123"}, &resp)
	if resp.Error != login.FAILED {
		t.Error("assert result is FAILED")
	}
}
