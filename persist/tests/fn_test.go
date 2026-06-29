//go:build integration

package tests

import (
	"menet/persist/tests/model"
	"sync"
)

type ManageTest struct {
	PersistManager    *model.PersistManager
	OldPersistManager *model.PersistManager
}

func (m *ManageTest) Exit() {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	m.PersistManager.Exit(wg)
	wg.Wait()
}

func (m *ManageTest) Run() {
	m.PersistManager.Run()
}

func (m *ManageTest) SyncDataPersist() error {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	return m.PersistManager.SyncData(wg)
}

func SetupFunction() *ManageTest {
	manage := &ManageTest{}
	manage.PersistManager = model.NewPersistManager(model.GetDB())
	manage.OldPersistManager = model.GPersistManager

	model.GPersistManager = manage.PersistManager
	manage.PersistManager.Run()
	return manage
}

func TeardownFunction(manage *ManageTest) {
	wg := &sync.WaitGroup{}
	_ = manage.PersistManager.LoadAll()
	manage.PersistManager.DeleteAll()
	_ = manage.PersistManager.UnloadAll()
	wg.Add(1)
	manage.PersistManager.Exit(wg)
	model.GPersistManager = manage.OldPersistManager
	wg.Wait()
}
