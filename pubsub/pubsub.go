package pubsub

import (
	"github.com/go-redis/redis"
	"log"
)

type PubSub struct {
	addr string
	db int
	channel string
}

func NewPubSub(addr string, db int, ch string) *PubSub {
	return &PubSub{addr:addr, db:db, channel:ch}
}

func (ps *PubSub)Subscribe(f func(string)) {
	client := redis.NewClient(&redis.Options{
		Addr:     ps.addr,
		Password: "",
		DB:       ps.db,
	})
	sub := client.Subscribe(ps.channel)
	go func() {
		for {
			msg, err := sub.ReceiveMessage()
			if err != nil {
				log.Println("redis subscribe error", err)
				continue
			}
			if msg.Channel == ps.channel {
				f(msg.Payload)
			}
		}
	}()
}

func (ps *PubSub)Publish(s string) {
	client := redis.NewClient(&redis.Options{
		Addr:     ps.addr,
		Password: "",
		DB:       ps.db,
	})
	client.Publish(ps.channel, s)
	client.Close()
}
