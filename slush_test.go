package main

import (
	"testing"
	"fmt"
)

func Test5050(t *testing.T) {
	var countBad int
	for i:=0; i<1000; i++ {
		var numNodes int = 100
		sl := Slush{
			a: 0.51,
			m: 10,
			k: 20,
		}
		sl.networkInit(numNodes)

		go sl.clientinit(5, Blue)
		go sl.clientinit(5, Red)
		for i := 0; i < numNodes; i++ {
			go sl.handleMsg(i)
		}

		var countsexp [3]int

		for j:=0; j<numNodes;j++ {
			x := <-sl.acceptmsg
			countsexp[x]++
		}

		if countsexp[0] != 0 {
			t.Error("something stayed uninitialised", countsexp, i)
			countBad++
		}
		if countsexp[1] > 0 {
			if countsexp[2] != 0 {
				t.Error("consensus error, nodes partitioned", countsexp, i)
				countBad++
			}
			if countsexp[1] != numNodes {
				t.Error("some nodes not in consensus", countsexp, i)
				countBad++
			}
		}
		if countsexp[2] > 0 {
			if countsexp[1] != 0 {
				t.Error("consensus error, nodes partitioned", countsexp, i)
			}
			if countsexp[2] != numNodes {
				t.Error("some nodes not in consensus", countsexp, i)
			}
		}
	}
	fmt.Println(countBad)
}

