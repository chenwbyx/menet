package ali

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"log"
	"menet/login"
	"net/http"
	"strconv"
	"time"
)

type LoginAli struct {
	AppId  int64
	Secret string
}

const VerifySessionBaseUrl = `http://sdk.9game.cn/cp/account.verifySession?`

func getResult(url string, data []byte) *VerifySessionResp {
	body := bytes.NewBuffer(data)
	res, err := http.Post(url, "application/json;charset=utf-8", body)
	if err != nil {
		return nil
	}
	result, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil
	}
	err = res.Body.Close()
	if err != nil {
		return nil
	}

	resp := &VerifySessionResp{}
	err = json.Unmarshal([]byte(result), resp)
	if err != nil {
		return nil
	}
	return resp
}

type VerifySessionReq struct {
	Id   int64             `json:"id"`
	Data map[string]string `json:"data"`
	Game map[string]int64  `json:"game"`
	Sign string            `json:"sign"`
}

type VerifySessionRespState struct {
	Code int64  `json:"code"`
	Msg  string `json:"msg"`
}

type VerifySessionResp struct {
	Id    int64                  `json:"id"`
	State VerifySessionRespState `json:"state"`
	Data  map[string]string      `json:"data"`
}

func (loginData *LoginAli) getVerifySessionReq(sid string) ([]byte, error) {
	req := VerifySessionReq{}
	req.Id = time.Now().Unix()
	req.Data = map[string]string{"sid": sid}
	req.Game = map[string]int64{"gameId": loginData.AppId}
	md5Ctx1 := md5.New()
	md5Ctx1.Write([]byte("sid=" + sid + loginData.Secret))
	cipherStr1 := md5Ctx1.Sum(nil)
	req.Sign = hex.EncodeToString(cipherStr1)

	ret, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (loginData *LoginAli) Register(data map[string]string) {
	appId, err := strconv.ParseInt(data["app_id"], 10, 0)
	if err != nil {
		panic("LoginAli Register fail" + err.Error())
	}
	loginData.AppId = appId
	loginData.Secret = data["app_key"]
}

func (loginData *LoginAli) Validate(jsonReq *login.CheckReq, jsonResp *login.CheckResp) {
	if jsonReq.Ticket == "" {
		jsonResp.Error = login.INVALID_TOKEN
		return
	}

	data, err := loginData.getVerifySessionReq(jsonReq.Ticket)
	if err != nil {
		jsonResp.Error = login.FAILED
		return
	}

	result := getResult(VerifySessionBaseUrl, data)
	if result != nil {
		if result.State.Code != 1 {
			log.Println("LoginAli fail. error msg", result.State.Code, result.State.Msg)
			jsonResp.Error = login.INVALID_TOKEN
			return
		}
		if accountId, exist := result.Data["accountId"]; exist {
			jsonReq.Uin = accountId
			jsonResp.App_data = accountId
			jsonResp.Error = login.SUCCESS
			if nickName, exist := result.Data["nickName"]; exist {
				jsonResp.User_name = nickName
			}
			return
		} else {
			jsonResp.Error = login.INVALID_TOKEN
			return
		}
	}

	jsonResp.Error = login.FAILED
}

func NewLoginAli() login.ILogin {
	return &LoginAli{}
}

func init() {
	login.RegisterDriver("ali", NewLoginAli)
}
