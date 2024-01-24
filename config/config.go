package config

import (
	"math"
	"os"
	"strings"
)

type Config struct {
	FastQuorum    uint
	ClassicQuorum uint
	Port          string
	Nodes         []string
}

func NewConfig() Config {
	// err := godotenv.Load()

	port := os.Getenv("PORT")
	nodes := os.Getenv("NODES")
	nodesSplit := strings.Split(nodes, ",")
	nrNodes := float64(len(nodesSplit))

	// ceiling(3N/4)
	fastQuorum := math.Ceil(nrNodes * 3 / 4)
	// floor(N/2) + 1
	classicQuorum := math.Floor(nrNodes/2) + 1

	cfg := Config{
		Nodes:         nodesSplit,
		Port:          port,
		ClassicQuorum: uint(classicQuorum),
		FastQuorum:    uint(fastQuorum),
	}

	return cfg
}
