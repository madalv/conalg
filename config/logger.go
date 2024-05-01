package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/lmittmann/tint"
)

func ConfigLogger(lvl slog.Level) {
	handler := tint.NewHandler(os.Stdout, &tint.Options{
		Level:      lvl,
		TimeFormat: time.RFC3339,
		AddSource:  true,
	})

	l := slog.New(handler)
	slog.SetDefault(l)
}

const (
	ID            = "id"
	STATUS        = "status"
	TIMESTAMP     = "timestamp"
	ERR           = "err"
	WAITLIST      = "waitlist"
	PRED          = "pred"
	UPDATE_ID     = "upd_id"
	UPDATE_PRED   = "upd_pred"
	UPDATE_STATUS = "upd_status"
	OUTCOME       = "outcome"
	FROM          = "from"
	TYPE          = "type"
	RESULT        = "result"
	REPLY         = "reply"
	CFG           = "cfg"
	DIFF          = "diff"
	ADDR          = "address"
	NR_LIST       = "nr_ls"
)
