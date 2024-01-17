package main

type Module interface {
	Propose(command string) 
}

type Application interface {
	Execute(command string)
}

