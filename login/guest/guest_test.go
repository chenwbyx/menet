package guest_test

import (
	"menet/login"
	_ "menet/login/guest"
	"testing"
)

func TestLoginGuest(t *testing.T) {
	cfg := map[string]map[string]string{
		"guest": {},
	}
	login.NewLogin(cfg)
	resp := login.CheckResp{}
	login.Validate(&login.CheckReq{
		"guest", "", "", "", "", ""}, &resp)
	if resp.Error != login.FAILED {
		t.Error("assert result is FAILED")
	}
	login.Validate(&login.CheckReq{
		"guest", "ergou", "", "", "", ""}, &resp)
	if resp.Error != login.SUCCESS {
		t.Error("assert result is SUCCESS")
	}
}
