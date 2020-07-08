package tests

import (
	"github.com/stretchr/testify/assert"
	"menet/persist/core"
	"menet/persist/tests/model"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestMarkPersist(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	p1 := model.NewPersist(uid, 1)
	assert.EqualValues(t, 0, model.GPersistManager.GetPersistByUidId(uid, 1).Value)
	p1.Value = 100
	assert.EqualValues(t, 100, model.GPersistManager.GetPersistByUidId(uid, 1).Value)
	_ = model.GPersistManager.MarkUpdate(p1)

	manage.Exit()
	manage.Run()
	_ = model.GPersistManager.Load(uid)

	assert.EqualValues(t, 100, model.GPersistManager.GetPersistByUidId(uid, 1).Value)
}

func TestMarkPersistUnload(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)

	// 创建
	p1 := model.NewPersist(uid, 1)
	assert.EqualValues(t, 0, model.GPersistManager.GetPersistByUidId(uid, 1).Value)

	// 导出 再次导入
	_ = model.GPersistManager.Unload(uid)
	manage.Exit()
	manage.Run()

	p1.Value += 100
	err := model.GPersistManager.MarkUpdate(p1)
	assert.EqualValues(t, core.EPersistErrorNotInMemory, err)

	_ = model.GPersistManager.Load(uid)

	p1.Value += 100
	err = model.GPersistManager.MarkUpdate(p1)
	assert.EqualValues(t, core.EPersistErrorOutOfDate, err)

	pNew := model.GPersistManager.GetPersistByUidId(uid, 1)
	pNew.Value += 100
	err = model.GPersistManager.MarkUpdate(pNew)
	assert.EqualValues(t, nil, err)
	assert.EqualValues(t, 100, model.GPersistManager.GetPersistByUidId(uid, 1).Value)
}

func TestMarkPersistUnloadFail(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	done := make(chan bool)
	subCoroutineRun := make(chan bool)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)

	// 创建
	p1 := model.NewPersist(uid, 1)
	assert.EqualValues(t, 0, model.GPersistManager.GetPersistByUidId(uid, 1).Value)

	// 一个协程持有p1不间断的修改数据.  导出后, 内存中有两份p1, 该函数内的旧值成为脏数据, 会污染新值
	go func(pDirty *model.Persist) {
		pDirty.Value += 1
		err := model.GPersistManager.MarkUpdate(pDirty)
		assert.EqualValues(t, nil, err)
		subCoroutineRun <- true
		hasUnload := false
		for {
			select {
			case <-done:
				return
			case <-time.Tick(time.Millisecond * 10):
				pDirty.Value += 1
				err := model.GPersistManager.MarkUpdate(pDirty)
				// 一旦检测到导出后, mark操作应当失败
				if model.GPersistManager.LoadState(uid) != model.EPersistLoadStateMemory {
					hasUnload = true
				}
				if hasUnload {
					assert.EqualValues(t, core.EPersistErrorNotInMemory, err)
				}
			}
		}
	}(p1)

	<-subCoroutineRun
	// 导出 再次导入
	_ = model.GPersistManager.Unload(uid)
	manage.Exit()
	manage.Run()
	_ = model.GPersistManager.Load(uid)
	pNew1 := model.GPersistManager.GetPersistByUidId(uid, 1)

	//log.Println("show value", p1.Value, " ", pNew1.Value)

	// 导出 再次导入
	_ = model.GPersistManager.Unload(uid)
	manage.Exit()
	manage.Run()
	_ = model.GPersistManager.Load(uid)
	pNew2 := model.GPersistManager.GetPersistByUidId(uid, 1)

	//log.Println("show value", p1.Value, " ", pNew1.Value, " ", pNew2.Value)

	// 导出后, 内存中的数据, 应该丢弃.  如需修改应当重新查询
	assert.EqualValues(t, pNew2.Value, pNew1.Value)

	done <- true

	err := model.GPersistManager.MarkUpdate(p1)
	assert.EqualValues(t, core.EPersistErrorOutOfDate, err)
}

func TestMarkPersistRepeated(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	p1 := model.NewPersist(uid, 1)
	assert.EqualValues(t, 0, model.GPersistManager.GetPersistByUidId(uid, 1).Value)
	p1.Value = 100
	assert.EqualValues(t, 100, model.GPersistManager.GetPersistByUidId(uid, 1).Value)
	_ = model.GPersistManager.MarkUpdate(p1)
	_ = model.GPersistManager.MarkUpdate(p1)

	manage.Exit()
	manage.Run()
	_ = model.GPersistManager.Load(uid)

	assert.EqualValues(t, 100, model.GPersistManager.GetPersistByUidId(uid, 1).Value)
}

func TestConcurrenceMarkPersist(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	p1 := model.NewPersist(uid, 1)
	assert.EqualValues(t, 0, model.GPersistManager.GetPersistByUidId(uid, 1).Value)

	const ESumValue = 1000

	changeValue := func(wg *sync.WaitGroup, p *model.Persist) {
		defer wg.Done()
		for {
			oldValue := atomic.LoadInt32(&p.Value)
			if oldValue >= ESumValue {
				return
			}
			_ = model.GPersistManager.MarkUpdate(p)
			if atomic.CompareAndSwapInt32(&p.Value, oldValue, oldValue+1) {
				err := model.GPersistManager.MarkUpdate(p)
				assert.EqualValues(t, nil, err)
			}
		}
	}

	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go changeValue(wg, p1)
	}

	manage.Exit()
	manage.Run()
	_ = model.GPersistManager.Load(uid)
	wg.Wait()
	assert.EqualValues(t, ESumValue, model.GPersistManager.GetPersistByUidId(uid, 1).Value)
}

func TestChangeIndexKey1(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	p1 := model.NewPersist(uid, 1)
	model.GPersistManager.SetIndexKeyUidState(p1, 1)
	p2 := model.NewPersist(uid, 2)
	model.GPersistManager.SetIndexKeyUidState(p2, 1)
	assert.EqualValues(t, 2, len(model.GPersistManager.GetPersistsByUidState(uid, 1)))
	model.GPersistManager.SetIndexKeyUidState(p1, 2)
	assert.EqualValues(t, 1, len(model.GPersistManager.GetPersistsByUidState(uid, 1)))
	assert.EqualValues(t, 1, len(model.GPersistManager.GetPersistsByUidState(uid, 2)))
}

func TestChangeIndexKey2(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	p1 := model.NewPersist(uid, 1)
	model.GPersistManager.SetIndexKeyState(p1, 1)
	p2 := model.NewPersist(uid, 2)
	model.GPersistManager.SetIndexKeyState(p2, 1)
	assert.EqualValues(t, 2, len(model.GPersistManager.GetPersistsByUidState(uid, 1)))
	model.GPersistManager.SetIndexKeyState(p1, 2)
	assert.EqualValues(t, 1, len(model.GPersistManager.GetPersistsByUidState(uid, 1)))
	assert.EqualValues(t, 1, len(model.GPersistManager.GetPersistsByUidState(uid, 2)))
}

func TestChangeIndexKey3(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	p1 := model.NewPersist(uid, 1)
	model.GPersistManager.SetIndexKeyStateState1(p1, 1, 1)
	p2 := model.NewPersist(uid, 2)
	model.GPersistManager.SetIndexKeyStateState1(p2, 1, 2)
	assert.EqualValues(t, 2, len(model.GPersistManager.GetPersistsByState(1)))
	assert.EqualValues(t, 1, len(model.GPersistManager.GetPersistsByStateState1(1, 2)))
	model.GPersistManager.SetIndexKeyStateState1(p2, 1, 1)
	assert.EqualValues(t, 2, len(model.GPersistManager.GetPersistsByStateState1(1, 1)))
}

func TestMissingMark(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	pMem := model.NewPersist(uid, 1)
	pMem.Value = 1
	_ = model.GPersistManager.MarkUpdate(pMem)

	// missing mark
	pMem.Value = 2

	// 模拟下次启动数据不一致
	_ = model.GPersistManager.Unload(uid)
	manage.Exit()
	manage.Run()

	// 重新导入数据
	_ = model.GPersistManager.Load(uid)
	pDB := model.GPersistManager.GetPersistByUidId(uid, 1)
	assert.EqualValues(t, true, pDB.Value != pMem.Value)
	assert.EqualValues(t, true, pDB != pMem)
}

func TestMarkError(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)

	p1 := model.NewPersist(uid, 1)
	err := model.GPersistManager.DeletePersist(p1)
	assert.EqualValues(t, nil, err)
	err = model.GPersistManager.MarkUpdate(p1)
	assert.EqualValues(t, core.EPersistErrorOutOfDate, err)

}
