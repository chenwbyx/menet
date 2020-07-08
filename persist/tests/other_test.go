package tests

import (
	"github.com/stretchr/testify/assert"
	"menet/persist/tests/model"
	"sync"
	"testing"
)

func TestRun(t *testing.T) {
	defer model.Exit()
	assert.EqualValues(t, true, model.Dead())
	model.Run()
	assert.EqualValues(t, false, model.Dead())
}

func TestExit(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	model.GPersistManager.Exit(&wg)
	wg.Wait()
}

func TestExitGlobal(t *testing.T) {
	model.Exit()
}

func TestRestart(t *testing.T) {
	defer model.Exit()
	assert.EqualValues(t, false, !model.Dead())
	model.Run()
	assert.EqualValues(t, true, !model.Dead())
	model.Exit()
	assert.EqualValues(t, false, !model.Dead())
	model.Run()
	assert.EqualValues(t, true, !model.Dead())
}

func TestOverload(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)

	p1 := model.NewPersist(uid, 1)
	for i := int32(1); i < 1000*50; i++ {
		//fmt.Println(i, time.Now().Unix())
		p1.Value = i
		_ = model.GPersistManager.MarkUpdate(p1)
	}
}

func TestGetAllPersist(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	_ = model.GPersistManager.Load(2)
	model.NewPersist(uid, 1)
	model.NewPersist(uid, 2)
	model.NewPersist(uid, 3)
	model.NewPersist(2, 1)
	model.NewPersist(2, 2)

	// unload
	model.NewPersist(3, 1)
	assert.EqualValues(t, 5, len(model.GPersistManager.GetAll()))
}
