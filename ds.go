package main

const (
	FAST_PEND = "FAST_PENDING"
	SLOW_PEND = "SLOW PENDING"
	ACC       = "ACCEPTED"
	REJ       = "REJECTED"
	STABLE    = "STABLE"
)

// logical clock
var global_ts = uint(0)

type Status struct {
	timestamp uint
	pred      []string
	status    string
	ballot    uint
	forced    bool
}

// map of commands and their status
var history map[string]Status

// array mapping command c to its ballot nr
var ballots map[string]uint
