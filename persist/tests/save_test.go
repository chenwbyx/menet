package tests

import (
	"bytes"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"log"
	"menet/persist/core"
	"menet/persist/tests/model"
	"menet/persist/tests/protocol"
	"testing"
)

func TestSaveError(t *testing.T) {
	manage := SetupFunction()
	defer manage.PersistManager.RemoveFile()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)

	firstCls := model.NewPersist(1, 1)
	firstCls.Value = 100
	_ = model.GPersistManager.MarkUpdate(firstCls)

	secondCls := &protocol.Persist{Uid: 1, Id: 1}
	assert.EqualValues(t, core.EPersistErrorAlreadyExist, model.GPersistManager.NewPersist(secondCls))

	assert.EqualValues(t, 100, model.GPersistManager.GetPersistByUidId(1, 1).Value)

}

func TestRunWithBombFile(t *testing.T) {
	data := []byte(`Persist [{"Data":{"Uid":1,"Id":0,"State":0,"Value":0,"String":"","ValueMap":null,"ValueArray":[0,0,0,0],"ValueSliceList":null,"MapMap":null,"MapList":null,"MapSList":null,"SListMap":null,"SListList":null,"SListSList":null,"ListMap":[null,null,null,null],"ListList":[[0,0,0,0],[0,0,0,0],[0,0,0,0],[0,0,0,0]],"ListSList":[null,null,null,null]},"Op":1},{"Data":{"Uid":2,"Id":0,"State":0,"Value":0,"String":"","ValueMap":null,"ValueArray":[0,0,0,0],"ValueSliceList":null,"MapMap":null,"MapList":null,"MapSList":null,"SListMap":null,"SListList":null,"SListSList":null,"ListMap":[null,null,null,null],"ListList":[[0,0,0,0],[0,0,0,0],[0,0,0,0],[0,0,0,0]],"ListSList":[null,null,null,null]},"Op":1},{"Data":{"Uid":3,"Id":0,"State":0,"Value":0,"String":"","ValueMap":null,"ValueArray":[0,0,0,0],"ValueSliceList":null,"MapMap":null,"MapList":null,"MapSList":null,"SListMap":null,"SListList":null,"SListSList":null,"ListMap":[null,null,null,null],"ListList":[[0,0,0,0],[0,0,0,0],[0,0,0,0],[0,0,0,0]],"ListSList":[null,null,null,null]},"Op":1},{"Data":{"Uid":4,"Id":0,"State":0,"Value":0,"String":"","ValueMap":null,"ValueArray":[0,0,0,0],"ValueSliceList":null,"MapMap":null,"MapList":null,"MapSList":null,"SListMap":null,"SListList":null,"SListSList":null,"ListMap":[null,null,null,null],"ListList":[[0,0,0,0],[0,0,0,0],[0,0,0,0],[0,0,0,0]],"ListSList":[null,null,null,null]},"Op":1}]`)
	pos := bytes.IndexByte(data, byte(' '))
	assert.EqualValues(t, true, pos != -1)
	persistData := data[pos+1:]
	err := json.Unmarshal(persistData, &model.GPersistManager.FailQueue)
	assert.EqualValues(t, true, err == nil)
	model.GPersistManager.SaveFile()
	model.GPersistManager.FailQueue = model.GPersistManager.FailQueue[0:0]
	if model.Dead() {
		model.Run()
	}
	assert.EqualValues(t, 4, len(model.GPersistManager.FailQueue))
	model.GPersistManager.FailQueue = model.GPersistManager.FailQueue[0:0]
	model.Exit()
	engine := model.GetDB()
	_, _ = engine.QueryString("delete from persist;")
}

func TestReadBombFile(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	data := []byte(`Persist [{"Data":{"Uid":1,"Id":0,"State":0,"Value":0,"String":"","ValueMap":null,"ValueArray":[0,0,0,0],"ValueSliceList":null,"MapMap":null,"MapList":null,"MapSList":null,"SListMap":null,"SListList":null,"SListSList":null,"ListMap":[null,null,null,null],"ListList":[[0,0,0,0],[0,0,0,0],[0,0,0,0],[0,0,0,0]],"ListSList":[null,null,null,null]},"Op":1},{"Data":{"Uid":2,"Id":0,"State":0,"Value":0,"String":"","ValueMap":null,"ValueArray":[0,0,0,0],"ValueSliceList":null,"MapMap":null,"MapList":null,"MapSList":null,"SListMap":null,"SListList":null,"SListSList":null,"ListMap":[null,null,null,null],"ListList":[[0,0,0,0],[0,0,0,0],[0,0,0,0],[0,0,0,0]],"ListSList":[null,null,null,null]},"Op":1},{"Data":{"Uid":3,"Id":0,"State":0,"Value":0,"String":"","ValueMap":null,"ValueArray":[0,0,0,0],"ValueSliceList":null,"MapMap":null,"MapList":null,"MapSList":null,"SListMap":null,"SListList":null,"SListSList":null,"ListMap":[null,null,null,null],"ListList":[[0,0,0,0],[0,0,0,0],[0,0,0,0],[0,0,0,0]],"ListSList":[null,null,null,null]},"Op":1},{"Data":{"Uid":4,"Id":0,"State":0,"Value":0,"String":"","ValueMap":null,"ValueArray":[0,0,0,0],"ValueSliceList":null,"MapMap":null,"MapList":null,"MapSList":null,"SListMap":null,"SListList":null,"SListSList":null,"ListMap":[null,null,null,null],"ListList":[[0,0,0,0],[0,0,0,0],[0,0,0,0],[0,0,0,0]],"ListSList":[null,null,null,null]},"Op":1}]`)
	pos := bytes.IndexByte(data, byte(' '))
	assert.EqualValues(t, true, pos != -1)
	name := string(data[:pos])
	persistData := data[pos+1:]

	m := core.GetIPersistByName(name)
	assert.EqualValues(t, true, manage != nil)
	_ = m.RecoverBomb(persistData)

	engine := model.GetDB()
	_, _ = engine.QueryString("delete from persist;")
}

func TestRecoverBombFile(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	data := []byte(`Persist [{"Data":{"Uid":1,"Id":0,"State":0,"Value":0,"String":"","ValueMap":null,"ValueArray":[0,0,0,0],"ValueSliceList":null,"MapMap":null,"MapList":null,"MapSList":null,"SListMap":null,"SListList":null,"SListSList":null,"ListMap":[null,null,null,null],"ListList":[[0,0,0,0],[0,0,0,0],[0,0,0,0],[0,0,0,0]],"ListSList":[null,null,null,null]},"Op":1},{"Data":{"Uid":2,"Id":0,"State":0,"Value":0,"String":"","ValueMap":null,"ValueArray":[0,0,0,0],"ValueSliceList":null,"MapMap":null,"MapList":null,"MapSList":null,"SListMap":null,"SListList":null,"SListSList":null,"ListMap":[null,null,null,null],"ListList":[[0,0,0,0],[0,0,0,0],[0,0,0,0],[0,0,0,0]],"ListSList":[null,null,null,null]},"Op":1},{"Data":{"Uid":3,"Id":0,"State":0,"Value":0,"String":"","ValueMap":null,"ValueArray":[0,0,0,0],"ValueSliceList":null,"MapMap":null,"MapList":null,"MapSList":null,"SListMap":null,"SListList":null,"SListSList":null,"ListMap":[null,null,null,null],"ListList":[[0,0,0,0],[0,0,0,0],[0,0,0,0],[0,0,0,0]],"ListSList":[null,null,null,null]},"Op":1},{"Data":{"Uid":4,"Id":0,"State":0,"Value":0,"String":"","ValueMap":null,"ValueArray":[0,0,0,0],"ValueSliceList":null,"MapMap":null,"MapList":null,"MapSList":null,"SListMap":null,"SListList":null,"SListSList":null,"ListMap":[null,null,null,null],"ListList":[[0,0,0,0],[0,0,0,0],[0,0,0,0],[0,0,0,0]],"ListSList":[null,null,null,null]},"Op":1}]`)
	pos := bytes.IndexByte(data, byte(' '))
	assert.EqualValues(t, true, pos != -1)
	name := string(data[:pos])
	persistData := data[pos+1:]

	m := core.GetIPersistByName(name)
	assert.EqualValues(t, true, manage != nil)
	err := m.RecoverBomb(persistData)
	assert.EqualValues(t, true, err == nil)
	// 触发一个sql error日志
	err = m.RecoverBomb(persistData)
	assert.EqualValues(t, true, err != nil)

	_ = model.GPersistManager.LoadAll()
	assert.EqualValues(t, true, len(model.GPersistManager.GetPersistsByUid(1)) == 1)

	engine := model.GetDB()
	_, _ = engine.QueryString("delete from persist;")
	log.Println()
}

func TestSyncDataError(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	pMem := model.NewPersist(uid, 1)
	pMem.Value = 1
	_ = model.GPersistManager.MarkUpdate(pMem)
	// missing mark
	pMem.Value = 2

	_ = model.GPersistManager.Unload(uid)
	manage.Exit()
	manage.Run()

	// 模拟下次启动数据不一致
	_ = model.GPersistManager.Load(uid)

	pDB := model.GPersistManager.GetPersistByUidId(uid, 1)
	assert.EqualValues(t, 1, pDB.Value)

}

func TestSyncData(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	pMem := model.NewPersist(uid, 1)
	pMem.Value = 1
	_ = model.GPersistManager.MarkUpdate(pMem)
	_ = model.NewPersist(uid, 2)
	_ = model.NewPersist(uid+1, 1)
	_ = model.NewPersist(uid+1, 2)
	_ = model.NewPersist(uid+2, 1)
	// missing mark
	pMem.Value = 2

	manage.Exit()

	_ = manage.SyncDataPersist()

	manage.Run()

	_ = model.GPersistManager.Unload(uid)

	manage.Exit()

	// 模拟下次启动数据不一致
	_ = model.GPersistManager.Load(uid)

	pDB := model.GPersistManager.GetPersistByUidId(uid, 1)
	assert.EqualValues(t, pMem.Value, pDB.Value)

}
