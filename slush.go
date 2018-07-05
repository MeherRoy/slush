package main

import (
	"math/rand"
	"time"
)

type node struct {
	color int
	identity int
	incoming chan message

}

type message struct {
	color int
	ornode int
	destnode int
	msgtype int
}

func networkinit(numnodes int) []node {
	players := make([]node, numnodes)
	for i := 0; i < numnodes; i++ {
		players[i].identity = i
		players[i].color = uncolored
		players[i].incoming = make(chan message, 100)
	}
	return players
}

func update(index int, players []node, numberSample int, waitTime time.Duration) {
		go handleMsg(index int, players []node,)
		//this code handles query to other nodes

		sleeptime = rand.Intn(1000)
		time.Sleep(sleeptime*time.Millisecond)
	}
}

func query(index int, players []node, numberSample int, waitTime time.Duration) {
	if players[index].color != uncolored {
		for i := 0; i < numberSample; i++ {
			sampleNode := rand.Intn(len(players))
			players[sampleNode].incoming <- message{players[index].color, index, sampleNode, query}
		}
	}
}

func handleMsg(index int, players []node) {
	for {
		processmsg := <-players[index].incoming
		if players[index].color == uncolored || processmsg.msgtype == initialisation {
			players[index].color = processmsg.color
		} else if players[index].color == uncolored || processmsg.msgtype == query {
			players[index].color = processmsg.color
			players[processmsg.ornode].incoming <- message{players[index].color, index, processmsg.ornode, response}
		} else if players[index].color != uncolored || processmsg.msgtype == query {
			players[processmsg.ornode].incoming <- message{players[index].color, index, processmsg.ornode, response}
		}
	}
}

func clientinit(num int, color int, players []node) {
	for i := 0; i < num; i++ {
		index := rand.Intn(len(players))
		players[index].incoming <- message{color, -1, index,initialisation}
	}
}

const (
	uncolored = iota
	red
	blue
)

const (
	query = iota
	response
	initialisation
)

func main() {
	players := networkinit(100)
	go clientinit(5,blue, players)
	go clientinit(3,red, players)
	for i := 0; i < len(players); i++ {
		go update(i, players, 5, 1000)
	}
	time.Sleep(1000*time.Millisecond)
}
