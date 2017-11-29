package main

import (
	"github.com/smohr1824/Networks/Core"
	"fmt"
)

func main() {
	network := Core.NewNetwork(true)
	network.AddEdge("A", "B", 1.0)
	network.AddEdge("B", "C", 2.0)
	network.AddEdge("B", "D", 3.0)
	network.AddEdge("C", "D", 2.0)
	network.AddEdge("D", "B", 1.0)

	err := network.AddEdge("A","A", 1.0)
	if err != nil {
		fmt.Println(err.Error())
	}

	degree, _ := network.OutDegree("B")
	A := network.AdjacencyMatrix()
	fmt.Printf("Degree of B is %d\n", degree )
	network.RemoveEdge("C","D")
	A = network.AdjacencyMatrix()
	_ = A[0][0]

}
