package tests

import (
	"github.com/stretchr/testify/assert"
	"menet/persist/core"
	"menet/persist/tests/model"
	"testing"
	"time"
)

func TestDeletePersist(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	p1 := model.NewPersist(uid, 1)
	_ = model.GPersistManager.DeletePersist(p1)
	assert.EqualValues(t, true, model.GPersistManager.GetPersistByUidId(uid, 1) == nil)
	_ = model.GPersistManager.NewPersist(p1)
	assert.EqualValues(t, true, model.GPersistManager.GetPersistByUidId(uid, 1) != nil)
}

func TestDeletePersistUnload(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	p1 := model.NewPersist(uid, 1)

	// 模拟下次启动数据不一致
	_ = model.GPersistManager.Unload(uid)
	manage.Exit()
	manage.Run()

	err := model.GPersistManager.DeletePersist(p1)
	assert.EqualValues(t, core.EPersistErrorNotInMemory, err)

	_ = model.GPersistManager.Load(uid)

	err = model.GPersistManager.DeletePersist(p1)
	assert.EqualValues(t, core.EPersistErrorOutOfDate, err)

	pNew := model.GPersistManager.GetPersistByUidId(uid, 1)
	err = model.GPersistManager.DeletePersist(pNew)
	assert.EqualValues(t, nil, err)
	assert.EqualValues(t, (*model.Persist)(nil), model.GPersistManager.GetPersistByUidId(uid, 1))
}

func TestDeletePersistUnloadFail(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	done := make(chan bool)
	subCoroutineRun := make(chan bool)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)

	const EMaxLength = 100
	var pArray [EMaxLength]*model.Persist
	// 创建
	for i := int32(0); i < EMaxLength; i++ {
		p := model.NewPersist(uid, i)
		pArray[i] = p
		assert.EqualValues(t, true, p != nil)
	}

	pAll := model.GPersistManager.GetAll()
	assert.EqualValues(t, EMaxLength, len(pAll))

	// 一个协程持有p1不间断的修改数据.  导出后, 内存中有两份p1, 该函数内的旧值成为脏数据, 会污染新值
	go func(pArray [EMaxLength]*model.Persist) {
		idx := 0
		err := model.GPersistManager.DeletePersist(pArray[idx])
		assert.EqualValues(t, nil, err)
		subCoroutineRun <- true
		hasUnload := false
		for {
			select {
			case <-done:
				return
			case <-time.Tick(time.Millisecond * 10):
				idx += 1
				err := model.GPersistManager.DeletePersist(pArray[idx])
				// 一旦检测到导出后, mark操作应当失败
				if model.GPersistManager.LoadState(uid) != model.EPersistLoadStateMemory {
					hasUnload = true
				}
				if hasUnload {
					assert.EqualValues(t, core.EPersistErrorNotInMemory, err)
				}
			}
		}
	}(pArray)

	<-subCoroutineRun
	// 导出 再次导入
	_ = model.GPersistManager.Unload(uid)
	manage.Exit()
	manage.Run()
	_ = model.GPersistManager.Load(uid)
	pAllNew1 := model.GPersistManager.GetAll()

	//log.Println("show value", len(pAll), " ", len(pAllNew1))

	// 导出 再次导入
	_ = model.GPersistManager.Unload(uid)
	manage.Exit()
	manage.Run()
	_ = model.GPersistManager.Load(uid)
	pAllNew2 := model.GPersistManager.GetAll()

	//log.Println("show value", len(pAll), " ", len(pAllNew1), " ", len(pAllNew2))

	// 导出后, 内存中的数据, 应该丢弃.  如需修改应当重新查询
	assert.EqualValues(t, len(pAllNew1), len(pAllNew2))

	done <- true

	for i := int32(0); i < EMaxLength; i++ {
		err := model.GPersistManager.DeletePersist(pArray[i])
		if err != nil {
			assert.EqualValues(t, core.EPersistErrorOutOfDate, err)
		}
	}
}

func TestDeleteAllPersist(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	model.NewPersist(uid, 1)
	model.NewPersist(uid, 2)
	model.NewPersist(uid, 3)
	model.GPersistManager.DeleteAll()
	assert.EqualValues(t, true, model.GPersistManager.GetPersistByUidId(uid, 1) == nil)
	assert.EqualValues(t, true, model.GPersistManager.GetPersistByUidId(uid, 2) == nil)
	assert.EqualValues(t, true, model.GPersistManager.GetPersistByUidId(uid, 3) == nil)
	model.NewPersist(uid, 1)
	assert.EqualValues(t, true, model.GPersistManager.GetPersistByUidId(uid, 1) != nil)
}
