package probe

import (
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
)

var (
	addr = ":8080"
)

type Status struct {
	Started bool
	Living  bool
}

// start a small webserver to get startup and liveness probes
func StartProbeServer(status *Status) {
	h := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/livez":
			log.Debug("handle liveness probe")
			if !status.Living {
				ctx.Response.SetStatusCode(503)
			}
		case "/startz":
			log.Debug("handle startup probe")
			if !status.Started {
				ctx.Response.SetStatusCode(503)
			}
		default:
			log.Warn("unknown probe request")
			ctx.Response.SetStatusCode(501)
		}
	}

	if err := fasthttp.ListenAndServe(addr, h); err != nil {
		log.Error("Error in ListenAndServe: %s", err)
	}
	log.WithField("uri", addr).Debug("started probe server")
}
