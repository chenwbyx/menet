package core

import (
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"sync"
)

type PersistError string

func (e PersistError) Error() string { return string(e) }

const ELoadPollingTimeOut = 5

const EPersistStateDisk = 0
const EPersistStateLoading = 1
const EPersistStateMemory = 2
const EPersistStatePrepareUnloading = 3
const EPersistStateUnloading = 4

const EPersistErrorEngineNil = PersistError("persist: engine is nil")           // 启动关闭错误: 数据库连接失败
const EPersistErrorTempFileExist = PersistError("persist: temp file exist")     // 启动关闭错误: 存在临时bomb文件
const EPersistErrorInvalidBombFile = PersistError("persist: invalid bomb file") // 启动关闭错误: 无效的bomb文件
const EPersistErrorUnknownError = PersistError("persist: unknown error")        // 导入导出错误: 未知错误, 可能是并发引起
const EPersistErrorIncorrectState = PersistError("persist: incorrect state")    // 导入导出错误: 重复全导入或正在全导出
const EPersistErrorUnloading = PersistError("persist: unloading state")         // 导入导出错误: 正在导出, 导出完成后方可导入
const EPersistErrorAlreadyLoadAll = PersistError("persist: already load all")   // 导入导出错误: 已经全导入不能再按照key操作
const EPersistErrorLoading = PersistError("persist: loading state")             // 导入导出错误: 正在导入, 导入完成后方可导出
const EPersistErrorAlreadyLoad = PersistError("persist: already load")          // 导入导出错误: 重复导入
const EPersistErrorAlreadyUnload = PersistError("persist: already unload")      // 导入导出错误: 重复导出
const EPersistErrorNil = PersistError("persist: nil")                           // 增删改查错误: 非法的内存地址或空指针
const EPersistErrorAlreadyExist = PersistError("persist: already exist")        // 增删改查错误: 对象已经存在
const EPersistErrorNotInMemory = PersistError("persist: not in memory")         // 增删改查错误: 数据不在内存中
const EPersistErrorOutOfDate = PersistError("persist: out of date")             // 增删改查错误: 数据过期, 应当重新查询

type IPersist interface {
	Sync() (err error)
	Exit(wg *sync.WaitGroup)
	Run() (err error)
	Dead() bool
	RecoverBomb(bomb []byte) (err error)
	SyncData(wg *sync.WaitGroup) (err error)
}

type IPersistUser interface {
	IPersist
	Load(Uid int32) (err error)
	Unload(Uid int32) (err error)
	SetLoadState2Memory(Uid int32)
	LoadState(Uid int32) int32
}

var gPersistMap = make(map[string]IPersist)
var gPersistUserMap = make(map[string]IPersistUser)

func RegisterPersist(name string, persist IPersist) {
	if _, ok := gPersistMap[name]; ok {
		panic(errors.New("repeated register persist " + name))
	}
	gPersistMap[name] = persist

	persistUser, ok := persist.(IPersistUser)
	if ok {
		if _, ok := gPersistUserMap[name]; ok {
			fmt.Printf("%#v\n", gPersistUserMap)
			panic(errors.New("repeated register persist user " + name))
		}
		gPersistUserMap[name] = persistUser
	} else {
	}
}

func GetIPersistByName(name string) IPersist {
	persist, ok := gPersistMap[name]
	if ok {
		return persist
	} else {
		return nil
	}
}

func Load(uid int32) (err error) {
	for _, persist := range gPersistUserMap {
		err = persist.Load(uid)
		if err != nil {
			return err
		}
	}
	return
}

func SetLoadState2Memory(uid int32) {
	for _, persist := range gPersistUserMap {
		persist.SetLoadState2Memory(uid)
	}
	return
}

func Unload(uid int32) (err error) {
	for _, persist := range gPersistUserMap {
		err = persist.Unload(uid)
		if err != nil {
			return err
		}
	}
	return
}

func LoadState(uid int32) (stateList []int32) {
	for _, persist := range gPersistUserMap {
		stateList = append(stateList, persist.LoadState(uid))
	}
	return
}

func SyncPersist() error {
	for _, persist := range gPersistMap {
		err := persist.Sync()
		if err != nil {
			return err
		}
	}
	return nil
}

func RunPersist() error {
	for _, persist := range gPersistMap {
		err := persist.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

func DeadPersist() bool {
	for _, persist := range gPersistMap {
		dead := persist.Dead()
		if dead {
			return true
		}
	}
	return false
}

func ExitPersist() {
	var wg sync.WaitGroup
	for key := range gPersistMap {
		wg.Add(1)
		go gPersistMap[key].Exit(&wg)
	}
	wg.Wait()
}

func SyncDataPersist() error {
	var wg sync.WaitGroup
	ch := make(chan error, len(gPersistMap))
	for key := range gPersistMap {
		wg.Add(1)
		go func(name string) {
			err := gPersistMap[name].SyncData(&wg)
			if err != nil {
				ch <- err
			}
		}(key)
	}
	wg.Wait()
	select {
	case err, ok := <-ch:
		if ok {
			return err
		} else {
			return nil
		}
	default:
		return nil
	}
}
