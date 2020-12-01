package agent

import (
	"fmt"
	"log"
	"os"

	"github.com/valyala/fasthttp"
)

func (r *Receiver) ReadyHTTPHandler(ctx *fasthttp.RequestCtx) {
	go func() {
		defer os.Exit(0)
		r.finishWg.Add(1)
		r.finishWg.Wait()
		for {
			if err := r.Postman.ConfirmFinish(r.traceCh, r.tidCh); err != nil {
				log.Println(err)
				continue
			}
			log.Printf("Done")
			return
		}
	}()
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
	r.DataPort = "8082"
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
