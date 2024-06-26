package model

import (
	"fmt"

	gs "github.com/deckarep/golang-set/v2"
	"log/slog"
	"github.com/madalv/conalg/pb"
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
	Result    bool
	Pred      gs.Set[string]
	Timestamp uint64
	Ballot    uint
}

func NewResponse(reqID string, t string, status bool, pred gs.Set[string], from, ts uint64, ballot uint) Response {
	return Response{
		From:      fmt.Sprintf("NODE_%d", from),
		RequestID: reqID,
		Type:      t,
		Result:    status,
		Pred:      pred,
		Timestamp: ts,
		Ballot:    ballot,
	}
}

func FromResponsePb(pb *pb.Response) Response {
	if pb == nil {
		slog.Error("pb.Response is nil")
		return Response{}
	}
	if pb.Pred == nil {
		pb.Pred = []string{}
	}
	pred := gs.NewSet[string](pb.Pred...)
	return Response{
		From:      pb.From,
		RequestID: pb.ResquestId,
		Type:      pb.Type,
		Result:    pb.Result,
		Pred:      pred,
		Timestamp: pb.Timestamp,
		Ballot:    uint(pb.Ballot),
	}
}

func ToResponsePb(r Response) *pb.Response {
	if r.Pred == nil {
		r.Pred = gs.NewSet[string]()
	}
	return &pb.Response{
		ResquestId: r.RequestID,
		Ballot:     uint64(r.Ballot),
		Timestamp:  r.Timestamp,
		Pred:       r.Pred.ToSlice(),
		Result:     r.Result,
		From:       r.From,
		Type:       r.Type,
	}
}
