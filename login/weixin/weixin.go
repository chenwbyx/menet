package weixin

import (
	"bytes"
	"encoding/json"
	"github.com/axgle/mahonia"
	"io/ioutil"
	"menet/login"
	"net/http"
)

type LoginWeiXin struct {
	AppId                 string
	Secret                string
	AccessTokenUrlFormat  string
	RefreshTokenUrlFormat string
	CheckTokenUrlFormat   string
	GetIconBaseUrlFormat  string
}

const AccessTokenBaseUrl = `https://api.weixin.qq.com/sns/oauth2/access_token?`
const RefreshTokenBaseUrl = `https://api.weixin.qq.com/sns/oauth2/refresh_token?`
const CheckTokenBaseUrl = `https://api.weixin.qq.com/sns/auth?`
const GetIconBaseUrl = `https://api.weixin.qq.com/sns/userinfo?`

func joinAccessTokenUrlFormat(appid, secret string) string {
	buffer := bytes.Buffer{}
	buffer.WriteString(AccessTokenBaseUrl)
	buffer.WriteString("appid=")
	buffer.WriteString(appid)
	buffer.WriteString("&secret=")
	buffer.WriteString(secret)
	buffer.WriteString("&grant_type=authorization_code")
	buffer.WriteString("&code=")

	return buffer.String()
}

func joinRefreshTokenUrlFormat(appid, secret string) string {
	buffer := bytes.Buffer{}
	buffer.WriteString(RefreshTokenBaseUrl)
	buffer.WriteString("appid=")
	buffer.WriteString(appid)
	buffer.WriteString("&grant_type=refresh_token")
	buffer.WriteString("&refresh_token=")

	return buffer.String()
}

func joinCheckTokenUrl(accessToken, uin string) string {
	buffer := bytes.Buffer{}
	buffer.WriteString(CheckTokenBaseUrl)
	buffer.WriteString("access_token=")
	buffer.WriteString(accessToken)
	buffer.WriteString("&openid=")
	buffer.WriteString(uin)

	return buffer.String()
}

func joinGetIconUrl(accessToken, uin string) string {
	buffer := bytes.Buffer{}
	buffer.WriteString(GetIconBaseUrl)
	buffer.WriteString("access_token=")
	buffer.WriteString(accessToken)
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

func getUserProfile(token, uin string) (icon, name string) {
	url := joinGetIconUrl(token, uin)
	result := getResult(url)
	icon, name = "", ""
	if result != nil {
		_, existErrorCode := (*result)["errcode"]
		if !existErrorCode {
			if headimgurl, ok := (*result)["headimgurl"]; ok {
				icon = headimgurl.(string)
			}
			if nickname, ok := (*result)["nickname"]; ok {
				name = nickname.(string)
			}
		}

	}
	return
}

func (loginData *LoginWeiXin) Register(data map[string]string) {
	loginData.AppId = data["appid"]
	loginData.Secret = data["secret"]
	loginData.AccessTokenUrlFormat = joinAccessTokenUrlFormat(loginData.AppId, loginData.Secret)
	loginData.RefreshTokenUrlFormat = joinRefreshTokenUrlFormat(loginData.AppId, loginData.Secret)
}

func (loginData *LoginWeiXin) Validate(jsonReq *login.CheckReq, jsonResp *login.CheckResp) {
	var accessToken, refreshToken string
	decoderUTF8 := mahonia.NewDecoder("utf8")
	encoderUTF8 := mahonia.NewEncoder("utf8")
	// 第一次登录, 获取token
	if jsonReq.Ticket != "" {
		token := decoderUTF8.ConvertString(jsonReq.Ticket)
		result := getResult(loginData.AccessTokenUrlFormat + token)
		if result != nil {
			errCode, existErrorCode := (*result)["errcode"]
			if !existErrorCode {
				if access_token, ok := (*result)["access_token"]; ok {
					accessToken = access_token.(string)
				}
				if refresh_token, ok := (*result)["refresh_token"]; ok {
					refreshToken = refresh_token.(string)
				}
				jsonResp.Token = encoderUTF8.ConvertString(accessToken)
				if uin, ok := (*result)["openid"]; ok {
					jsonReq.Uin = uin.(string)
				}
			} else {
				if int(errCode.(float64)) == 40029 {
					jsonResp.Error = login.INVALID_TOKEN
				} else {
					jsonResp.Error = login.FAILED
				}
				return
			}
		} else {
			jsonResp.Error = login.FAILED
			return
		}
	}
	// 第一次登录, 自动刷新token
	if refreshToken != "" {
		result := getResult(loginData.RefreshTokenUrlFormat + refreshToken)
		if result != nil {
			errCode, existErrorCode := (*result)["errcode"]
			if !existErrorCode {
				if access_token, ok := (*result)["access_token"]; ok {
					accessToken = access_token.(string)
					jsonResp.Token = encoderUTF8.ConvertString(accessToken)
				}
				if uin, ok := (*result)["openid"]; ok {
					jsonReq.Uin = uin.(string)
				}

			} else {
				if int(errCode.(float64)) == 40030 {
					jsonResp.Error = login.INVALID_TOKEN
				} else {
					jsonResp.Error = login.FAILED
				}
			}
		} else {
			jsonResp.Error = login.FAILED
			return
		}
	}
	// 第二次登录, 用刷新过后的token验证
	if len(jsonReq.Extra) > 0 {
		accessToken = decoderUTF8.ConvertString(jsonReq.Extra[0])
	}

	// 没有token 不能继续了
	if accessToken == "" || jsonReq.Uin == "" {
		jsonResp.Error = login.FAILED
		return
	}

	result := getResult(joinCheckTokenUrl(accessToken, jsonReq.Uin))
	if result != nil {
		errCode, existErrorCode := (*result)["errcode"]
		//if !existErrorCode || (existErrorCode && errCode.(int) == 0) {
		if !existErrorCode || (existErrorCode && int(errCode.(float64)) == 0) {
			jsonResp.Error = login.SUCCESS
			jsonResp.User_profile, jsonResp.User_name = getUserProfile(accessToken, jsonReq.Uin)
			return
		} else {
			jsonResp.Error = login.INVALID_TOKEN
			return
		}
	}

	jsonResp.Error = login.FAILED
}

func NewLoginWeixin() login.ILogin {
	return &LoginWeiXin{}
}

func init() {
	login.RegisterDriver("weixin", NewLoginWeixin)
}
