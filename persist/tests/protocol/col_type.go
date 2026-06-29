//go:build ignore

package protocol

import (
	"log"
	"time"
)

//type DirectGiftBag struct {
//	Uid         int32 `xorm:"pk"        hash:"group=1;unique=0" hash:"group=2;unique=1" hash:"group=3;unique=0"`
//	Id          int32 `xorm:"pk"        hash:"group=1;unique=0" hash:"group=2;unique=1"` // 礼包配置ID
//	PackId      int32 `xorm:""                                  hash:"group=2;unique=1"` // 礼包ID
//	RefreshTime int64 `xorm:""`                                                          // 下一次刷新时间戳
//	BuyCount    int32 `xorm:""`                                                          // 已购买次数
//	AwardCount  int32 `xorm:""`                                                          // 领奖次数
//}

type Persist struct {
	Uid            int32                    `xorm:"pk"        hash:"group=1;unique=0" hash:"group=2;unique=1" hash:"group=4;unique=0"`
	Id             int32                    `xorm:"pk"                                hash:"group=2;unique=1"`
	State          int32                    `xorm:"default 0" hash:"group=1;unique=0" hash:"group=3;unique=0" hash:"group=5;unique=0"` // 0:new   1:update   2:delete
	State1         int32                    `xorm:"default 0"                         hash:"group=3;unique=0"`
	Time           int64                    `xorm:"-"` // 时间, 不写回数据库
	Value          int32                    `xorm:"default 0"`
	String         string                   `xorm:"varchar(8)"`
	ValueMap       map[int32]bool           `xorm:"json"`
	ValueArray     [4]int32                 `xorm:"json"`
	ValueSliceList [][2]int32               `xorm:"json"`
	MapMap         map[int32]map[int32]bool `xorm:"json"`
	MapList        map[int32][4]int32       `xorm:"json"`
	MapSList       map[int32][]int32        `xorm:"json"`
	SListMap       []map[int32]bool         `xorm:"json"`
	SListList      [][4]int32               `xorm:"json"`
	SListSList     [][]int32                `xorm:"json"`
	ListMap        [4]map[int32]bool        `xorm:"json"`
	ListList       [4][4]int32              `xorm:"json"`
	ListSList      [4][]int32               `xorm:"json"`
}

type InnerSyncMap struct {
	SyncMap *MapInt32Int32
}

type PersistSyncMap struct {
	Uid            int32                  `xorm:"pk"        hash:"group=1;unique=1"`
	SyncMap        *MapInt32Int32         `xorm:"json"`
	StructSyncMap  InnerSyncMap           `xorm:"json"`
	SyncMapSyncMap *MapInt32MapInt32Int32 `xorm:"json"`
}

func (m *Persist) CopyTo(t *Persist) {
	*t = *m
}

func (m *Persist) Overload(queueSize int, lastWriteBackTime time.Duration) {
	log.Println("Overload", queueSize, " ", lastWriteBackTime)
}
