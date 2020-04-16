package http

import (
	"github.com/valyala/fasthttp"
	"github.com/graphicweave/ox/database/model"
	"encoding/json"
	"github.com/rs/xid"
	"github.com/sirupsen/logrus"
	"net/http"
	"fmt"
	"github.com/graphicweave/ox/payment_gatway"
	"github.com/tidwall/gjson"
	"github.com/graphicweave/ox/grpc/mail"
	//"bytes"
	"time"
	//"strconv"
	"html/template"
	"sync"
	"reflect"
	//"github.com/graphicweave/ox/exchange_rate"
	mail2 "github.com/graphicweave/ox/mail"
)

var (
	once sync.Once
	html *template.Template
)

const shortForm = "2006-01-02"

func PurchaseTicketsHandler(ctx *fasthttp.RequestCtx) {

	body := ctx.Request.Body()

	var t model.TicketDetail
	var payment model.Payment

	err := json.Unmarshal(body, &t)
	guid := xid.New()
	t.ID = guid.String()

	if err != nil {
		logrus.Error("Could not parse ticket : ", err)
		ctx.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &payment)

	if err != nil {
		logrus.Error("Could not parse payment details : ", err)
		ctx.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	if t.NoOfAttendees == 0 {

		_, err := t.PurchaseTicketsDB()

		if err != nil {
			logrus.Error("Couldn't save purchase details : ", err)
			ctx.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		logrus.Info("Visitors added")

		json.NewEncoder(ctx).Encode(map[string]interface{}{"ticketId": guid, "eventId": t.EventID})

	} else {
		eventInfo, err := model.CheckAvailability(t.EventID, float64(t.NoOfAttendees))

		if err != nil {
			logrus.Error("No tickets Available", err)
			errorMsg := fmt.Sprint("%d tickets not available : ", t.NoOfAttendees)
			ctx.Error(errorMsg, http.StatusInternalServerError)
			return
		}

		key, err := t.PurchaseTicketsDB()

		if err != nil {
			logrus.Error("Couldn't save purchase details : ", err)
			ctx.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		stripeToken, err := payment_gatway.GetToken(payment.CardDetail)

		if err != nil {
			model.RollBack(key, eventInfo)

			logrus.Error("stripe Token error : ", err)
			ctx.Error("Invalid card details : ", http.StatusInternalServerError)
			return
		}

		chargeDetails, err := payment_gatway.ChargeAmount(payment.CardDetail.Currency, stripeToken.ID, uint64(payment.AmountCharged))

		if err != nil {
			model.RollBack(key, eventInfo)

			logrus.Error("payment Gatway : could not compelete transaction", err)
			ctx.Error("Could not make payment", http.StatusInternalServerError)
			return
		}

		//exhangeRate, err := exchange_rate.GetExchangeRate(payment.CardDetail.Currency)
		//
		//if err != nil {
		//	logrus.Error("exchange Rate : could not get the exchange rate", err)
		//	ctx.Error("could not get exchange rate", http.StatusInternalServerError)
		//	return
		//}

		err = model.PurchaseDetails(chargeDetails, key)
		if err != nil {
			logrus.Error("Couldn't update payment Status : ", err)
			ctx.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		logrus.Info("Payment done")

		go func() {

			mail.ConfirmationEmail(mail2.ConfirmationRequestDetail{
				Name: t.Name,
				TicketNumber: t.ID,
				TicketAmount: t.NoOfAttendees,
				TicketPurchaseDate: time.Unix(t.PurchasedOn/1000, 0).Format(time.RFC822),
				EmailId: t.Email,
			})
		}()

		go func() {
			if payment.SaveCardDetail {
				t.BillingDetails.UserId = t.UserID
				err := t.BillingDetails.SaveBillingDetails()
				if err != nil {
					logrus.Error("Could not save billing details", err)
				} else {
					logrus.Infoln("billing details saved")
				}
			}
		}()

		logrus.Info("Ticket(s) purchased")

		json.NewEncoder(ctx).Encode(map[string]interface{}{"ticketId": guid, "eventId": t.EventID})
	}
}

func UpdateAttendeeHandler(ctx *fasthttp.RequestCtx) {

	body := string(ctx.Request.Body())
	var personalDetail model.PersonalDetail
	var eventDetail map[string]interface{}
	guid := xid.New()

	data := gjson.Get(body, "data")
	ticketId := gjson.Get(body, "ticketId")
	isAttendee := gjson.Get(body, "isAttendee")
	eventDetail_ := gjson.Get(body, "eventDetail")
	organizer := gjson.Get(body, "organizer")
	err := json.Unmarshal([]byte(data.String()), &personalDetail)

	if err != nil {
		logrus.Error("error in  personal Detail", err)
		ctx.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	e := json.Unmarshal([]byte(eventDetail_.String()), &eventDetail)
	if e != nil {
		logrus.Error("Error in eventDetail", e)
	}
	eventDetail["emailId"] = personalDetail.Email
	eventDetail["username"] = personalDetail.Name
	eventDetail["organizer"] = organizer.Raw
	eventDetail["userId"] = guid.String()
	eventDetail["ticketId"] = ticketId.String()

	if isAttendee.Bool() {
		go func() {
			mail.AttendeeConfirmation(eventDetail)
		}()
	} else {
		go func() {
			mail.VisitorEmail(eventDetail)
		}()
	}

	personalDetail.Id = guid.String()
	personalDetail.Confirmation = "PENDING"

	err = model.UpdateAttendee(personalDetail, ticketId.String(), isAttendee.Bool())

	if err != nil {
		logrus.Error("error in update", err)
		ctx.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	logrus.Info("Attendee / Visitors Added")
	json.NewEncoder(ctx).Encode("ok")
}

func GetTicketHandler(ctx *fasthttp.RequestCtx) {

	id := ctx.QueryArgs().Peek("id")

	tickets, err := model.GetTicket(string(id))
	if err != nil {
		logrus.Error("error in db", err)
		return
	}

	if err := json.NewEncoder(ctx).Encode(tickets); err != nil {
		logrus.Error("couldn't encode response", err)
		ctx.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

}

func GetBillingDetailHandler(ctx *fasthttp.RequestCtx){
	id := ctx.QueryArgs().Peek("id")
    fmt.Println("id",id)
	detail, err :=  model.GetBillingDetail(string(id))
	fmt.Println(detail)
	if err != nil {
		logrus.Error("error in db", err)
		ctx.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(ctx).Encode(detail); err != nil {
		logrus.Error("couldn't encode response", err)
		ctx.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

type data struct {
	Day     string
	Month   string
	Name    string
	Address string
}

func PayoutsHandler(ctx *fasthttp.RequestCtx) {

	userId := ctx.QueryArgs().Peek("id")
	//etcr := ctx.QueryArgs().Peek("etcr")

	events, err := model.GetEventsByUserId(string(userId))

	if err != nil {
		logrus.Error("couldn't fetch events of userId : %s : ", userId, err)

		return
	}

	//const t = `<div class='item__table'><div class='item__date'><div class='item__day'>{{.Day}}</div><div class='item__month'>{{.Month}}</div></div><div class='item__description'><div class='item__name'>{{.Name}}</div><div class='item__address'>{{.Address}}</div></div></div>`
	//
	//once.Do(func() {
	//	html, _ = template.New("html_template").Parse(t)
	//})
	//
	//for k, v := range events {
	//	s := ""
	//	w := bytes.NewBufferString(s)
	//	t, _ := time.Parse(shortForm, v["date"].(string))
	//
	//	d := data{
	//		Name:    v["name"].(string),
	//		Month:   t.Format("Jan"),
	//		Day:     strconv.Itoa(t.Day()),
	//		Address: v["venueName"].(string) + ", " + v["venueCity"].(string),
	//	}
	//
	//	err := html.Execute(w, d)
	//
	//	if err != nil {
	//		logrus.Error("error in html template", err)
	//	}
	//	price, _ := v["price"].(float64)
	//	sold, _ := v["sold"].(float64)
	//	if v["sold"] == nil {
	//		v["sold"] = 0
	//	} else {
	//		v["sold"] = strconv.FormatFloat(v["sold"].(float64),'f', 0, 64)
	//	}
	//
	//	commissionRate, _ := strconv.ParseFloat(string(etcr), 32)
	//
	//	revenue := price * sold
	//	events[k]["name"] = v["name"].(string)
	//	if events[k]["includeTax"] == "true" {
	//		taxRate, _ := strconv.ParseFloat(v["taxRate"].(string), 32)
	//		events[k]["revenue"] = revenue + (revenue*taxRate)/100
	//
	//	} else {
	//		events[k]["revenue"] = revenue
	//	}
	//	events[k]["payout"] = events[k]["revenue"].(float64) - (events[k]["revenue"].(float64)*commissionRate)/100
	//
	//}
	logrus.Info("payouts streamed")
	json.NewEncoder(ctx).Encode(map[string]interface{}{"data": events})

}

func GetAttendeesHandler(ctx *fasthttp.RequestCtx) {

	ticketID := ctx.QueryArgs().Peek("id")
	details, err := model.GetAttendees(string(ticketID), true)

	if err != nil {
		logrus.Error("couldn't fetch Attendees : %s : ", string(ticketID), err)
		ctx.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(ctx).Encode(map[string]interface{}{"data": details})
}

func GetVisitorsHandler(ctx *fasthttp.RequestCtx) {

	ticketID := ctx.QueryArgs().Peek("id")
	details, err := model.GetAttendees(string(ticketID), false)

	if err != nil {
		logrus.Error("couldn't fetch Visitors : %s : ", string(ticketID), err)
		ctx.Error(http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(ctx).Encode(map[string]interface{}{"data": details})
}

func GetTicketDetailsHandler(ctx *fasthttp.RequestCtx) {

	ticketID := fmt.Sprintf("%s", ctx.UserValue("ticketId"))

	logrus.Info("got ticket details request for ticket id: ", ticketID)

	response := AnyResponse{}

	if ticketID == "" {
		response.Status = "ERR"
		response.Message = "'ticketId' missing"
	}

	if !reflect.DeepEqual(response, AnyResponse{}) {
		if err := json.NewEncoder(ctx).Encode(&response); err != nil {
			httpError(ctx, err)
		}
		return
	}

	var detail TicketDetail

	if err := json.NewEncoder(ctx).Encode(detail.GetTicketDetails(ticketID)); err != nil {
		httpError(ctx, err)
		return
	}
}
func GetAttendeeInfo(ctx *fasthttp.RequestCtx) {

	ticketID := fmt.Sprintf("%s", ctx.UserValue("ticketId"))

	logrus.Info("got attendee request for ticket id: ", ticketID)

	response := AnyResponse{}

	if ticketID == "" {
		response.Status = "ERR"
		response.Message = "'ticketId' missing"
	}

	if !reflect.DeepEqual(response, AnyResponse{}) {
		if err := json.NewEncoder(ctx).Encode(&response); err != nil {
			httpError(ctx, err)
		}
		return
	}

	resp, err := model.GetAttendeeInfo(ticketID)
	if err != nil {
		logrus.Error("get attendee info", err)
		httpError(ctx, err)
		return
	}

	if err := json.NewEncoder(ctx).Encode(resp); err != nil {
		httpError(ctx, err)
		return
	}
}
