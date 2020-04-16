package http

/**
 * Created by  ♅ Salfi Farooq on 23/06/17.
 */

import (
	"github.com/valyala/fasthttp"
	"github.com/buaazp/fasthttprouter"
	"net/http"
	log "github.com/sirupsen/logrus"
)

func StartServer(addr string) error {

	fastRouter := fasthttprouter.New()

	fastRouter.MethodNotAllowed = func(ctx *fasthttp.RequestCtx) {
		ctx.Error(http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}

	fastRouter.NotFound = func(ctx *fasthttp.RequestCtx) {
		ctx.Error(http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	fastRouter.PanicHandler = func(ctx *fasthttp.RequestCtx, i interface{}) {
		log.Errorln("something broke")
		log.Errorln("error: ", i)
		ctx.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	fastRouter.POST("/purchase-tickets", cors(PurchaseTicketsHandler))
	fastRouter.POST("/update-attendee", cors(UpdateAttendeeHandler))
    fastRouter.GET("/get-billing-detail",cors(GetBillingDetailHandler))
	fastRouter.GET("/get-ticket", cors(GetTicketHandler))
	fastRouter.GET("/payouts", cors(PayoutsHandler))
	fastRouter.GET("/get-attendees", cors(GetAttendeesHandler))
	fastRouter.GET("/get-visitors", cors(GetVisitorsHandler))
	fastRouter.GET("/ticket-details/:ticketId", cors(GetTicketDetailsHandler))
	fastRouter.GET("/get-attendee-info/:ticketId", cors(GetAttendeeInfo))
	fastRouter.GET("/ping", cors(func(ctx *fasthttp.RequestCtx) {
		ctx.WriteString("PONG\n")
	}))

	return newServer(fastRouter).ListenAndServe(addr)
}

func newServer(fastRouter *fasthttprouter.Router) *fasthttp.Server {

	return &fasthttp.Server{
		Name:              "ox",
		Handler:           fastRouter.Handler,
		ReduceMemoryUsage: true,
		LogAllErrors:      true,
	}
}

func cors(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {

		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")
		ctx.Response.Header.Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		ctx.Response.Header.Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		ctx.Response.Header.Set("Content-Type", "application/json; charset=UTF-8")

		next(ctx)
	}
}

func httpError(ctx *fasthttp.RequestCtx, err error) {
	log.Error("internal server error: " + err.Error())
	ctx.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

