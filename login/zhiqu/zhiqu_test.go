package zhiqu_test

import (
	"menet/login"
	_ "menet/login/zhiqu"
	"testing"
)

func TestLoginZhiqu(t *testing.T) {
	cfg := map[string]map[string]string{
		"zhiqu": {
			"app_id":  "T036",
			"app_key": "0aa43a3f755de3fd5a0ecc7be4f86233",
		},
	}
	login.NewLogin(cfg)
	resp := login.CheckResp{}
	login.Validate(&login.CheckReq{
		"zhiqu123", "101", "7788123", nil, "", ""}, &resp)
	if resp.Error != login.INVALID_DOMAIN {
		t.Error("assert result is INVALID_DOMAIN")
	}

	login.Validate(&login.CheckReq{
		"zhiqu", "101", "7788123", nil, "", ""}, &resp)
	if resp.Error != login.INVALID_TOKEN {
		t.Error("assert result is INVALID_TOKEN")
	}

	login.Validate(&login.CheckReq{
		"zhiqu", "", "", nil, "7788123", ""}, &resp)
	if resp.Error != login.FAILED {
		t.Error("assert result is FAILED")
	}
}
