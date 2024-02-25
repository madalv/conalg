package models

import (
	"conalg/pb"

	gs "github.com/deckarep/golang-set/v2"
)

type Type string

const (
	FASTP_REPLY Type = "FASTP_REPLY"
	SLOWP_REPLY Type = "SLOWP_REPLY"
	RETRY_REPLY Type = "RETRY_REPLY"
)

type Response struct {
	From      string
	RequestID string
	Type      Type
	Status    bool
	Pred      gs.Set[string]
	Timestamp uint64
}

func FromFastProposeResponse(pb *pb.FastProposeResponse) Response {
	pred := gs.NewSet[string](pb.Pred...)
	return Response{
		From:      pb.From,
		RequestID: pb.ResquestId,
		Type:      FASTP_REPLY,
		Status:    pb.Result,
		Pred:      pred,
		Timestamp: pb.Time,
	}
}
