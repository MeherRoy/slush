package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
	"math"
	"github.com/gopherjs/gopherjs/js"
)

type Slush struct {
	players []node        // nodes in the simulation
	m       int           // number of times to sample
	k       int           // sample size
	a       float32       // percentage after which we change color [0,1)
	timeout time.Duration // timeout after which we re-sample
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

/*func update(index int, players []node, numberSample int, waitTime time.Duration) {
	//go handleMsg(index, players []node,)
	//this code handles query to other nodes

	sleeptime = rand.Intn(1000)
	time.Sleep(sleeptime * time.Millisecond)

}*/

/*func (s *Slush) query(index int,  numberSample int, waitTime time.Duration) {
	if players[index].color != uncolored {
		for i := 0; i < numberSample; i++ {
			sampleNode := rand.Intn(len(players))
			players[sampleNode].incoming <- message{players[index].color, index, sampleNode, query}
		}
	}
}*/

func (s *Slush) handleMsg(index int) {
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
	}
}

// main Slush loop
func (s *Slush) slushLoop(id int) {
	// do m rounds of sampling
	for i := 0; i < s.m; i++ {
		// Print statement for debugging
		// fmt.Println(id, "Slush", i)
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
		}
		if float32(s.players[id].countBlue) > s.a*float32(s.k) {
				s.players[id].color = Blue
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
	}

	// respond with our color
	s.players[msg.ornode].incoming <- message{s.players[id].color, id, msg.ornode, response, msg.round}
}

func (s *Slush) clientinit(num int, color int) {
	for i := 0; i < num; i++ {
		index := rand.Intn(len(s.players))
		s.players[index].incoming <- message{color, -1, index, initialisation, 0}
		// Prints the indexes for debugging
		// fmt.Println(index)
	}
}

const (
	uncolored = iota
	Red
	Blue
)

const (
	query = iota
	response
	initialisation
)

func main() {
	var numNodes int = 100
	sl := Slush{
		a: 0.51,
		m: 5,
		k: 10,
	}
	sl.networkInit(numNodes)

	go sl.clientinit(5, Blue)
	go sl.clientinit(5, Red)
	for i := 0; i < numNodes; i++ {
		go sl.handleMsg(i)
	}

	var counts [3]int

	for i:=0; i<numNodes;i++ {
		x := <- sl.acceptmsg
		counts[x]++
	}
	fmt.Println(counts)

	document := js.Global.Get("document")
	canvas := document.Call("getElementById", "cnvs")

	loc := getlocrad(400.0, numNodes, 500, 500)
	for i := range loc {
		DrawNode(canvas, int(loc[i].X), int(loc[i].Y), i,0, int(loc[i].r))
	}
}

//this function get the locations of
func getlocrad (radiusPx float64, numNodes int, centerX int, centerY int) ([]location) {
	var loc = make([]location , numNodes)
	radOffset := float64(2 * math.Pi / float64(numNodes))
	for i:=0; i< numNodes; i++ {
		radDistance := float64(i) * radOffset
		loc[i].Y = float64(centerY) + radiusPx * math.Cos(radDistance)
		loc[i].X = float64(centerX) + radiusPx * math.Sin(radDistance)
		loc[i].r = 0.4 * 2 * radiusPx * math.Sin(radOffset / 2)
	}
	return loc
}

func DrawNode(canvas *js.Object, x, y, id int, color int, radius int) {
	var fillstyle string
	// map colors blue or red to the correct JS fillstyles
	if color == uncolored {
		fillstyle = "#D3D3D3"
	} else if color == Red {
		fillstyle = "#FF0000"
 	} else if color == Blue {
		fillstyle = "#0000FF"
	}

	ctx := canvas.Call("getContext", "2d")
	// TODO get this from node color
	ctx.Set("fillStyle", fillstyle)
	// ctx.Call("fillRect", x, y, 20, 25)

	ctx.Call("beginPath")
	ctx.Call("arc",x,y,radius,0,2 * math.Pi)
	ctx.Call("fill")

	ctx.Set("font", "7px Arial")
	ctx.Set("fillStyle", "#FFFFFF")
	// TODO set to node id
	str := fmt.Sprint(id)
	ctx.Call("fillText", str, x-5, y)
}

func DrawLine(canvas *js.Object, x1, y1, x2, y2 int) {
	ctx := canvas.Call("getContext", "2d")
	ctx.Set("strokeStyle", "#000000")
	ctx.Call("moveTo", x1, y1)
	ctx.Call("lineTo", x2, y2)
	ctx.Call("stroke")
}
