package caesar

// TODO reformat this shit into a module
const (
	FAST_PEND = "FAST_PENDING"
	SLOW_PEND = "SLOW PENDING"
	ACC       = "ACCEPTED"
	REJ       = "REJECTED"
	STABLE    = "STABLE"
)

type Status struct {
	timestamp uint
	pred      []string
	status    string
	ballot    uint
	forced    bool
}

// map of commands and their status
// TODO private kv store
var history map[string]Status

// array mapping command c to its ballot nr
var ballots map[string]uint
