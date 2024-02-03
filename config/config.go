package config

import (
	"math"
	"os"
	"strings"
)

type Config struct {
	FastQuorum    int
	ClassicQuorum int
	Port          string
	Nodes         []string
	ID            string
}

func NewConfig() Config {
	// err := godotenv.Load()

	port := os.Getenv("PORT")
	nodes := os.Getenv("NODES")
	id := os.Getenv("ID")
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
		ID:            id,
	}

	return cfg
}
