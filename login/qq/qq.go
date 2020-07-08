package qq

import (
	"bytes"
	"encoding/json"
	"github.com/axgle/mahonia"
	"io/ioutil"
	"menet/login"
	"net/http"
)

type LoginQQ struct {
	AppId  string
	Secret string
}

const GetUserInfoBaseUrl = `https://graph.qq.com/user/get_user_info?`

func joinGetUserInfoUrl(accessToken, uin, appid string) string {
	buffer := bytes.Buffer{}
	buffer.WriteString(GetUserInfoBaseUrl)
	buffer.WriteString("access_token=")
	buffer.WriteString(accessToken)
	buffer.WriteString("&oauth_consumer_key=")
	buffer.WriteString(appid)
	buffer.WriteString("&openid=")
	buffer.WriteString(uin)

	return buffer.String()
}

func getResult(url string) *map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		return nil
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	retMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(body), &retMap)
	if err != nil {
		return nil
	}
	return &retMap
}

func (loginData *LoginQQ) Register(data map[string]string) {
	loginData.AppId = data["appid"]
	loginData.Secret = data["secret"]
}

func (loginData *LoginQQ) Validate(jsonReq *login.CheckReq, jsonResp *login.CheckResp) {
	decoderUTF8 := mahonia.NewDecoder("utf8")
	//encoderUTF8 := mahonia.NewEncoder("utf8")
	token := decoderUTF8.ConvertString(jsonReq.Ticket)
	uin := decoderUTF8.ConvertString(jsonReq.Uin)

	url := joinGetUserInfoUrl(token, uin, loginData.AppId)
	result := getResult(url)
	var icon, name, icon_1, icon_2 string
	if result != nil {
		errCode, existErrorCode := (*result)["ret"]
		if !existErrorCode || (int(errCode.(float64)) == 0) {
			if headimgurl_1, ok := (*result)["figureurl_qq_1"]; ok {
				icon_1 = headimgurl_1.(string)
			}
			if headimgurl_2, ok := (*result)["figureurl_qq_2"]; ok {
				icon_2 = headimgurl_2.(string)
			}
			if icon_2 == "" {
				icon = icon_1
			} else {
				icon = icon_2
			}
			if nickname, ok := (*result)["nickname"]; ok {
				name = nickname.(string)
			}
			jsonResp.User_profile, jsonResp.User_name = icon, name
			jsonResp.Error = login.SUCCESS
			return
		} else {
			jsonResp.Error = login.INVALID_TOKEN
			return
		}
	}
	jsonResp.Error = login.FAILED
}

func NewLoginQQ() login.ILogin {
	return &LoginQQ{}
}

func init() {
	login.RegisterDriver("qq", NewLoginQQ)
}
