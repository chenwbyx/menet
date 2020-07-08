package ranking

import (
	"time"
)

type RankItem struct {
	Id         int64
	Score      int64
	updateTime int64
}

type tScore struct {
	score      int64
	updateTime int64
}

type RankList struct {
	dict  map[int64]*RankItem
	ranks *SkipList
}

func NewRankList() *RankList {
	rl := &RankList{}
	rl.dict = make(map[int64]*RankItem)
	rl.ranks = NewSkipList()
	return rl
}

func (rl *RankList) Set(id, score int64) {
	now := time.Now().Unix()
	if old, ok := rl.dict[id]; ok {
		if old.Score == score {
			return
		}
		r := &RankItem{Id: id, Score: score, updateTime: now}
		rl.ranks.Remove(old)
		rl.ranks.Insert(r)
		rl.dict[id] = r
	} else {
		item := &RankItem{Id: id, Score: score, updateTime: now}
		rl.ranks.Insert(item)
		rl.dict[id] = item
	}
}

func (rl *RankList) Delete(id int64) {
	if v, ok := rl.dict[id]; ok {
		rl.ranks.Remove(v)
		delete(rl.dict, id)
	}
}

func (rl *RankList) Rank(id int64) int {
	if v, ok := rl.dict[id]; ok {
		return rl.ranks.GetRank(v)
	}
	return -1
}

func (rl *RankList) Score(id int64) int64 {
	if v, ok := rl.dict[id]; ok {
		return v.Score
	}
	return 0
}

// precondition: start > 0 and (start <= stop or stop == -1)
// return inclusive range [start, stop]
func (rl *RankList) Range(start, stop int) []*RankItem {
	var items []*RankItem
	if start <= 0 || (stop < start && stop != -1) {
		return items
	}
	curNode := rl.ranks.GetNodeByRank(start)
	i := start
	if stop == -1 {
		stop = rl.ranks.curLength
	}
	for curNode != nil && i <= stop {
		items = append(items, curNode.Value)
		curNode = curNode.next[0]
		i++
	}
	return items
}

func (rl *RankList) GetIdByRank(r int) int64 {
	item := rl.ranks.GetByRank(r)
	if item != nil {
		return item.Id
	}
	return -1
}

func (rl *RankList) Count() int {
	if rl.ranks.curLength != len(rl.dict) {
		panic("rank list bug ")
	}
	return len(rl.dict)
}
