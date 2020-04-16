package mail

import (
	"google.golang.org/grpc"
	"github.com/graphicweave/ox/mail"
	"context"
	"github.com/spf13/viper"
	"github.com/sirupsen/logrus"
)

/**
 * Created by  ♅ Salfi Farooq on 23/06/17.
 */

func _connect() (*grpc.ClientConn, error) {
	if conn, err := grpc.Dial(viper.GetString("INJUN_GRPC_ADDR"), grpc.WithInsecure()); err != nil {
		logrus.Error("Error in Connection", err.Error())
		return nil, err
	} else {
		return conn, nil
	}
}

func AttendeeConfirmation(attendeeInfo map[string]interface{}) {
	if conn, err := _connect(); err != nil {
		logrus.Error("Error in Connection", err.Error())
	} else {

		defer conn.Close()

		client := mail.NewMailServiceClient(conn)

		attendee := &mail.AttendeeConfirmationDetail{
			EventName: attendeeInfo["name"].(string),
			EventFullVenue: attendeeInfo["venue_name"].(string) + " " +
				attendeeInfo["address"].(string) + " " +
				attendeeInfo["venue_state"].(string),
			EventDateTime: attendeeInfo["start_date"].(string) + " " +
				attendeeInfo["start_time"].(string),
			EmailId:               attendeeInfo["emailId"].(string),
			EventCoverImage:       attendeeInfo["cover_image_upload_id"].(string),
			Name:                  attendeeInfo["username"].(string),
			EventOrganizerCompany: attendeeInfo["organizer"].(string),
			EventURL:              viper.GetString("EVENT_URL") + attendeeInfo["id"].(string),
			ConfirmationURL:       viper.GetString("CONFIRMATION_URL")+attendeeInfo["userId"].(string)+"/"+attendeeInfo["ticketId"].(string),
		}

		_, e := client.ConfirmAttendee(context.Background(), attendee)

		if e != nil {
			logrus.Error("couldn't send confirmation email to attendee", e)
		}

	}

}

func VisitorEmail(visitorInfo map[string]interface{}) {
	if conn, err := _connect(); err != nil {
		logrus.Error("Error in Connection", err.Error())
	} else {

		defer conn.Close()

		client := mail.NewMailServiceClient(conn)

		visitor := &mail.VisitorDetail{
			EventName: visitorInfo["name"].(string),
			EventFullVenue: visitorInfo["venue_name"].(string) + " " +
				visitorInfo["address"].(string) + " " +
				visitorInfo["venue_state"].(string),
			EventDateTime: visitorInfo["start_date"].(string) + " " +
				visitorInfo["start_time"].(string),
			EmailId:               visitorInfo["emailId"].(string),
			EventCoverImage:       visitorInfo["cover_image_upload_id"].(string),
			Name:                  visitorInfo["username"].(string),
			EventOrganizerCompany: visitorInfo["organizer"].(string),
			EventURL:              viper.GetString("EVENT_URL") + visitorInfo["id"].(string),
			Coordinates: []float64{30.03, 74.01},
		}

		_, e := client.VisitorEmail(context.Background(), visitor)

		if e != nil {
			logrus.Error("couldn't send confirmation email to visitor", e)
		}

	}
}

func ConfirmationEmail(detail mail.ConfirmationRequestDetail) {
	if conn, err := _connect(); err != nil {
		logrus.Error("Error in Connection", err.Error())
	} else {

		defer conn.Close()

		client := mail.NewMailServiceClient(conn)
		_, e := client.ConfirmationEmail(context.Background(), &detail)

		if e != nil {
			logrus.Error("couldn't send confirmation email", e)
		}
	}
}