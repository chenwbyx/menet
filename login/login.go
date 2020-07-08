// Package login provide a login interface and some implement method
// Usage:
//
// import(
//   "login"
// )
//
// login.NewLogin(serverConfig.Login)
//
// Use it like this:
//
//	login.Validate(...)
//
package login

type CheckReq struct {
	Domain    string `json:"domain"`
	Uin       string `json:"uin"`
	Ticket    string `json:"ticket"`
	Extra     []string `json:"extra"`
	App_data  string `json:"app_data"`
	Server_id string `json:"server_id"`
}

const (
	SUCCESS           int = 0x00
	FAILED            int = 0x01
	INVALID_TOKEN     int = 0x02
	INVALID_DOMAIN    int = 0x03
	INVALID_SERVER_ID int = 0x04
)

type CheckResp struct {
	Error        int    `json:"error"`
	Token        string `json:"token"`
	App_data     string `json:"app_data"`
	User_id      int64  `json:"user_id"`
	User_name    string `json:"user_name"`
	User_profile string `json:"user_profile"`
	First        bool   `json:"first"`
}

type ILogin interface {
	Register(data map[string]string)
	Validate(jsonReq *CheckReq, jsonResp *CheckResp)
}

// Instance is a function create a new ILogin Instance
type Instance func() ILogin

var adapters = make(map[string]Instance)

var validators = make(map[string]ILogin)

// RegisterDriver makes a ILogin adapter available by the adapter name.
// If RegisterDriver is called twice with the same name or if driver is nil,
// it panics.
func RegisterDriver(name string, adapter Instance) {
	if adapter == nil {
		panic("Login: RegisterDriver adapter is nil")
	}
	if _, ok := adapters[name]; ok {
		panic("Login: RegisterDriver called twice for adapter " + name)
	}
	adapters[name] = adapter
}

func NewLogin(cfgMap map[string]map[string]string) (err error) {
	err = nil
	for key, value := range cfgMap {
		newFunc, ok := adapters[key]
		if !ok {
			panic("Login: Register adapter is nil. driver name " + key)
		}
		instance, ok := validators[key]
		if ok {
			panic("Login: Register adapter called twice. driver name " + key)
		}
		instance = newFunc()
		instance.Register(value)
		validators[key] = instance
	}
	return
}

func Validate(jsonReq *CheckReq, jsonResp *CheckResp) {
	instance, ok := validators[jsonReq.Domain]
	if !ok {
		jsonResp.Error = INVALID_DOMAIN
		return
	}
	instance.Validate(jsonReq, jsonResp)
	return
}
