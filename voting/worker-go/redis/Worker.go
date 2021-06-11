package redis

import (
	"context"
	"encoding/json"
	"time"

	m "atix.de/voting/worker-go/model"
	redis "github.com/go-redis/redis/v8"
	log "github.com/sirupsen/logrus"
)

type worker struct {
	rdb *redis.Client
	key string
}

type Worker interface {
	Read(chan<- m.Vote) error
}

var ctx = context.Background()

func NewWorker(redisHost, redisPw string, dbNum int, redisKey string) Worker {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPw, // no password set
		DB:       dbNum,   // use default DB
	})

	return worker{
		rdb: rdb,
		key: redisKey,
	}
}

func (w worker) Read(channel chan<- m.Vote) error {
	for {
		// BLPop returns an array of [<key>,<value>]
		jsonVotes, err := w.rdb.BLPop(ctx, 5*time.Minute, w.key).Result()
		log.WithField("val", jsonVotes).
			Debug("got votes from redis")
		if err != nil {
			return err
		}
		if len(jsonVotes) < 2 {
			continue
		}

		// get last element in array
		jVote := jsonVotes[len(jsonVotes)-1]

		var vote m.Vote
		log.WithField("val", jVote).Debug("unmarshalling vote")
		e0 := json.Unmarshal([]byte(jVote), &vote)
		if e0 != nil {
			log.Warn(e0)
			continue
		}
		log.WithField("vote", vote).
			Info("processing vote")
		channel <- vote
	}
	return nil
}
