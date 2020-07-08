//go:generate go install menet/persist
//go:generate persist -src=menet/persist/tests/protocol -fileName=col_type.go -unload=true Persist PersistSyncMap
package model

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"log"
	"menet/persist/core"
	"sync"
	"time"
)

var gEngine *xorm.Engine
var engineOnce sync.Once

var DBAddr = "root:123456@tcp(172.168.11.10:3306)/persist_test?charset=utf8mb4"

func GetDB() *xorm.Engine {
	engineOnce.Do(func() {
		var err error
		gEngine, err = xorm.NewEngine("mysql", DBAddr)
		if err != nil {
			log.Println("GetDB error", err)
		} else {
			//gEngine.ShowSQL(true)
			gEngine.SetMaxIdleConns(2)            //设置连接池中的保持连接的最大连接数
			gEngine.SetMaxOpenConns(10)           //设置连接池的打开的最大连接数
			gEngine.SetConnMaxLifetime(time.Hour) //设置连接超时时间
		}
	})
	return gEngine
}

func Register(name string, persist core.IPersist) {
	core.RegisterPersist(name, persist)
}

func Exit() {
	log.Println("Save Begin")
	core.ExitPersist()
	log.Println("Save End")
}

func Run() error {
	log.Println("Run Begin")
	err := core.RunPersist()
	if err != nil {
		return err
	}
	log.Println("Run End")
	return nil
}

func Dead() bool {
	return core.DeadPersist()
}

func Init() {
	engine := GetDB()
	if engine == nil {
		panic(core.EPersistErrorEngineNil)
	}
	// create scheme here
	//engine.Sync(&Account{})
	core.SyncPersist()

	// create scheme here
}
