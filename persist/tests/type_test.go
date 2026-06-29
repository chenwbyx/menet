//go:build integration

package tests

import (
	"github.com/stretchr/testify/assert"
	"menet/persist/tests/model"
	"menet/persist/tests/protocol"
	"testing"
)

func TestNewSyncMapNull(t *testing.T) {
	manage := SetupFunction()
	defer TeardownFunction(manage)

	uid := int32(1)

	_ = model.GPersistSyncMapManager.Load(uid)
	cls := model.GPersistSyncMapManager.GetPersistSyncMapByUid(uid)
	assert.EqualValues(t, (*model.PersistSyncMap)(nil), cls)
	cls = &protocol.PersistSyncMap{
		Uid:            uid,
		SyncMap:        nil,
		StructSyncMap:  protocol.InnerSyncMap{},
		SyncMapSyncMap: nil,
	}
	assert.EqualValues(t, (*protocol.MapInt32Int32)(nil), cls.SyncMap)
	assert.EqualValues(t, (*protocol.MapInt32Int32)(nil), cls.StructSyncMap.SyncMap)
	_ = model.GPersistSyncMapManager.NewPersistSyncMap(cls)

	manage.Exit()
	manage.Run()
	_ = model.GPersistSyncMapManager.Load(uid)

	cls = model.GPersistSyncMapManager.GetPersistSyncMapByUid(uid)
	assert.EqualValues(t, (*protocol.MapInt32Int32)(nil), cls.SyncMap)
	_ = model.GPersistSyncMapManager.MarkUpdate(cls)

}
