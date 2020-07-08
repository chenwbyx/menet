package guest

import (
	"github.com/axgle/mahonia"
	"menet/login"
)

type LoginGuest struct {
}

func (loginData *LoginGuest) Register(data map[string]string) {
}

func (loginData *LoginGuest) Validate(jsonReq *login.CheckReq, jsonResp *login.CheckResp) {
	decoderUTF8 := mahonia.NewDecoder("utf8")
	if jsonReq.Uin == "" {
		jsonResp.Error = login.FAILED
		return
	}
	uin := decoderUTF8.ConvertString(jsonReq.Uin)
	jsonResp.User_profile, jsonResp.User_name = "http://img4.imgtn.bdimg.com/it/u=2802024393,2096819393&fm=200&gp=0.jpg", "游客"
	jsonReq.Uin = uin
	jsonResp.Error = login.SUCCESS
}

func NewLoginGuest() login.ILogin {
	return &LoginGuest{}
}

func init() {
	login.RegisterDriver("guest", NewLoginGuest)
}
