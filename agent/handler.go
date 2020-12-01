package agent

import (
	"fmt"
	"log"

	"github.com/valyala/fasthttp"
)

func (r *Receiver) ReadyHTTPHandler(ctx *fasthttp.RequestCtx) {
	go r.TraceSearcher()
	go func() {
		for {
			if err := r.Postman.TracePublisher(r.traceCh); err != nil {
				log.Println(err)
			}
		}
	}()

	go func() {
		for {
			if err := r.Postman.TraceIDSubscriber(r.tidCh); err != nil {
				log.Println(err)
			}
		}
	}()

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (r *Receiver) SetParamHandler(ctx *fasthttp.RequestCtx) {
	r.DataPort = string(ctx.QueryArgs().Peek("port"))
	// TODO: TEST
	// r.DataPort = "8081"
	r.DataURL = fmt.Sprintf("http://127.0.0.1:%s/trace%s.data", r.DataPort, r.DataSuffix)

	go r.PullData([]byte{})

	ctx.SetStatusCode(fasthttp.StatusOK)
}

// RunHTTPSvr Run HTTP server
func (r *Receiver) RunHTTPSvr() {
	m := func(ctx *fasthttp.RequestCtx) {
		switch string(ctx.Path()) {
		case "/ready":
			r.ReadyHTTPHandler(ctx)
		case "/setParameter":
			r.SetParamHandler(ctx)
		default:
			ctx.Error("Unsupported path", fasthttp.StatusNotFound)
		}
	}

	fasthttp.ListenAndServe(fmt.Sprintf(":%s", r.HttpPort), m)
}
