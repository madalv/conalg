package models

import gs "github.com/deckarep/golang-set/v2"

type StatusUpdate struct {
	RequestID string
	Status    Status
	Pred      gs.Set[string]
}
