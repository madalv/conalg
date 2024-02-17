package main

import (
	"conalg/caesar"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gookit/slog"
)

// TODO add start/end times for requests to track


// TODO move this bs somewhere else
type SampleApp struct {
	conalg caesar.Conalg
}

func (s *SampleApp) DetermineConflict(c1, c2 []byte) bool {
	return len(c1) == len(c2)
}

func (s *SampleApp) Execute(c []byte) {
	slog.Infof(" ... doing whatever i want with %s", c)
}

func (s *SampleApp) SetConalgModule(m caesar.Conalg) {
	s.conalg = m
}

func main() {

	app := SampleApp{}
	conalg := caesar.InitConalgModule(&app)
	app.SetConalgModule(conalg)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	router.POST("/propose", func(c *gin.Context) {
		var json struct {
			Command string `json:"command"`
		}
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		app.conalg.Propose([]byte(json.Command))
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := os.Getenv("SAMPLEAPP_PORT")
	router.Run(port)
}
