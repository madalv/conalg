package config

import (
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/gookit/slog"
	"github.com/joho/godotenv"
)

type Config struct {
	FastQuorum    int
	ClassicQuorum int
	Port          string
	Nodes         []string
	ID            uint64
}

func NewConfig() Config {
	err := godotenv.Load()

	port := os.Getenv("PORT")
	nodes := os.Getenv("NODES")
	id, err := strconv.Atoi(os.Getenv("ID"))
	if err != nil {
		slog.Fatal(err)
	}
	nodesSplit := strings.Split(nodes, ",")
	nrNodes := float64(len(nodesSplit))

	// ceiling(3N/4)
	fastQuorum := math.Ceil(nrNodes * 3 / 4)
	// floor(N/2) + 1
	classicQuorum := math.Floor(nrNodes/2) + 1

	cfg := Config{
		Nodes:         nodesSplit,
		Port:          port,
		ClassicQuorum: int(classicQuorum),
		FastQuorum:    int(fastQuorum),
		ID:            uint64(id),
	}

	return cfg
}
