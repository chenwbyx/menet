package tests

import (
	"github.com/stretchr/testify/assert"
	"github.com/wbexwbex/queue"
	"menet/persist/tests/model"
	"menet/persist/tests/protocol"
	"strconv"
	"testing"
	"time"
)

func TestMergeSqlError(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)
	manage.Exit()

	uid := int32(1)
	syncQueue := queue.New()
	cls := &protocol.Persist{
		Uid:      uid,
		Id:       1,
		ValueMap: map[int32]bool{1: true, 2: false},
		SListMap: []map[int32]bool{{1: true, 2: false}, {3: true, 4: false}},
	}
	syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpInsert})
	syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpDelete})
	syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpUpdate})

	model.GPersistManager.MergeQueue(syncQueue)
	assert.EqualValues(t, 1, syncQueue.Length())
	assert.EqualValues(t, 0, len(manage.PersistManager.InsertQueue))

}

func TestMergeSqlInsertInsert(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)
	manage.Exit()

	uid := int32(1)
	syncQueue := queue.New()
	cls := &protocol.Persist{
		Uid:      uid,
		Id:       0,
		ValueMap: map[int32]bool{1: true, 2: false},
		SListMap: []map[int32]bool{{1: true, 2: false}, {3: true, 4: false}},
	}

	syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpInsert})

	for i := int32(1); i < 100; i++ {
		pNew := &protocol.Persist{
			Uid:      uid,
			Id:       i,
			ValueMap: map[int32]bool{1: true, 2: false},
			SListMap: []map[int32]bool{{1: true, 2: false}, {3: true, 4: false}},
		}
		syncQueue.Add(&model.PersistSync{Data: pNew, Op: model.EPersistOpInsert})
	}

	model.GPersistManager.MergeQueue(syncQueue)

	assert.EqualValues(t, 0, syncQueue.Length())
	assert.EqualValues(t, 100, len(manage.PersistManager.InsertQueue))
}

func TestMergeSqlInsertUpdate(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)
	manage.Exit()

	uid := int32(1)
	syncQueue := queue.New()
	cls := &protocol.Persist{
		Uid:      uid,
		Id:       1,
		ValueMap: map[int32]bool{1: true, 2: false},
		SListMap: []map[int32]bool{{1: true, 2: false}, {3: true, 4: false}},
	}

	syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpInsert})

	for i := int32(1); i < 100; i++ {
		syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpUpdate})
	}

	model.GPersistManager.MergeQueue(syncQueue)

	assert.EqualValues(t, 0, syncQueue.Length())
	assert.EqualValues(t, 1, len(manage.PersistManager.InsertQueue))
}

func TestMergeSqlInsertDelete(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)
	manage.Exit()

	uid := int32(1)
	syncQueue := queue.New()
	cls := &protocol.Persist{
		Uid:      uid,
		Id:       1,
		ValueMap: map[int32]bool{1: true, 2: false},
		SListMap: []map[int32]bool{{1: true, 2: false}, {3: true, 4: false}},
	}

	for i := int32(1); i < 100; i++ {
		syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpInsert})
		syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpDelete})
	}

	model.GPersistManager.MergeQueue(syncQueue)
	assert.EqualValues(t, 0, syncQueue.Length())
	assert.EqualValues(t, 0, len(manage.PersistManager.InsertQueue))
}

func TestMergeSqlUpdateInsert(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)
	manage.Exit()

	uid := int32(1)
	syncQueue := queue.New()

	for i := int32(1); i < 100; i++ {
		pNew := &protocol.Persist{
			Uid:      uid,
			Id:       i,
			ValueMap: map[int32]bool{1: true, 2: false},
			SListMap: []map[int32]bool{{1: true, 2: false}, {3: true, 4: false}},
		}
		syncQueue.Add(&model.PersistSync{Data: pNew, Op: model.EPersistOpUpdate})
		// 无法合并, 最终插入数据库失败
		syncQueue.Add(&model.PersistSync{Data: pNew, Op: model.EPersistOpInsert})
	}

	model.GPersistManager.MergeQueue(syncQueue)
	assert.EqualValues(t, 198, syncQueue.Length())
	assert.EqualValues(t, 0, len(manage.PersistManager.InsertQueue))

}

func TestMergeSqlUpdateUpdate(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)
	manage.Exit()

	uid := int32(1)
	syncQueue := queue.New()
	cls := &protocol.Persist{
		Uid:      uid,
		Id:       1,
		ValueMap: map[int32]bool{1: true, 2: false},
		SListMap: []map[int32]bool{{1: true, 2: false}, {3: true, 4: false}},
	}

	for i := int32(1); i < 100; i++ {
		syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpUpdate})
	}

	model.GPersistManager.MergeQueue(syncQueue)
	assert.EqualValues(t, 1, syncQueue.Length())
	assert.EqualValues(t, 0, len(manage.PersistManager.InsertQueue))

}

func TestMergeSqlUpdateDelete(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)
	manage.Exit()

	uid := int32(1)
	syncQueue := queue.New()
	cls := &protocol.Persist{
		Uid:      uid,
		Id:       1,
		ValueMap: map[int32]bool{1: true, 2: false},
		SListMap: []map[int32]bool{{1: true, 2: false}, {3: true, 4: false}},
	}

	syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpUpdate})
	syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpDelete})

	model.GPersistManager.MergeQueue(syncQueue)
	assert.EqualValues(t, 1, syncQueue.Length())
	assert.EqualValues(t, 0, len(manage.PersistManager.InsertQueue))
}

func TestMergeSqlDeleteInsert(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)
	manage.Exit()

	uid := int32(1)
	syncQueue := queue.New()
	cls := &protocol.Persist{
		Uid:      uid,
		Id:       1,
		ValueMap: map[int32]bool{1: true, 2: false},
		SListMap: []map[int32]bool{{1: true, 2: false}, {3: true, 4: false}},
	}

	for i := int32(1); i < 100; i++ {
		syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpDelete})
		syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpInsert})
	}

	model.GPersistManager.MergeQueue(syncQueue)
	assert.EqualValues(t, 1, syncQueue.Length())
	assert.EqualValues(t, 0, len(manage.PersistManager.InsertQueue))
}

func TestMergeSqlDeleteUpdate(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)
	manage.Exit()

	uid := int32(1)
	syncQueue := queue.New()
	cls := &protocol.Persist{
		Uid:      uid,
		Id:       1,
		ValueMap: map[int32]bool{1: true, 2: false},
		SListMap: []map[int32]bool{{1: true, 2: false}, {3: true, 4: false}},
	}

	for i := int32(1); i < 100; i++ {
		syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpDelete})
		// 无法合并, 最终插入数据库失败
		syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpUpdate})
	}

	model.GPersistManager.MergeQueue(syncQueue)
	assert.EqualValues(t, 198, syncQueue.Length())
	assert.EqualValues(t, 0, len(manage.PersistManager.InsertQueue))
}

func TestMergeSqlDeleteDelete(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)
	manage.Exit()

	uid := int32(1)
	syncQueue := queue.New()
	cls := &protocol.Persist{
		Uid:      uid,
		Id:       1,
		ValueMap: map[int32]bool{1: true, 2: false},
		SListMap: []map[int32]bool{{1: true, 2: false}, {3: true, 4: false}},
	}

	for i := int32(1); i < 100; i++ {
		syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpDelete})
		// 无法合并, 最终插入数据库失败
		syncQueue.Add(&model.PersistSync{Data: cls, Op: model.EPersistOpDelete})
	}

	model.GPersistManager.MergeQueue(syncQueue)
	assert.EqualValues(t, 198, syncQueue.Length())
	assert.EqualValues(t, 0, len(manage.PersistManager.InsertQueue))
}

func TestSqlInsert(t *testing.T) {
	insertNum := int32(201)

	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)
	_ = model.GPersistManager.Load(uid)

	//tBegin := time.Now()
	for i := int32(0); i < insertNum; i++ {
		p1 := &protocol.Persist{Uid: uid, Id: i}
		_ = model.GPersistManager.NewPersist(p1)
	}

	engine := model.GetDB()
	for {
		ret, err := engine.QueryString("select count(*) as num from persist")
		if err != nil {
			break
		}
		if ret[0]["num"] == strconv.Itoa(int(insertNum)) {
			//log.Println(time.Now().Sub(tBegin).String())
			break
		}

		time.Sleep(time.Second / 100)
	}

}
