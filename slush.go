package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type slush struct {
	players []node        // nodes in the simulation
	m       int           // number of times to sample
	k       int           // sample size
	a       float32       // percentage after which we change color [0,1)
	timeout time.Duration // timeout after which we re-sample
}

type node struct {
	color     int
	identity  int
	started   bool
	round     int
	countRed  int
	countBlue int
	incoming  chan message
	signal    chan bool
	lock      sync.Mutex
}

type message struct {
	color    int
	ornode   int
	destnode int
	msgtype  int
	round    int
}

func (s *slush) networkInit(numnodes int) {
	s.players = make([]node, numnodes)
	for i := 0; i < numnodes; i++ {
		s.players[i].identity = i
		s.players[i].color = uncolored
		s.players[i].incoming = make(chan message, 100)
		s.players[i].signal = make(chan bool)
		//comment
	}
}

/*func update(index int, players []node, numberSample int, waitTime time.Duration) {
	//go handleMsg(index, players []node,)
	//this code handles query to other nodes

	sleeptime = rand.Intn(1000)
	time.Sleep(sleeptime * time.Millisecond)

}*/

/*func (s *slush) query(index int,  numberSample int, waitTime time.Duration) {
	if players[index].color != uncolored {
		for i := 0; i < numberSample; i++ {
			sampleNode := rand.Intn(len(players))
			players[sampleNode].incoming <- message{players[index].color, index, sampleNode, query}
		}
	}
}*/

func (s *slush) handleMsg(index int) {
	for {
		// get the msg from channel
		processmsg := <-s.players[index].incoming

		// handle init msg if we are uninitialized
		if processmsg.msgtype == initialisation && s.players[index].color == uncolored {
			s.players[index].color = processmsg.color
		}

		if processmsg.msgtype == query {
			s.OnQuery(index, processmsg)
		}

		// start main slush loop when we first get a color
		if s.players[index].color != uncolored && !s.players[index].started {
			s.players[index].started = true
			go s.slushLoop(index)
		}

		// count query replies
		if s.players[index].started && processmsg.round == s.players[index].round &&
			processmsg.msgtype == response {

			if processmsg.color == blue {
				s.players[index].lock.Lock()
				s.players[index].countBlue++
				s.players[index].lock.Unlock()
			} else {
				s.players[index].lock.Lock()
				s.players[index].countRed++
				s.players[index].lock.Unlock()
			}
			fmt.Println(index, "reply", processmsg, s.players[index].countBlue, s.players[index].countRed)
			// signal slushLoop that we got k  replies
			if s.players[index].countBlue+s.players[index].countRed == s.k {
				s.players[index].signal <- true
			}
		}

		// discard ooo msgs
		if s.players[index].started && processmsg.round != s.players[index].round &&
			processmsg.msgtype == response {
			// discard out of order msg
		}
	}
}

// main slush loop
func (s *slush) slushLoop(id int) {
	// do m rounds of sampling
	for i := 0; i < s.m; i++ {
		fmt.Println(id, "slush", i)
		s.players[id].round = i

		// sample k neigbours
		sampled := make([]bool, len(s.players))
		sampled[id] = true
		s.players[id].countRed = 0
		s.players[id].countBlue = 0
		for j := 0; j < s.k; j++ {
			sample := 0
			// keep trying to get a non-sampled node (SLOW if k ~= len(players))
			for {
				sample = rand.Intn(len(s.players))
				if !sampled[sample] {
					sampled[sample] = true
					break // stop retrying
				}
			}
			// send query to sample node
			s.players[sample].incoming <- message{s.players[id].color, id, sample, query, i}

		}
		// wait for replies
		<-s.players[id].signal
		s.players[id].lock.Lock()
		if float32(s.players[id].countRed) > s.a*float32(s.k) {
			s.players[id].color = red
		}
		if float32(s.players[id].countBlue) > s.a*float32(s.k) {
				s.players[id].color = blue
		}
		s.players[id].lock.Unlock()
	}

	fmt.Println(id, " accepted ", s.players[id].color)
}

// Process msg
func (s *slush) OnQuery(id int, msg message) {
	// If uninitialized set our color to the color of the query sender
	if s.players[id].color == uncolored {
		s.players[id].color = msg.color
	}

	// respond with our color
	s.players[msg.ornode].incoming <- message{s.players[id].color, id, msg.ornode, response, msg.round}
}

func (s *slush) clientinit(num int, color int) {
	for i := 0; i < num; i++ {
		index := rand.Intn(len(s.players))
		s.players[index].incoming <- message{color, -1, index, initialisation, 0}
		fmt.Println(index)
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

	sl := slush{
		a: 0.51,
		m: 5,
		k: 10,
	}
	sl.networkInit(100)

	go sl.clientinit(5, blue)
	go sl.clientinit(3, red)
	for i := 0; i < 100; i++ {
		go sl.handleMsg(i)
	}

	time.Sleep(1000 * time.Millisecond)
}
