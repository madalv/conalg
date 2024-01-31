package caesar

import gs "github.com/deckarep/golang-set/v2"

const (
	FASTP_REPLY = "FASTP_REPLY"
	SLOWP_REPLY = "SLOWP_REPLY"
	ACK         = "ACK"
	NACK        = "NACK"
)

type Response struct {
	ID        string
	Type      string // FASTP_REPLY or SLOWP_REPLY
	Status    string // ACK or NACK
	Pred      gs.Set[string]
	Timestamp uint64
}
