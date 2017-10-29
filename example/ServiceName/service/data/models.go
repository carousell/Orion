package data

import "github.com/carousell/DataAccessLayer/dal/marshaller"

type Message struct {
	Msg  marshaller.NullString `json:"message"`
	Time marshaller.NullTime   `json:"time"`
	UUID marshaller.NullString `json:"uuid,primary"`
}
