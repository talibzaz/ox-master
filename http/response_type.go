package http

import (
	"github.com/graphicweave/ox/database/model"
	log "github.com/sirupsen/logrus"
)

type AnyResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type TicketDetail struct {
	Response AnyResponse `json:"response"`
	Details  interface{} `json:"details"`
}

func (d TicketDetail) GetTicketDetails(ticketId string) TicketDetail {

	ticketDetail := TicketDetail{}
	response := AnyResponse{}

	details, err := model.GetTicketDetailsById(ticketId)

	if err != nil {
		log.Errorln("failed to get ticket details for ticket", ticketId)
		log.Errorln("ERR", err)
		response.Status = "ERR"
		response.Message = "Failed to get "
		ticketDetail.Response = response
		return ticketDetail
	}

	response.Status = "OK"
	ticketDetail.Response = response
	ticketDetail.Details = details

	return ticketDetail
}