package ranking

import (
	"testing"
	"time"
	"fmt"
)

func TestSkipList(t *testing.T) {
	sl := NewSkipList()
	now := time.Now()
	e1 := &RankItem{Id:1,Score:1000,updateTime:now.Unix()}
	sl.Insert(e1)
	e2 := &RankItem{Id: 2, Score: 1000, updateTime: now.Add(-time.Second).Unix()}
	sl.Insert(e2)
	e3 := &RankItem{Id: 3, Score: 99, updateTime: now.Unix()}
	sl.Insert(e3)

	fmt.Println("compare", lessOrEq(e1, e1))

	if sl.GetRank(e1) != 2 {
		t.Fail()
	}
	if sl.GetRank(e2) != 1 {
		t.Fail()
	}
	if sl.GetRank(e3) != 3 {
		t.Fail()
	}
	sl.Dprint()
	sl.Remove(e1)
	fmt.Println(sl.GetRank(e1))
	sl.Dprint()
}

func TestRanking(t *testing.T) {
	rl := NewRankList()
	for i := 0; i < 100; i++ {
		rl.Set(1, int64(i))
		rl.Set(2, int64(i))
	}
	if rl.Count() != 2 {
		t.Fail()
	}
}