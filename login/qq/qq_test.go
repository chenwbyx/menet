package qq_test

import (
	"menet/login"
	_ "menet/login/qq"
	"testing"
)

func TestLoginQQ(t *testing.T) {
	//defer func() {
	//	if p := recover(); p != nil {
	//		if p != "Login: Register adapter called twice. driver name weixin" {
	//			t.Error(p)
	//		}
	//	}
	//}()
	cfg := map[string]map[string]string{
		"qq": {
			"appid":  "1106393515APP",
			"secret": "tLkInLYOeZQmc21Z",
		},
	}
	login.NewLogin(cfg)
	resp := login.CheckResp{}
	login.Validate(&login.CheckReq{
		"qq", "123", "asfsdf", "", "1.0.0", ""}, &resp)
	if resp.Error != login.INVALID_TOKEN {
		t.Error("assert result is FAILED")
	}
}
