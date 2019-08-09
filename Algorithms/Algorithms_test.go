package Algorithms

import (
	"github.com/smohr1824/Networks/Core"
	"testing"
)

func TestSLPABasic(t *testing.T) {
	ser := Core.NewNetworkSerializer("|")
	G, err := ser.ReadNetworkFromFile("displays2.dat", false)
	if err != nil {
		t.Errorf("Error reading test file")
	}

	communities := ConcurrentSLPA(G, 20, 0.3, 3000, 2, 2)
	if len(communities) == 0 {
		t.Errorf("No communities found")
	} else {
		t.Logf("Found %d communities", len(communities))
		t.Log(communities)
	}
}

func TestBipartiteBasic(t *testing.T) {
	G := Core.NewNetwork(false)

	for i:= 0; i < 1001; i += 2 {
		G.AddVertex(uint32(i))
		G.AddVertex(uint32(i + 1))
	}

	var err error
	for i := 0; i < 1001; i += 2 {
		err = G.AddEdge(uint32(i), uint32(i + 1), 1.0)
	}
	if err != nil {
		t.Error("Error adding edge")
	} else {
		for k := 1001; k > 1; k -= 2 {
			err = G.AddEdge(uint32(k), uint32(k-3), 1.0)
		}
	}

	if err != nil {
		t.Error("Error adding edge")
	}
	isIt, R, B := ConcurrentBipartite(G, 4)
	if !isIt {
		t.Errorf("Bipartite network found not to be bipartite")
	} else 	{
		t.Log("Bipartite")
		t.Logf("R is %d items long", len(R))
		t.Logf("B is %d items long", len(B))
	}

	consistent := true
	if R[0] % 2 != 0 {
		// red set is odd -- check for all odds in R, all evens in B
		for i := 0; i < len(R); i++ {
			if R[i] % 2 == 0 {
				consistent = false
			}
		}
		for k:= 0; k < len(B); k++ {
			if B[k] % 2 != 0 {
				consistent = false
			}
		}
	} else {
		// red set is even -- check for all evens in R, all odds in B
		for i := 0; i < len(R); i++ {
			if R[i] % 2 != 0 {
				consistent = false
			}
		}
		for k:= 0; k < len(B); k++ {
			if B[k] % 2 == 0 {
				consistent = false
			}
		}
	}
	if !consistent {
		t.Error("At least one inconsistent member was found in one of the sets")
	}
}