package ranklist

import (
	"github.com/go-redis/redis"
	"strconv"
	"runtime"
	"log"
	"sync"
	"fmt"
)

type RankList struct {
	client *redis.Client
	rank   string
}

var gRedisMap sync.Map
func GetRankList(rank, addr string, db int) *RankList {
	r := new(RankList)
	redisAddr := fmt.Sprint(addr,":", db)
	if c, ok := gRedisMap.Load(redisAddr); ok {
		r.client = c.(*redis.Client)
	} else {
		newClient := redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: "",
			DB:       db,
		})
		ci, ok := gRedisMap.LoadOrStore(redisAddr, newClient)
		if ok {
			newClient.Close()
		}
		r.client = ci.(*redis.Client)
	}

	r.rank = rank

	return r
}

func (r *RankList) Set(uid int32, score int64) {
	cmd := r.client.ZAdd(r.rank, redis.Z{Score: float64(-score), Member: uid})
	err := cmd.Err()
	if err != nil {
		pc, file, line, ok := runtime.Caller(2)
		if ok {
			log.Printf("%s:%d %s\n", file, line, runtime.FuncForPC(pc))
		}
		log.Println(cmd, err)
	}
}

//var gRedisCount int32
//var redisTime int64
//func (r *RankList) Set(uid int32, score int64) {
//	cmd := r.client.ZAdd(r.rank, redis.Z{Score: float64(-score), Member: uid})
//	err := cmd.Err()
//	if err != nil {
//		pc, file, line, ok := runtime.Caller(2)
//		if ok {
//			log.Printf("%s:%d %s\n", file, line, runtime.FuncForPC(pc))
//		}
//		log.Println(cmd, err)
//	}
//	atomic.AddInt32(&gRedisCount, 1)
//	if atomic.LoadInt64(&redisTime)+5 < time.Now().Unix() {
//		log.Println("redis set count", atomic.LoadInt32(&gRedisCount))
//		atomic.StoreInt64(&redisTime, time.Now().Unix())
//	}
//}

func (r *RankList) SetRankItems(items ...RankItem) {
	var zs []redis.Z
	for _, x := range items {
		zs = append(zs, redis.Z{Score: float64(-x.Score), Member: x.Uid})
	}
	r.client.ZAdd(r.rank, zs...)
}

func (r *RankList) Add(uid int32, score int64) {
	r.client.ZIncr(r.rank, redis.Z{Score: float64(-score), Member: uid})
}

func (r *RankList) Rank(uid int32) int32 {
	rank, err := r.client.ZRank(r.rank, strconv.FormatInt(int64(uid), 10)).Result()
	if err != nil {
		return -1
	}
	return int32(rank) + 1
}

func (r *RankList) Score(uid int32) int64 {
	score, err := r.client.ZScore(r.rank, strconv.FormatInt(int64(uid), 10)).Result()
	if err != nil {
		return 0
	}
	return int64(-score)
}

func (r *RankList) GetScore(uid int32) (int64, error) {
	score, err := r.client.ZScore(r.rank, strconv.FormatInt(int64(uid), 10)).Result()
	if err != nil {
		return 0, err
	}
	return int64(-score), nil
}

func (r *RankList) Range(start, stop int32) (users []int32) {
	if start > 0 {
		start = start - 1
	}
	if stop > 0 {
		stop = stop - 1
	}
	s, err := r.client.ZRange(r.rank, int64(start), int64(stop)).Result()
	if err != nil {
		return nil
	}
	for _, x := range s {
		uid, _ := strconv.ParseInt(x, 10, 32)
		users = append(users, int32(uid))
	}
	return
}

type RankItem struct {
	Uid   int32
	Score int64
}

func (r *RankList) RangeWithScore(start, stop int32) (ranks []RankItem) {
	if start > 0 {
		start = start - 1
	}
	if stop > 0 {
		stop = stop - 1
	}
	zs, err := r.client.ZRangeWithScores(r.rank, int64(start), int64(stop)).Result()
	if err != nil {
		return nil
	}
	for _, x := range zs {
		uid, _ := strconv.ParseInt(x.Member.(string), 10, 32)
		ranks = append(ranks, RankItem{int32(uid), int64(-x.Score)})
	}
	return
}

func (r *RankList) GetUidByRank(rank int32) int32 {
	zs, err := r.client.ZRange(r.rank, int64(rank), int64(rank)).Result()
	if err != nil {
		return -1
	}
	if len(zs) == 0 {
		return -1
	}
	uid, _ := strconv.ParseInt(zs[0], 10, 32)
	return int32(uid)
}

func (r *RankList) Delete(uid int32) {
	r.client.ZRem(r.rank, uid)
}

func (r *RankList) Count() int32 {
	n, err := r.client.ZCard(r.rank).Result()
	if err != nil {
		return -1
	}
	return int32(n)
}

func (r *RankList) Clear() {
	r.client.Del(r.rank)
}

func (r *RankList) Close() {
	//r.client.Close()
}
