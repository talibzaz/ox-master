package model

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/arangodb/go-driver"
	"fmt"
	"github.com/stripe/stripe-go"
	"errors"
	"github.com/graphicweave/ox/ox_idl"
	currency "github.com/younisshah/go-currency-code"
	"time"
)

/**
 * Created by  ♅ Salfi Farooq on 23/06/18.
 */

type TicketDetail struct {
	ID             string
	EventID        string
	UserID         string
	Name           string
	Email          string
	NoOfVisitors   int32
	NoOfAttendees  int32
	TaxRate        float64
	TaxAmount      float64
	TicketPrice    float64
	AmountCharged  float64
	BillingDetails Billing
	Payment        bool
	Visitors       []interface{}
	Attendees      []interface{}
	PurchasedOn    int64
	TimeZone       string
	Zone           string
	ExchangeRate   float64
}

type Billing struct {
	Country  string
	Address1 string
	Address2 string
	City     string
	State    string
	PostCode string
	UserId   string
	CreatedOn int64
}

type PersonalDetail struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Confirmation string `json:"confirmation"`
}

type EventInfo struct {
	eventID string
	sold    int32
}

type ResponseChan struct {
	Done chan struct{}
	Err  chan error
}

type AttendeeInfo struct {
	Attendees    int
	RegAttendees int
	Visitors     int
	RegVisitors  int
}

func (t *TicketDetail) PurchaseTicketsDB() (string, error) {

	ctx := context.Background()

	db, err := GetDatabase(ctx)

	if err != nil {
		logrus.Error("purchase Ticket : Error in database ", err)
		return "", err
	}

	collection, err := db.Collection(ctx, viper.GetString("TICKETS_COLLECTION"))
	if err != nil {
		logrus.Error("purchase Ticket : Error in collection ", err)
		return "", err
	}

	t.Payment = false

	meta, err := collection.CreateDocument(ctx, t)
	if err != nil {
		logrus.Error("purchase Ticket : Couldn't create document", err)
		return "", err
	}

	return meta.Key, nil
}

func CheckAvailability(eventID string, ticketQuantity float64) (EventInfo, error) {

	ctx := context.Background()

	db, err := GetDatabase(ctx)
	if err != nil {
		logrus.Error("Purchase Ticket : Error in database ", err)
		return EventInfo{}, err
	}
	collection := viper.GetString("EVENTS_COLLECTION")
	query := fmt.Sprintf("FOR e IN %s FILTER e.eventDetail.id == '%s' RETURN { sold: e.ticket.sold, quantity: e.ticket.quantity ,key: e._key}", collection, eventID)
	cursor, err := db.Query(ctx, query, nil)

	if err != nil {
		logrus.Error("error in cursor ", err)
		return EventInfo{}, err
	}

	var data map[string]interface{}

	_, err = cursor.ReadDocument(ctx, &data)
	defer cursor.Close()

	if err != nil {
		logrus.Error("get ticket : error in database ", err)
		return EventInfo{}, err
	}

	fmt.Println("asas", data["sold"])
	sold, ok := data["sold"].(float64)
	if !ok {
		sold = 0
	}

	quantity, _ := data["quantity"].(float64)

	if ticketQuantity <= quantity-sold {
		query := fmt.Sprintf("For e in %s filter e._key == '%s' let t = e.ticket update e with {ticket : MERGE(t,{'sold': %d})} in %s", collection, data["key"], int64(sold+ticketQuantity), collection)
		_, err := db.Query(ctx, query, nil)

		if err != nil {
			return EventInfo{}, err
		}

		return EventInfo{eventID: data["key"].(string), sold: int32(sold)}, nil
	} else {
		return EventInfo{}, errors.New("tickets not available")
	}

}

func UpdateAttendee(data PersonalDetail, ticketID string, isAttendee bool) (error) {

	ctx := context.Background()

	db, err := GetDatabase(ctx)

	if err != nil {
		logrus.Error("Update Ticket : Error in database ", err)
		return err
	}

	bindVars := make(map[string]interface{})
	bindVars["ticketId"] = ticketID
	bindVars["data"] = data

	var query string

	if isAttendee {
		query = `for t in tickets 
					filter t.ID == @ticketId
  						 UPDATE t with {'Attendees': PUSH(t.Attendees,@data)} in tickets`
	} else {
		query = `for t in tickets 
  					 filter t.ID == @ticketId
  						 UPDATE t with {'Visitors': PUSH(t.Visitors,@data)} in tickets`
	}

	if err != nil {
		logrus.Error("Update Ticket : Error in collection ", err)
		return err
	}
	_, err = db.Query(ctx, query, bindVars)
	return err
}

func ConfirmAttendeeEmail(userID, ticketId string) (error) {
	ctx := context.Background()

	db, err := GetDatabase(ctx)

	if err != nil {
		logrus.Error("Update Attendee Confirmation : Error in database ", err)
		return err
	}

	query := fmt.Sprintf(`FOR e in tickets 
                                   filter e.ID == '%s'
                                    LET alt = (  FOR x in e.Attendees
    												 LET n = ( x.id == '%s' ?
															   MERGE( x , {confirmation:"APPROVED"}) :
																x )
     											  return n )
                                 update e with {Attendees:alt} in tickets
						`, ticketId, userID)

	_, err = db.Query(ctx, query, nil)

	logrus.Info("Email approved")
	return err

}

func GetTicketsByUserID(userID string, c chan ox_idl.Ticket, respChan ResponseChan) {

	ctx := context.Background()
	db, err := GetDatabase(ctx)

	if err != nil {
		logrus.Error("get ticket : error in database ", err)
		respChan.Err <- err
		return
	}

	query := fmt.Sprintf(`FOR t IN tickets
                                    FOR e IN events
										FILTER t.EventID == e.eventDetail.id && t.UserID == '%s'
   										 RETURN { amount: t.AmountCharged,
           										  ticketID:t.ID,
                                                  eventID: t.EventID,
												  eventName: e.eventDetail.name,
                                                  startDate: e.eventDetail.start_date,	 
      							                  startTime:e.eventDetail.start_time,
                                                  currency:e.ticket.currency,
                                                  endDate: e.eventDetail.end_date,	 
      							                  endTime:e.eventDetail.end_time,
                                                  zone : e.eventDetail.zone,
        										  timeZone : e.eventDetail.timezone,
                                                  ticketZone: t.Zone,
                                                  ticketTimeZone : t.TimeZone,
     						                      venueCity:e.eventDetail.venue_city,
       						                      venueName: e.eventDetail.venue_name,
                                                  company: e.organizer.name,
                                                  image: e.eventDetail.cover_image_thumbnail_upload_id,
                                                  logo:  e.organizer.upload_id,
                                                  purchasedOn: t.PurchasedOn
                                                }`, userID)

	cursor, err := db.Query(ctx, query, nil)
	if err != nil {
		logrus.Error("get ticket : error in get query ", err)
		respChan.Err <- err
		return
	}
	defer cursor.Close()

	var ticket ox_idl.Ticket
	for {
		_, err := cursor.ReadDocument(ctx, &ticket)
		if driver.IsNoMoreDocuments(err) {
			logrus.Info("tickets streamed")
			close(respChan.Done)
			break
		} else if err != nil {
			logrus.Error("get ticket : cursor", err)
			respChan.Err <- err
		}
		c <- ticket
	}

}

func GetEventsByUserId(userId string) ([]map[string]interface{}, error) {
	ctx := context.Background()

	db, err := GetDatabase(ctx)

	if err != nil {
		logrus.Error("get Event : error in database ", err)
		return nil, err
	}
	//TODO check event zone to get only past events
	query := fmt.Sprintf(`FOR e IN events 
		FILTER e.eventDetail.user_id == '%s' && e.eventDetail.end_date < DATE_ISO8601(DATE_ADD(DATE_NOW(), -1, "day")) && e.status == 'PUBLISHED'
		LET etShare = (
		    FOR t IN tickets
		    FILTER e.eventDetail.id == t.EventID && t.Payment == true
          	RETURN t.ExchangeRate == 0 ? 0 : (t.AmountCharged - t.TaxAmount) / t.ExchangeRate * e.eventDetail.et_commission_rate * 0.01 
		)
		
		LET revenue = (
        RETURN SUM (
    		FOR t IN tickets 
    		FILTER t.EventID == e.eventDetail.id && t.Payment == true
    		RETURN t.ExchangeRate == 0 ? 0 : t.AmountCharged/t.ExchangeRate
		)   
        )[0]
        
		RETURN {
   			id: e.eventDetail.id,
   			name: e.eventDetail.name,
			sold: (
			    LET sold = (
                RETURN SUM(
                    FOR t IN tickets
                    FILTER e.eventDetail.id == t.EventID && t.Payment == true
                    RETURN t.NoOfAttendees
                    )
                )
                return sold
            )[0][0],
			revenue: revenue,
   			payout: revenue  - sum(etShare),
   			status : e.payouts_status  == "PAID" ? "PAID" : "UNPAID"  
		}`, userId)

	cursor, err := db.Query(ctx, query, nil)
	defer cursor.Close()
	var events []map[string]interface{}
	for {
		var event map[string]interface{}
		_, err := cursor.ReadDocument(ctx, &event)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			logrus.Error("get events : cursor", err)
		}
		events = append(events, event)
	}
	return events, nil
}

func GetTicket(id string) (TicketDetail, error) {

	ctx := context.Background()

	db, err := GetDatabase(ctx)

	if err != nil {
		logrus.Error("get ticket : error in database ", err)
		return TicketDetail{}, err
	}

	query := fmt.Sprintf("FOR d IN %s  FILTER d._key == '%s'  RETURN d  ", viper.GetString("TICKETS_COLLECTION"), id)

	cursor, err := db.Query(ctx, query, nil)
	defer cursor.Close()

	var ticket TicketDetail

	_, err = cursor.ReadDocument(ctx, &ticket)
	if err != nil {
		logrus.Error("get ticket : error in database ", err)
		return TicketDetail{}, err
	}
	return ticket, nil

}

func GetBillingDetail(userId string) (map[string]interface{}, error) {
	ctx := context.Background()

	db, err := GetDatabase(ctx)

	if err != nil {
		logrus.Error("get ticket : error in database ", err)
		return nil, err
	}

	query := fmt.Sprintf("let  y  = (FOR d IN %s  FILTER d.UserId == '%s' SORT d.CreatedOn ASC RETURN d) RETURN LAST(y)  ", viper.GetString("CARD_COLLECTION"), userId)


	cursor, err := db.Query(ctx, query, nil)


	if err != nil || cursor == nil {
		logrus.Error("get ticket : error in database ", err)
		return nil,err
	}
	defer cursor.Close()


	detail := map[string]interface{}{}
	_, err = cursor.ReadDocument(ctx, &detail)
	if err != nil {
		logrus.Error("get ticket : error in database ", err)
		return nil, err
	}
	return detail, nil

}

func RollBack(id string, eventInfo EventInfo) {
	ctx := context.Background()

	db, err := GetDatabase(ctx)
	if err != nil {
		logrus.Error("get ticket : error in database ", err)
	}
	collection, err := db.Collection(ctx, viper.GetString("TICKETS_COLLECTION"))
	if err != nil {
		logrus.Error("get ticket : error in database ", err)
	}
	collection.RemoveDocument(ctx, id)
	query := fmt.Sprintf("For e in %s filter e._key == '%s' let t = e.ticket update e with {ticket : MERGE(t,{'sold': %d})} in %s", viper.GetString("EVENTS_COLLECTION"), eventInfo.eventID, eventInfo.sold, viper.GetString("EVENTS_COLLECTION"))
	fmt.Println(query)
	_, err = db.Query(ctx, query, nil)
	if err != nil {
		logrus.Error("error in Rollback", err)
	}
	logrus.Info("Rollback")

}

func PurchaseDetails(c *stripe.Charge, key string) (error) {
	ctx := context.Background()

	db, err := GetDatabase(ctx)
	if err != nil {
		logrus.Error("update Ticket : Error in database ", err)
		return err
	}
	collection, err := db.Collection(ctx, viper.GetString("TICKETS_COLLECTION"))
	if err != nil {
		logrus.Error("update Ticket : Error in collection ", err)
		return err
	}

	patch := map[string]interface{}{
		"PurchaseDetails": c,
		"Payment":         true,
	}
	_, err = collection.UpdateDocument(ctx, key, patch)
	logrus.Info("Details Updated")
	return err
}

func GetAttendees(ticketId string, isVisitor bool) ([]interface{}, error) {
	ctx := context.Background()
	db, err := GetDatabase(ctx)

	if err != nil {
		return nil, err
	}

	var query string

	if isVisitor {
		query = fmt.Sprintf(`For t in tickets
    			filter t.ID == '%s'
   				 return t.Visitors `, ticketId)
	} else {
		query = fmt.Sprintf(`For t in tickets
    			filter t.ID == '%s'
   				 return t.Attendees `, ticketId)
	}

	cursor, err := db.Query(ctx, query, nil)

	defer cursor.Close()

	var tickets []interface{}

	_, err = cursor.ReadDocument(ctx, &tickets)
	if err != nil {
		logrus.Error("get ticket : error in database ", err)
		return nil, err
	}
	return tickets, nil

}

func AttendeeCount(ticketID string) (ox_idl.AttendeesStatus, error) {

	ctx := context.Background()
	db, err := GetDatabase(ctx)

	if err != nil {
		logrus.Error("database error :  ", err)
		return ox_idl.AttendeesStatus{}, err
	}

	query := fmt.Sprintf(`FOR t IN tickets
    Filter t.ID == '%s'
    return {visitorCount : LENGTH(t.Visitors) ,attendeeCount : LENGTH(t.Attendees)}`, ticketID)

	cursor, err := db.Query(ctx, query, nil)

	var count ox_idl.AttendeesStatus

	_, err = cursor.ReadDocument(ctx, &count)
	if err != nil {
		logrus.Error("get AttendeeCount :  ", err)
		return ox_idl.AttendeesStatus{}, err
	}
	return count, nil

}

type TicketDetails struct {
	Event []struct {
		Currency  string `json:"currency"`
		Name      string `json:"name"`
		StartDate string `json:"start_date"`
		StartTime string `json:"start_time"`
		Venue     string `json:"venue"`
		Symbol    string `json:"symbol"`
		UploadId  string `json:"upload_id"`
	} `json:"event"`
	Ticket []struct {
		Amount float64 `json:"amount"`
		RegAttendees []struct {
			Confirmation string `json:"confirmation"`
			Email        string `json:"email"`
			ID           string `json:"id"`
			Name         string `json:"name"`
		} `json:"regAttendees"`
		RegVisitors	[]struct {
			Name 	string	`json:"name"`
			Email	string	`json:"email"`
		}	`json:"regVisitors"`
		EventID string `json:"eventId"`
		TotalAttendees	int	`json:"totalAttendees"`
		TotalVisitors	int	 `json:"totalVisitors"`
	} `json:"ticket"`
}

func GetTicketDetailsById(ticketId string) (TicketDetails, error) {

	ctx := context.Background()

	db, err := GetDatabase(ctx)

	var ticketDetail TicketDetails

	if err != nil {
		logrus.Error("Get ticket details by Id: failed to get ArangoDB connection ", err)
		logrus.Error("ticket id ", ticketId)
		return ticketDetail, err
	}

	query := `
		LET ticket = ( FOR t in tickets 
    		FILTER t.ID == @ticket_id
    		RETURN {
        		regAttendees: t.Attendees,
				regVisitors: t.Visitors,
        		amount: t.AmountCharged,
        		eventId: t.EventID,
				totalAttendees: t.NoOfAttendees,
				totalVisitors: t.NoOfVisitors
    		}
		)
    
		RETURN { ticket: ticket, event: ticket[0] ? (
                FOR e in events 
                    FILTER e.eventDetail.id == ticket[0].eventId
                        LET detail = e.eventDetail
                        RETURN {
                            name: detail.name,
                            start_date: detail.start_date,
							start_time: detail.start_time,
                            venue: CONCAT(TRIM(detail.venue_city), ', ', TRIM(detail.venue_state)),
                            currency: e.ticket.currency,
							upload_id: detail.cover_image_upload_id
                }): []
        }
	`

	bindVars := map[string]interface{}{"ticket_id": ticketId}

	cursor, err := db.Query(ctx, query, bindVars)

	if err != nil {
		logrus.Error("Get ticket details by Id: failed to get execute query ", err)
		logrus.Error("ticket id ", ticketId)
		return ticketDetail, err
	}
	_, err = cursor.ReadDocument(ctx, &ticketDetail)

	if err != nil {
		logrus.Error("Get ticket details by Id: failed to read document ", err)
		logrus.Error("ticket id ", ticketId)
		return ticketDetail, err
	}

	if len(ticketDetail.Event) > 0 {

		curr, err := currency.FromCurrencyName(ticketDetail.Event[0].Currency)

		if err != nil {
			logrus.Error("Get ticket details by Id: failed to read get currency symbol ", err)
			logrus.Error("ticket id ", ticketId)
			return ticketDetail, err
		}

		ticketDetail.Event[0].Symbol = curr["symbol"].(string)
	}

	return ticketDetail, nil

}

func PrintTicketForAttendee(ticketId, attendeeId string) (ox_idl.PrintTicketResponse, error) {

	ctx := context.Background()

	db, err := GetDatabase(ctx)

	var printTicketResponse ox_idl.PrintTicketResponse

	if err != nil {
		logrus.Error("print ticket for attendee: failed to get ArangoDB connection ", err)
		logrus.Error("ticket id: ", ticketId, "attendee id: ", attendeeId)
		return printTicketResponse, err
	}

	query := `
		LET ticket = (
    		FOR t in tickets 
        		FILTER t.ID == @ticket_id
        		RETURN {
					ticket_number: t.ID,
            		event_id: t.EventID,
            		purchased_by: t.Name,
            		purchased_on: t.PurchasedOn,
            		attendee: (
                		LET attendee = (
                    		FOR a in t.Attendees
                        		FILTER a.id == @attendee_id
                        		RETURN {email: a.email, name: a.name}
                		)
                		RETURN attendee[0] ? attendee[0]: {}
            		)[0]
        	}
		)

		LET event = (
    		FOR e in events 
        		FILTER e.eventDetail.id == ticket[0].event_id
        		LET d = e.eventDetail
        			RETURN {
            			event_name: d.name,
						ticket_name: e.ticket.name,
						start_date: d.start_date,
						start_time: d.start_time,
            			venue: CONCAT(TRIM(d.venue_city), ', ', TRIM(d.venue_state)),
						organizer: e.organizer.name
        	}
		)
    
		RETURN {
    		event: event[0] ? event[0] : {},
    		ticket: ticket[0] ? ticket[0] : {}
		}
	`

	bindVars := map[string]interface{}{"ticket_id": ticketId, "attendee_id": attendeeId}

	cursor, err := db.Query(ctx, query, bindVars)

	if err != nil {
		logrus.Error("print ticket for attendee: failed to query ArangoDB ", err)
		logrus.Error("ticket id: ", ticketId, "attendee id: ", attendeeId)
		return printTicketResponse, err
	}

	_, err = cursor.ReadDocument(ctx, &printTicketResponse)

	if err != nil {
		logrus.Error("print ticket for attendee: failed to read doc from ArangoDB ", err)
		logrus.Error("ticket id: ", ticketId, "attendee id: ", attendeeId)
		return printTicketResponse, err
	}

	return printTicketResponse, nil
}

func PrintAllAttendeeTickets(ticketId string) (ox_idl.PrintAllTicketsResponse, error) {

	ctx := context.Background()

	db, err := GetDatabase(ctx)

	var printTicketResponse ox_idl.PrintAllTicketsResponse

	if err != nil {
		logrus.Error("print ticket for attendee: failed to get ArangoDB connection ", err)
		logrus.Error("ticket id: ", ticketId)
		return printTicketResponse, err
	}

	query := `
		LET ticket = (
    		FOR t in tickets 
        		FILTER t.ID == @ticket_id
        		RETURN {
					ticket_number: t.ID,
            		event_id: t.EventID,
            		purchased_by: t.Name,
            		purchased_on: t.PurchasedOn,
            		attendees: FLATTEN ((
                				LET attendees = (
                    			FOR a in t.Attendees
                        			FILTER a.confirmation == 'APPROVED'
                        			RETURN {email: a.email, name: a.name}
                				)
                				RETURN attendees
                	)
            	)
        	}
		)

		LET event = (
    		FOR e in events 
        		FILTER e.eventDetail.id == ticket[0].event_id
        		LET d = e.eventDetail
        			RETURN {
            			event_name: d.name,
						ticket_name: e.ticket.name,
						start_date: d.start_date,
						start_time: d.start_time,
            			venue: CONCAT(TRIM(d.venue_city), ', ', TRIM(d.venue_state)),
						organizer: e.organizer.name
        	}
		)
    
		RETURN {
    		event: event[0] ? event[0] : {},
    		ticket: ticket[0] ? ticket[0] : {}
		}
	`

	bindVars := map[string]interface{}{"ticket_id": ticketId}

	cursor, err := db.Query(ctx, query, bindVars)

	if err != nil {
		logrus.Error("print ticket for attendee: failed to query ArangoDB ", err)
		logrus.Error("ticket id: ", ticketId)
		return printTicketResponse, err
	}

	_, err = cursor.ReadDocument(ctx, &printTicketResponse)

	if err != nil {
		logrus.Error("print ticket for attendee: failed to read doc from ArangoDB ", err)
		logrus.Error("ticket id: ", ticketId, "attendee id: ")
		return printTicketResponse, err
	}

	return printTicketResponse, nil
}

func (b *Billing) SaveBillingDetails() error {

	ctx := context.Background()

	db, err := GetDatabase(ctx)
	if err != nil {
		logrus.Error("save billing details : couldn't connect to ArangoDB", err)
		return err
	}

	collection, err := db.Collection(ctx, viper.GetString("CARD_COLLECTION"))
	if err != nil {
		logrus.Error("save billing details : Error in collection ", err)
		return err
	}
	b.CreatedOn = time.Now().UTC().Unix()

	_, err = collection.CreateDocument(ctx, b)
	if err != nil {
		logrus.Error("save billing details : Couldn't create document", err)
		return err
	}

	return nil

}

func GetAttendeeInfo(ticketId string) (AttendeeInfo, error) {
	ctx := context.Background()

	db, err := GetDatabase(ctx)

	var attendeeInfo AttendeeInfo

	if err != nil {
		logrus.Error("Get Attendee Info: failed to get ArangoDB connection ", err)
		logrus.Error("ticket id ", ticketId)
		return attendeeInfo, err
	}

	query :=
		`FOR t in tickets 
    FILTER t.ID == @ticket_id
		RETURN {
			attendees: t.NoOfAttendees,
        	regAttendees: LENGTH(t.Attendees),
			visitors: t.NoOfVisitors,
			regVisitors: LENGTH(t.Visitors)
			
    	}`

	bindVars := map[string]interface{}{"ticket_id": ticketId}
	cursor, err := db.Query(ctx, query, bindVars)
	if err != nil {
		logrus.Error("get Attendee Info : error in database ", err)
		return attendeeInfo, err
	}

	defer cursor.Close()

	_, err = cursor.ReadDocument(ctx, &attendeeInfo)
	if err != nil {
		logrus.Error("get Attendee Info : error in database ", err)
		return attendeeInfo, err
	}

	return attendeeInfo, nil
}
