package zhiqu

import (
	"bytes"
	"encoding/json"
	"github.com/axgle/mahonia"
	"io/ioutil"
	"log"
	"menet/login"
	"net/http"
)

type LoginZhiqu struct {
	AppId   string
	Secret  string
	BaseUrl string
}

const AccessTokenBaseUrl = `http://zqintpass.metekonline.com/gametoken/default.aspx?`

func JoinAppId(base, appid string) string {
	buffer := bytes.Buffer{}
	buffer.WriteString(base)
	buffer.WriteString("appid=")
	buffer.WriteString(appid)
	return buffer.String()
}

func JoinUidToken(base, uid, token string) string {
	buffer := bytes.Buffer{}
	buffer.WriteString(base)
	buffer.WriteString("&uid=")
	buffer.WriteString(uid)
	buffer.WriteString("&token=")
	buffer.WriteString(token)
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
	log.Println(retMap, body)
	if err != nil {
		return nil
	}
	return &retMap
}

func (loginData *LoginZhiqu) Register(data map[string]string) {
	loginData.AppId = data["app_id"]
	loginData.Secret = data["app_key"]
	loginData.BaseUrl = JoinAppId(AccessTokenBaseUrl, loginData.AppId)
}

func (loginData *LoginZhiqu) Validate(jsonReq *login.CheckReq, jsonResp *login.CheckResp) {
	decoderUTF8 := mahonia.NewDecoder("utf8")
	//encoderUTF8 := mahonia.NewEncoder("utf8")
	if jsonReq.Ticket != "" && jsonReq.Uin != "" {
		token := decoderUTF8.ConvertString(jsonReq.Ticket)
		uid := decoderUTF8.ConvertString(jsonReq.Uin)
		result := getResult(JoinUidToken(loginData.BaseUrl, uid, token))
		if result != nil {
			errCode, exist := (*result)["isok"]
			if exist {
				if int(errCode.(float64)) == 0 {
					jsonResp.App_data = uid
					jsonResp.Error = login.SUCCESS
				} else {
					jsonResp.Error = login.INVALID_TOKEN
				}
			} else {
				jsonResp.Error = login.FAILED
			}
		} else {
			jsonResp.Error = login.FAILED
		}
	} else {
		jsonResp.Error = login.FAILED
	}
}

func NewLoginZhiqu() login.ILogin {
	return &LoginZhiqu{}
}

func init() {
	login.RegisterDriver("zhiqu", NewLoginZhiqu)
}
