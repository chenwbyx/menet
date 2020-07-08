package tests

import (
	"github.com/stretchr/testify/assert"
	"log"
	"menet/persist/tests/model"
	"sync"
	"testing"
)

func TestConcurrenceLoad(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	id := int32(1)
	num := 10
	wg := sync.WaitGroup{}
	wg.Add(num)
	err := model.GPersistManager.Load(uid)
	assert.EqualValues(t, nil, err)

	model.NewPersist(uid, id)

	err = model.GPersistManager.Unload(uid)
	assert.EqualValues(t, nil, err)

	for i := num; i > 0; i-- {
		go func(i int) {
			err = model.GPersistManager.Load(uid)
			assert.EqualValues(t, nil, err)
			assert.EqualValues(t, true, model.GPersistManager.GetPersistByUidId(uid, id) != nil)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestLoadAll(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	model.NewPersist(uid, 1)
	model.NewPersist(uid, 2)
	_ = model.GPersistManager.Unload(uid)

	manage.Exit()
	manage.Run()
	_ = model.GPersistManager.LoadAll()

	assert.EqualValues(t, true, len(model.GPersistManager.GetPersistsByUid(uid)) == 2)
}

func TestLoad(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	model.NewPersist(uid, 1)
	model.NewPersist(uid, 2)
	_ = model.GPersistManager.Unload(uid)

	manage.Exit()
	manage.Run()
	_ = model.GPersistManager.Load(uid)

	assert.EqualValues(t, true, model.GPersistManager.GetPersistByUidId(uid, 1) != nil)
	assert.EqualValues(t, true, model.GPersistManager.GetPersistByUidId(uid, 2) != nil)
}

// warning: 导出后,继续修改对象会失败
func TestUnloadAfterUpdate(t *testing.T) {
	pt := func(flag string) {
		//log.Println(flag,
		//	"stateUid:", model.GPersistManager.LoadState(int32(1)),
		//)
	}
	manage := SetupFunction()
	defer TeardownFunction(manage)

	subCoroutineRun := make(chan bool)

	num := int32(5 * 1000)
	realNum := int32(0)
	uid := int32(1)
	wg := sync.WaitGroup{}
	_ = model.GPersistManager.Load(uid)
	pt("debug A")
	model.NewPersist(uid, 1)

	// 疯狂修改数据
	wg.Add(1)
	go func() {
		subCoroutineRun <- true
		for i := int32(0); i < num; {
			if model.GPersistManager.LoadState(uid) == model.EPersistLoadStateDisk {
				pt("debug C")
				_ = model.GPersistManager.Load(uid)
				pt("debug D")
			}
			p := model.GPersistManager.GetPersistByUidId(uid, 1)
			if p != nil {
				realNum = i
				p.Value = i
				err := model.GPersistManager.MarkUpdate(p)
				if err != nil {
				} else {
					i++
				}
			} else {
			}
		}
		wg.Done()
	}()

	<-subCoroutineRun

	err := model.GPersistManager.Unload(uid)
	pt("debug B")
	if err != nil {
		log.Println("unload error ", err)
	}

	manage.Exit()
	manage.Run()

	wg.Wait()

	err = model.GPersistManager.Load(uid)
	assert.EqualValues(t, nil, err)
	p1 := model.GPersistManager.GetPersistByUidId(uid, 1)

	assert.EqualValues(t, true, p1 != nil)
	assert.EqualValues(t, realNum, p1.Value)
}
