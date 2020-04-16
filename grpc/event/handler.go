package event

import (
	"github.com/graphicweave/ox/ox_idl"
	"github.com/graphicweave/ox/database/model"
	"github.com/sirupsen/logrus"
	"context"
)

/**
 * Created by  ♅ Salfi Farooq on 23/06/17.
 */

type OXService struct {
}

func (ox *OXService) GetTicketsByID(t *ox_idl.TicketRequest, server ox_idl.OX_GetTicketsByIDServer) error {
	c := make(chan ox_idl.Ticket)
	respChan := model.ResponseChan{
		Done: make(chan struct{}),
		Err:  make(chan error),
	}
	go model.GetTicketsByUserID(t.Id, c, respChan)
	var err error
DONE:
	for {
		select {
		case <-respChan.Done:
			{
				close(respChan.Err)
				break DONE
			}
		case err = <-respChan.Err:
			{
				close(respChan.Done)
				close(respChan.Err)
				logrus.Error("couldn't retrieve tickets", err)
				break DONE
			}
		case ticket := <-c:
			server.Send(&ticket)
		}
	}

	defer func() {
		close(c)
	}()
	return err
}

func (ox *OXService) GetAttendeesStatus(ctx context.Context, t *ox_idl.TicketRequest) (*ox_idl.AttendeesStatus, error) {

	count, err := model.AttendeeCount(t.Id)

	if err != nil {
		logrus.Error("couldn't fetch Attendee/Visitor count for ticket : %s", t.Id)
		return nil, err

	}

	logrus.Info("Attendee count sent")
	return &count, nil
}

func (ox *OXService) ConfirmAttendee(ctx context.Context, a *ox_idl.ConfirmAttendeeTicket) (*ox_idl.Response, error) {

	err := model.ConfirmAttendeeEmail(a.UserId, a.TicketId)

	if err != nil {
		logrus.Error("couldn't confirm email", err)
		return &ox_idl.Response{Status: "ERR", Message: "Couldn't confirm Email."}, err
	}
	return &ox_idl.Response{Status: "OK", Message: "Email Confirmed."}, nil
}

func (ox *OXService) PrintAttendeeTicket(ctx context.Context, request *ox_idl.PrintTicketRequest) (*ox_idl.PrintTicketResponse, error)  {

	resp, err := model.PrintTicketForAttendee(request.TicketId, request.AttendeeId)

	if err != nil {
		logrus.Errorln("faield to get print ticket response")
		logrus.Errorln("ERR: ", err)
		return nil, err
	}
	return &resp, nil
}

func (ox *OXService) PrintAllAttendeeTickets(ctx context.Context, request *ox_idl.PrintAllTicketsRequest) (*ox_idl.PrintAllTicketsResponse, error)  {

	resp, err := model.PrintAllAttendeeTickets(request.TicketId)

	if err != nil {
		logrus.Errorln("faield to get print ticket response")
		logrus.Errorln("ERR: ", err)
		return nil, err
	}
	return &resp, nil
}
