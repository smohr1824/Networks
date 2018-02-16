// Copyright 2017 - 2018 Stephen T. Mohr, OSIsoft, LLC
// MIT License

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package main

import (
	"github.com/smohr1824/Networks/Core"
	"github.com/smohr1824/Networks/Algorithms"
	"fmt"
	"runtime"
	"math/rand"
	"time"
)
type iterationMsg struct {
	id int
	iteration int
}

func main() {
	cpuSetting := runtime.GOMAXPROCS(0)
	cpuAvailable := runtime.NumCPU()
	fmt.Println(fmt.Sprintf("Max number of CPUs/threads to use: %d", cpuSetting))
	fmt.Println(fmt.Sprintf("Available CPUs/threads: %d", cpuAvailable))

	/*network := Core.NewNetwork(true)
	network.AddEdge("0", "1", 1.1)
	network.AddEdge("0", "2", 2.2)
	network.AddEdge("1", "2", 3.14159)
	network.AddEdge("1", "3", 2.4)
	network.AddEdge("3", "0", 1.1)
	network.AddEdge("3", "2", 1.1)*/

	/*network2 := Core.NewNetwork(true)
	network2.AddEdge("A", "B", 1)
	network2.AddEdge("B", "C", 1)
	network2.AddEdge("C", "D", 3.14159)
	network2.AddEdge("E", "F", 2.4)
	network2.AddEdge("F", "G", 1.1)
	network2.AddEdge("A", "C", 1.1)
	network2.AddEdge("D", "E", 1)


	Algorithms.ConcurrentSLPA(*network2, 16, 2, 3, 3)*/

	/*network = Core.NewNetwork(false)
	network.AddEdge("A", "B", 1)
	network.AddEdge("A", "C", 1)
	network.AddEdge("A", "D", 2)
	network.AddEdge("B", "D", 4)
	network.AddEdge("C", "D", 3)
	network.AddEdge("D", "E", 1)
	network.AddEdge("E", "F", 5)
	network.AddEdge("E", "J", 1)
	network.AddEdge("E", "H", 2)
	network.AddEdge("F", "I", 3)
	network.AddEdge("H", "G", 1)
	network.AddEdge("F", "G", 1)
	network.AddEdge("G", "J", 1)
	network.AddEdge("I", "J", 1)
	network.AddEdge("H", "I", 2)

	communities:=Algorithms.ConcurrentSLPA(network, 20,.4, time.Now().Unix(), 2)*/

	sr := Core.NewDefaultNetworkSerializer()
	network, err := sr.ReadNetworkFromFile("C:\\Users\\smohr\\GoglandProjects\\src\\github.com\\smohr1824\\Networks\\TestApp\\tenset.dat", false)
	if err != nil {
		fmt.Println("Error on read")
		return
	}

	start := time.Now()
	communities := Algorithms.ConcurrentSLPA(network, 20, .4, time.Now().Unix(), 3)
	end := time.Now()

	dur := end.Sub(start)
	fmt.Println(fmt.Sprintf("Duration: %f", dur.Seconds()))
	fmt.Println(fmt.Sprintf("%d communities found", len(communities)))
	fmt.Println(communities)
	/*network3 := Core.NewNetwork(true)
	network3.AddEdge("A", "B", 1)
	network3.AddEdge("A", "C", 1)
	network3.AddEdge("C", "F", 1)
	network3.AddEdge("D", "C", 1)
	network3.AddEdge("E", "C", 1)
	network3.AddEdge("F", "G", 1)
	Algorithms.ConcurrentSLPA(network3, 16, 2, 2, 1)
	// status channel used by all goroutines
	canIGoChannel := make(chan iterationMsg, 8)
	currentIteration := make([] int, 4)
	permissionStatus := make([]bool, 4)
	var goChannels [4]chan bool

	// unique union of in edges for the nodes of a partition
	dependsOnList:= make([][]int, 4)
	dependencyToList := make([][]int, 4)
	dependsOnList[0] = []int {3}
	dependsOnList[1] = []int {0}
	dependsOnList[2] = []int {0, 1, 3}
	dependsOnList[3] = []int {1}

	// unique union of out edges for the nodes of a partition
	dependencyToList[0] = []int {1, 2}
	dependencyToList[1] = []int {2, 3}
	// partition 2 is zero-length
	dependencyToList[3] = []int {0, 2}

	for i:= 0; i < 4; i++ {
		currentIteration[i] = 0
		permissionStatus[i] = false
		// buffer size 1 permits the main routine to be non-blocking unless it outruns the goroutines, i.e., more than one iteration ahead
		goChannels[i] = make(chan bool, 1)
		go DoLabeling(i, canIGoChannel, goChannels[i], 4, network)

	}

	activeRoutines := 4
	for ; activeRoutines > 0; {
		select {
			case permissionMsg := <- canIGoChannel:
				currentIteration[permissionMsg.id] = permissionMsg.iteration
				if currentIteration[permissionMsg.id] == -1 {
					fmt.Println(fmt.Sprintf("Goroutine %d ending", permissionMsg.id))
					activeRoutines--
				}
				permissionStatus[permissionMsg.id] = true

				// if the dependencies of this partition are ready, signal ok and change permission status to false (not pending)
				if !DependenciesNotReady(permissionMsg.iteration, dependsOnList[permissionMsg.id], currentIteration) {
					goChannels[permissionMsg.id] <- true
					permissionStatus[permissionMsg.id] = false
				}

				// range over the list of partitions dependent on the requesting partition and signal them if they are active, pending, and this makes them ready
				for _, partitionIdx := range dependencyToList[permissionMsg.id] {
					if currentIteration[partitionIdx] != -1 && permissionStatus[partitionIdx] && !DependenciesNotReady(currentIteration[partitionIdx], dependsOnList[partitionIdx], currentIteration) {
						goChannels[partitionIdx] <- true
						permissionStatus[partitionIdx] = false
					}
				}
			default:
				time.Sleep(10 * time.Millisecond)

		}
	} */

	//
}

func DoLabeling(routineId int, askChannel chan <-  iterationMsg, waitChannel <- chan bool, numIterations int, network *Core.Network){
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)

	for i:= 0; i < numIterations; i++ {
		msg := NewIterationMessage(routineId, i)
		askChannel <- msg
		<-waitChannel
		// sleep in lieu of SLPA iteration
		//fmt.Println(fmt.Sprintf("Goutine %d executing iteration %d", routineId, i))
		dur := r.Intn(100)
		time.Sleep(time.Duration(dur) * time.Millisecond)
	}
	msg := NewIterationMessage(routineId, -1)
	askChannel <- msg
}

func main2() {

	cpuSetting := runtime.GOMAXPROCS(0)
	cpuAvailable := runtime.NumCPU()
	dummy := ""

	var goroutineChannels [4]chan iterationMsg	// channels for communicating the current iteration of dependencies to dependents
	switchBoard := make(chan iterationMsg, 8)	// channel for telling main what iteration a goroutine is on
	var testdeps [4][]int
	testdeps[1] = []int {3}
	testdeps[2] = []int {0, 1}
	testdeps[3] = []int {1}

	active := 4
	network := Core.NewNetwork(true)
	network.AddEdge("A", "B", 1.1)
	network.AddEdge("B", "C", 2.2)
	network.AddEdge("B", "D", 3.14159)
	network.AddEdge("C", "D", 2.4)
	network.AddEdge("D", "B", 1.1)

	depList:= make([][]int, 4)
	depList[0] = []int {2}
	depList[1] = []int {3}
	depList[2] = []int {0, 1}
	depList[3] = []int {1}

	status := make([] int, 4)

	for i:=0; i < 4; i++ {

		status[i] = 0
		goroutineChannels[i] = make(chan iterationMsg, 4)
		go DoLabels(i, switchBoard, goroutineChannels[i], 4, 4, depList[i], network)
		//msg := NewIterationMessage(i, 0)
		//goroutineChannels[i] <- msg
	}

	time.Sleep(1000 * time.Millisecond)
	for ; active > 0; {
		select {
			case msg := <- switchBoard:
				status[msg.id] = msg.iteration
				if (msg.iteration == -1) {
					active--;
					fmt.Println(fmt.Sprintf("Ending goroutine %d", msg.id ))
				} else {
					for j := 0; j < 4; j++ {
						// don't send status to the goroutine that sent it, and don't try to send to a finished goroutine (blocks)
						if j != msg.id && status[j] != -1 {
							goroutineChannels[j] <- msg
						}
					}
				}
			default:
				time.Sleep(100 * time.Millisecond)
		}
	}


	time.Sleep(2000 * time.Millisecond)

	for k:= 0; k < 4; k++ {
		close(goroutineChannels[k])
	}
	fmt.Println(fmt.Sprintf("Max number of CPUs/threads to use: %d", cpuSetting))
	fmt.Println(fmt.Sprintf("Available CPUs/threads: %d", cpuAvailable))
	fmt.Println("Press ENTER to end")
	fmt.Scanln(&dummy)


	err := network.AddEdge("A","A", 1.0)
	if err != nil {
		fmt.Println(err.Error())
	}

	degree, _ := network.OutDegree("B")
	A := network.AdjacencyMatrix()
	s := Core.NewDefaultNetworkSerializer()
	s.WriteNetworkToFile(network, "C:\\temp\\gonetwork.dat")
	newNet, err := s.ReadNetworkFromFile("C:\\temp\\gonetwork.dat", true)
	_ = newNet
	fmt.Printf("Degree of B is %d\n", degree )
	network.RemoveEdge("C","D")
	A = network.AdjacencyMatrix()
	_ = A[0][0]

	readNet, err := s.ReadNetworkFromFile("C:\\temp\\gonetwork2.dat", true)
	if err != nil {
		fmt.Printf(err.Error())
	} else {
		fmt.Printf("Order of read network is %d", readNet.Order())
	}

	s.WriteNetworkToFile(readNet, "C:\\temp\\gonetwork2_out.dat")



}

func DoLabels(routineId int, homePhone chan <- iterationMsg, notifications  <- chan iterationMsg, numRoutines int, numIterations int, deps []int,  network *Core.Network) {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	status := make([] int, numRoutines)

	for q:=0; q < numRoutines; q++ {
		status[q] = 0
	}

	i := 0
	breakOut := false
	for ; !breakOut; {
		select {
			case msg := <- notifications:
				status[msg.id] = msg.iteration
				continue
			default:
				if (DependenciesNotReady(i, deps, status)) {
					continue
				}	else {
					dur := r.Intn(100)
					time.Sleep(time.Duration(dur) * time.Millisecond)
					if i <= numIterations {
						msg := NewIterationMessage(routineId, i + 1)
						homePhone <- msg
					} else {
						msg := NewIterationMessage(routineId, -1)
						homePhone <- msg
						breakOut = true
					}

					i++
				}

		}

	}

}

func NewIterationMessage(id int, iteration int) iterationMsg {
	msg := new (iterationMsg)
	msg.id = id
	msg.iteration = iteration
	return *msg
}

func DependenciesNotReady(myiteration int, deps []int, iterations []int) bool {
	for _, dep := range deps {
		if myiteration - iterations[dep] > 1 {
			return true
		}
	}
	return false
}
