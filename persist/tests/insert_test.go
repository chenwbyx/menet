package tests

import (
	"github.com/stretchr/testify/assert"
	"menet/persist/core"
	"menet/persist/tests/model"
	"menet/persist/tests/protocol"
	"sync"
	"sync/atomic"
	"testing"
)

func TestNewPersistGlobal(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	p1 := model.NewPersist(uid, 1)
	assert.EqualValues(t, p1, model.GPersistManager.GetPersistByUidId(uid, 1))
}

func TestNewPersistGlobalUnload(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	p1 := model.NewPersist(uid, 1)
	assert.EqualValues(t, (*model.Persist)(nil), p1)
	assert.EqualValues(t, (*model.Persist)(nil), model.GPersistManager.GetPersistByUidId(uid, 1))
}

func TestNewPersistGlobalRepeated(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)
	uid := int32(1)

	_ = model.GPersistManager.Load(uid)
	p1 := model.NewPersist(uid, 1)
	p2 := model.NewPersist(uid, 1)
	assert.EqualValues(t, true, p1 != nil)
	assert.EqualValues(t, p2, p1)
	assert.EqualValues(t, p1, model.GPersistManager.GetPersistByUidId(uid, 1))
}

func TestNewPersist(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	cls := &protocol.Persist{
		Uid:      uid,
		Id:       1,
		ValueMap: map[int32]bool{1: true, 2: false},
		SListMap: []map[int32]bool{{1: true, 2: false}, {3: true, 4: false}},
	}
	err := model.GPersistManager.NewPersist(cls)
	assert.EqualValues(t, nil, err)
	assert.EqualValues(t, true, model.GPersistManager.GetPersistByUidId(uid, 1) != nil)
}

func TestNewPersistUnload(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	p1 := &protocol.Persist{
		Uid:      uid,
		Id:       1,
		ValueMap: map[int32]bool{1: true, 2: false},
		SListMap: []map[int32]bool{{1: true, 2: false}, {3: true, 4: false}},
	}
	err := model.GPersistManager.NewPersist(p1)
	assert.EqualValues(t, core.EPersistErrorNotInMemory, err)
	assert.EqualValues(t, (*model.Persist)(nil), model.GPersistManager.GetPersistByUidId(uid, 1))
}

func TestNewPersistRepeated(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)
	p1 := &protocol.Persist{
		Uid:      uid,
		Id:       1,
		Value:    1,
		ValueMap: map[int32]bool{1: true, 2: false},
		SListMap: []map[int32]bool{{1: true, 2: false}, {3: true, 4: false}},
	}
	p2 := &protocol.Persist{
		Uid:      uid,
		Id:       1,
		Value:    2,
		ValueMap: map[int32]bool{1: true, 2: false},
		SListMap: []map[int32]bool{{1: true, 2: false}, {3: true, 4: false}},
	}
	err := model.GPersistManager.NewPersist(p1)
	assert.EqualValues(t, nil, err)
	err = model.GPersistManager.NewPersist(p2)
	assert.EqualValues(t, core.EPersistErrorAlreadyExist, err)
	assert.EqualValues(t, true, model.GPersistManager.GetPersistByUidId(uid, 1) != nil)
	assert.EqualValues(t, true, model.GPersistManager.GetPersistByUidId(uid, 1) == p1)
	assert.EqualValues(t, true, model.GPersistManager.GetPersistByUidId(uid, 1) != p2)
}

func TestConcurrenceNew(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	id := int32(1)
	num := 100
	wg := sync.WaitGroup{}
	wg.Add(num)

	_ = model.GPersistManager.Load(uid)
	var sum int32

	for i := int32(num); i > 0; i-- {
		go func(i int32) {
			cls := &protocol.Persist{Uid: 1, Id: 1, Value: i}
			err := model.GPersistManager.NewPersist(cls)
			if err == nil {
				atomic.AddInt32(&sum, 1)
				assert.EqualValues(t, i, model.GPersistManager.GetPersistByUidId(uid, id).Value)
			}
			wg.Done()
		}(i)
	}
	assert.EqualValues(t, 1, atomic.LoadInt32(&sum))
	wg.Wait()
}
