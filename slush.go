package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
	"net/http"
	"log"
)

type Slush struct {
	players   []node        // nodes in the simulation
	m         int           // number of times to sample
	k         int           // sample size
	a         float32       // percentage after which we change color [0,1)
	timeout   time.Duration // timeout after which we re-sample
	acceptmsg chan int
}

type location struct {
	X float64
	Y float64
	r float64
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

var numNodes = 100

var sl = Slush{
	a: 0.51,
	m: 5,
	k: 10,
}

var colorChange chan ClientData

func (s *Slush) networkInit(numnodes int) {
	s.players = make([]node, numnodes)
	s.acceptmsg = make(chan int, numnodes)
	for i := 0; i < numnodes; i++ {
		s.players[i].identity = i
		s.players[i].color = uncolored
		s.players[i].incoming = make(chan message, 100)
		s.players[i].signal = make(chan bool)

		//comment
	}
}

//
func (s *Slush) handleMsg(index int) {
	for {
		// get the msg from channel
		processmsg := <-s.players[index].incoming

		// handle init msg if we are uninitialized
		if processmsg.msgtype == initialisation && s.players[index].color == uncolored {
			s.players[index].color = processmsg.color
			colorChange <- ClientData{index,colors[s.players[index].color]}
			time.Sleep(1000*time.Millisecond)
		}

		if processmsg.msgtype == query {
			s.OnQuery(index, processmsg)
		}

		// start main Slush loop when we first get a color
		if s.players[index].color != uncolored && !s.players[index].started {
			s.players[index].started = true
			go s.slushLoop(index)
		}

		// count query replies
		if s.players[index].started && processmsg.round == s.players[index].round &&
			processmsg.msgtype == response {

			if processmsg.color == Blue {
				s.players[index].lock.Lock()
				s.players[index].countBlue++
				s.players[index].lock.Unlock()
			} else {
				s.players[index].lock.Lock()
				s.players[index].countRed++
				s.players[index].lock.Unlock()
			}
			// logs all the communication
			// fmt.Println(index, "reply", processmsg, s.players[index].countBlue, s.players[index].countRed)
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
		time.Sleep(1000 * time.Millisecond)
	}
}

// main Slush loop
func (s *Slush) slushLoop(id int) {
	// do m rounds of sampling
	for i := 0; i < s.m; i++ {
		// Print statement for debugging
		fmt.Println(id, "Slush", i)
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
			s.players[id].color = Red
			colorChange <- ClientData{id,colors[s.players[id].color]}
		}
		if float32(s.players[id].countBlue) > s.a*float32(s.k) {
			s.players[id].color = Blue
			colorChange <- ClientData{id,colors[s.players[id].color]}
		}
		s.players[id].lock.Unlock()
	}
	//Print statement printing when a node accepts a particular message
	// fmt.Println(id, " accepted ", s.players[id].color)
	s.acceptmsg <- s.players[id].color
}

// Process msg
func (s *Slush) OnQuery(id int, msg message) {
	// If uninitialized set our color to the color of the query sender
	if s.players[id].color == uncolored {
		s.players[id].color = msg.color
		colorChange <- ClientData{id,colors[s.players[id].color]}
		time.Sleep(1000*time.Millisecond)
	}

	// respond with our color
	s.players[msg.ornode].incoming <- message{s.players[id].color, id, msg.ornode, response, msg.round}
}

func (s *Slush) clientinit(num int, color int) {
	for i := 0; i < num; i++ {
		index := rand.Intn(len(s.players))
		s.players[index].incoming <- message{color, -1, index, initialisation, 0}
	}
}

func (s *Slush) startSimulation() {
	s.networkInit(numNodes)
	for i := 0; i < numNodes; i++ {
		colorChange <- ClientData{i,colors[uncolored]}
	}
	time.Sleep(1000*time.Millisecond)
	go s.clientinit(1, Blue)
	go s.clientinit(1, Red)

	for i := 0; i < numNodes; i++ {
		go s.handleMsg(i)
	}

	var counts [3]int

	for i := 0; i < numNodes; i++ {
		x := <-s.acceptmsg
		counts[x]++
	}
	fmt.Println(counts)
}

const (
	uncolored = iota
	Red
	Blue
)

var colors = [3]string{
	"#AAAAAA", "#FF0000", "#0000FF",
}

const (
	query = iota
	response
	initialisation
)

func main() {
	colorChange = make(chan ClientData, 100)
	rand.Seed(time.Now().UTC().UnixNano())
	// Simulation starts inside handler
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", handler)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}