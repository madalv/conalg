package models

import (
	"conalg/pb"
	"fmt"

	gs "github.com/deckarep/golang-set/v2"
)

const (
	FASTP_REPLY = "FASTP_REPLY"
	SLOWP_REPLY = "SLOWP_REPLY"
	RETRY_REPLY = "RETRY_REPLY"
)

type Response struct {
	From      string
	RequestID string
	Type      string
	Status    bool
	Pred      gs.Set[string]
	Timestamp uint64
	Ballot    uint
}

func NewResponse(reqID string, t string, status bool, pred gs.Set[string], from, ts uint64, ballot uint) Response {
	return Response{
		From:      fmt.Sprintf("NODE_%d", from),
		RequestID: reqID,
		Type:      t,
		Status:    status,
		Pred:      pred,
		Timestamp: ts,
		Ballot:    ballot,
	}
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
		Ballot:    uint(pb.Ballot),
	}
}

func ToFastProposeResponse(r Response) *pb.FastProposeResponse {
	return &pb.FastProposeResponse{
		ResquestId: r.RequestID,
		Ballot:     uint64(r.Ballot),
		Time:       r.Timestamp,
		Pred:       r.Pred.ToSlice(),
		Result:     r.Status,
		From:       r.From,
	}
}
