package main

import (
	m "atix.de/voting/worker-go/model"
	"atix.de/voting/worker-go/postgres"
	"atix.de/voting/worker-go/probe"
	"atix.de/voting/worker-go/redis"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var status = probe.Status{}

func main() {
	log.Info("starting")
	c := viper.New()
	c.SetDefault("verbose", false)
	c.SetDefault("redis_key", "vote")
	c.SetDefault("redis_host", "redis:6379")
	c.SetDefault("redis_db", 0)
	c.SetDefault("redis_pass", "")
	c.SetDefault("postgres_host", "postgres:5432")
	c.SetDefault("postgres_user", "postgres")
	c.SetDefault("postgres_password", "")
	c.SetDefault("postgres_db", "postgres")
	c.AutomaticEnv()

	redisHost := c.GetString("redis_host")
	redisPass := c.GetString("redis_pass")
	redisDb := c.GetInt("redis_db")
	redisKey := c.GetString("redis_key")
	log.WithFields(log.Fields{
		"host": redisHost,
		"pass": "***",
		"db":   redisDb,
		"key":  redisKey,
	}).Info("got redis config")

	postgresUser := c.GetString("postgres_user")
	postgresPass := c.GetString("postgres_password")
	postgresHost := c.GetString("postgres_host")
	postgresDb := c.GetString("postgres_db")
	postgresUri := "postgres://" + postgresUser + ":" + postgresPass +
		"@" + postgresHost + "/" + postgresDb
	log.WithFields(log.Fields{
		"host": postgresHost,
		"user": postgresUser,
		"pass": "***",
		"db":   postgresDb,
	}).Info("got postgres config")

	// set log-level
	if c.GetBool("verbose") {
		log.Info("Set Log Level to debug")
		log.SetLevel(log.DebugLevel)
	}

	go probe.StartProbeServer(&status)

	log.Debug("initializing worker")
	rWorker := redis.NewWorker(redisHost, redisPass, redisDb, redisKey)
	pWorker := postgres.NewWorker(postgresUri)

	voteChan := make(chan m.Vote)

	log.Debug("starting worker")
	go func() {
		err := rWorker.Read(voteChan)
		log.Error(err)
		status.Living = false
	}()
	go func() {
		pWorker.Write(voteChan)
		log.Error("communication channel closed")
		status.Living = false
	}()

	status.Started = true
	status.Living = true

	log.Info("started")

	select {}
}
