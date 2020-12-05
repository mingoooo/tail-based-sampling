package agent

import (
	"fmt"
	"log"

	"github.com/valyala/fasthttp"
)

func (r *Receiver) ReadyHTTPHandler(ctx *fasthttp.RequestCtx) {
	log.Printf("Ready")
	go func() {
		// defer os.Exit(0)
		// defer trace.Stop()

		r.finishWg.Add(1)
		r.finishWg.Wait()
		for {
			if err := r.Postman.ConfirmFinish(r.traceCh, r.errTidSubCh, r.errTidPubCh); err != nil {
				log.Println(err)
				continue
			}
			log.Printf("Done")
			return
		}
	}()
	go r.TraceMarker()
	go func() {
		for {
			if err := r.Postman.TracePublisher(r.traceCh); err != nil {
				log.Println(err)
				continue
			}
			return
		}
	}()

	go func() {
		for {
			if err := r.Postman.TraceIDSubscriber(r.errTidSubCh); err != nil {
				log.Println(err)
				continue
			}
			return
		}
	}()

	go func() {
		for {
			if err := r.Postman.ErrTraceIdPublisher(r.errTidPubCh); err != nil {
				log.Println(err)
				continue
			}
			return
		}
	}()

	ctx.SetStatusCode(fasthttp.StatusOK)
}

func (r *Receiver) SetParamHandler(ctx *fasthttp.RequestCtx) {
	log.Printf("Set param")
	r.DataPort = string(ctx.QueryArgs().Peek("port"))
	// TODO: TEST
	r.DataPort = "8082"
	r.DataURL = fmt.Sprintf("http://127.0.0.1:%s/trace%s.data", r.DataPort, r.DataSuffix)

	rs, err := r.GetDataReaders()
	if err != nil {
		log.Fatalln(err)
	}

	go func() {
		if err := r.Filter(rs); err != nil {
			log.Fatalln("Filter: ", err)
		}
	}()

	go r.StartDownload(rs)
	// time.Sleep(5 * time.Second)

	log.Printf("Reply SetParam")
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

	fasthttp.ListenAndServe(fmt.Sprintf(":%s", r.HTTPPort), m)
}
