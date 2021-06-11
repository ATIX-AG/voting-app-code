package main

import (
	"math/rand"
	"strings"
	"time"

	"git.atix.de/voting/load/id"
	"git.atix.de/voting/load/probe"
	"git.atix.de/voting/load/vote"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/valyala/fasthttp"
)

var status = probe.Status{}

func main() {
	// get config
	c := viper.New()
	c.SetDefault("url", "vote")
	c.SetDefault("parallel_requests", 10)
	c.SetDefault("verbose", false)
	c.SetDefault("choices", "a,b")
	c.SetDefault("preference", "b")
	c.SetDefault("sleep", 0)
	c.SetDefault("redis_host", "redis:6379")
	c.SetDefault("redis_pw", "")
	c.SetDefault("redis_db_num", 0)
	c.AutomaticEnv()

	amount := c.GetInt("parallel_requests")
	sleep := c.GetInt("sleep")
	url := c.GetString("url")
	choices := strings.Split(c.GetString("choices"), ",")
	preference := c.GetString("preference")
	redisHost := c.GetString("redis_host")
	redisPw := c.GetString("redis_pw")
	redisDbNum := c.GetInt("redis_db_num")
	chooser, err := vote.NewChooser(choices, preference)
	log.WithFields(log.Fields{
		"choices":    choices,
		"sleep":      sleep,
		"url":        url,
		"amount":     amount,
		"preference": preference,
		"host":       redisHost,
		"pass":       "***",
		"db":         redisDbNum,
	}).Info("got config")
	go probe.StartProbeServer(&status)

	if err != nil {
		panic(err)
	}

	log.WithField("url", url).
		Info("fetching from url")

	// set log-level
	if c.GetBool("verbose") {
		log.Info("Set Log Level to debug")
		log.SetLevel(log.DebugLevel)
	}

	// seed random
	rand.Seed(time.Now().UTC().UnixNano())

	// create client
	client := &fasthttp.Client{
		NoDefaultUserAgentHeader:      true,
		MaxConnsPerHost:               amount,
		ReadBufferSize:                amount * 1024, // Make sure to set this big enough that your whole request can be read at once.
		WriteBufferSize:               amount * 1024, // Same but for your response.
		DisableHeaderNamesNormalizing: false,
	}

	idHelper := id.NewDB(redisHost, redisPw, redisDbNum)

	// listen for requests
	log.Info("creating requests")
	for i := 0; i < int(amount); i++ {
		go loopRequest(client, url, idHelper, *chooser, sleep)
	}
	log.Info("Initialization finished, running requests")

	status.Started = true
	status.Living = true
	select {}
}

func loopRequest(client *fasthttp.Client, url string, idHelper id.IdHelper, voter vote.Voter, sleep int) {
	for {
		cookie := idHelper.ClaimId()
		log.WithField("cookie", string(cookie)).Debug("claimed Cookie")
		vote := voter.Pick()
		cookie = doVoteRequest(client, url, cookie, vote)
		idHelper.DisownId(cookie)
		log.WithField("cookie", string(cookie)).Debug("Disowned Cookie")
		if sleep > 0 {
			time.Sleep(time.Duration(sleep) * time.Second)
		}
	}
}

func doVoteRequest(
	client *fasthttp.Client,
	url string,
	cookie []byte,
	vote string,
) []byte {
	defer trackTime(time.Now(), "vote")
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	log.WithField("url", url).
		Debug("Doing vote Request")
	req.SetRequestURI(url)
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.SetMethod("POST")
	req.SetBodyString("vote=" + vote)

	req.Header.SetCookie("", string(cookie))

	err := fasthttp.Do(req, resp)
	if err != nil {
		status.Living = false
		println(err.Error())
	}

	statusCode := resp.StatusCode()
	if statusCode >= 400 {
		log.WithField("status", statusCode).Warn("bad response status")
		status.Living = false
	}
	c := resp.Header.PeekCookie("voter_id")
	log.WithField("cookie", string(c)).
		Debug("got cookie")

	return c
}

func trackTime(start time.Time, name string) {
	elapsed := time.Since(start)
	log.WithField("took", elapsed).
		Debug("" + name + " finished")
}
