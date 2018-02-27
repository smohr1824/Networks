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

// Speaker-Listener Propagation Algorithm for community detection
// Main interest is concurrent implementation for use on multisocket/multicore machines

// Concurrent SLPA algorithm is adapted from:
// Konstantin Kuzmin, Mingming Chen, and Boleslaw K. Szymanski, “Parallelizing SLPA for Scalable Overlapping Community Detection,” Scientific Programming, vol. 2015, Article ID 461362, 18 pages, 2015. doi:10.1155/2015/461362
// Core SLPA is adapted from "Towards Linear Time Overlapping Community Detection in Social Networks", 2012, Jierui Xie and Boleslaw Szymanski as implemented in
// https://github.com/smohr1824/Graphs

package Algorithms

import (
	"github.com/smohr1824/Networks/Core"
	//"fmt"
	"time"
	"math/rand"
	"sync"
)

// message used by a goroutine to communicate which iteration it is on to the main calling routine, or for
// the main routine to communicate same to dependent goroutines
type IterationMessage struct {
	RoutineId int			// ordinal id of the goroutine sending the iteration count
	IterationNumber int		// ordinal count of the iteration the sender is processing
}

type DependencyList struct {
	RoutineId int					// ordinal id of the goroutine sending the list of dependencies
	ExternalDependencies []string	// array of vertex ids that are external to the sending goroutine
}

type PartitionBoundary struct {
	FirstIdx int
	LastIdx int
}

type LabelObservation struct {
	Key   interface{}
	Value interface{}
}

// Concurrent implementation of SLPA community detection algorithm per Kuzman, Chen, Szymanski 2015
func ConcurrentSLPA(G *Core.Network, iterations int, threshold float64, seed int64, concurrentCount int, minCommunitySize int) map[int][]string {
	vertices := G.Vertices(true)
	order := G.Order()
	// can't have more partitions than nodes; every partition must have at least one node
	if concurrentCount > order {
		concurrentCount = order
	}
	partSize := order/concurrentCount	// very important: integer division -- last partition may have up to partSize + (partSize - 1) nodes
	partCount := concurrentCount

	partition := 0	// one-based index of partitions

	// construct the partitions for each goroutine running concurrently
	partitionSlices := make([][]string, partCount)
	for i := 0; i < partCount; i++ {
		if partition < partCount - 1 {
			// new partition, populate the slice for the old partition
			low := i * partSize
			high := low + partSize
			partitionSlices[partition] = vertices[low:high]
			partition++
		} else {
				low := i * partSize
				high := order
				partitionSlices[partition] = vertices[low:high]
		}
	}

	// building the dependencies here is essential to the control structure synchronizing the goroutines,
	// and creating the lists of nodes with neighbors outside the partition and nodes with neighbors inside the partition is a natural side effect.
	// Unfortunately, it sacrifices some concurrency
	dependsOnList:= make([][]int, concurrentCount)
	dependencyToList := make([][]int, concurrentCount)

	internals := make([][]string, concurrentCount)	// for each partition, a list of nodes in the partition whose neighbors are wholly within the partition
	externals := make([][]string, concurrentCount)	// for each partition, a list of nodes with external dependencies

	// iterate through the partition slices and establish the list of partitions that depend on this one (dependencyToList),
	// and the list of partitions that this one depends on (dependsOnList)
	// We need these for the control structure (messaging), so we cannot, sadly, do this concurrently by partition.
	for partitionIdx, partition := range partitionSlices {
		for _, nodeId := range partition {
			hasExternalDependencies := false
			nodeNeighbors := G.GetNeighbors(nodeId)
			for neighbor, _ := range nodeNeighbors {
				if indexStringInSlice(neighbor, partition) == -1 {
					// has neighbors external to the partition
					hasExternalDependencies = true
					// not in the partition, use integer division to get the partition number where it is found (dependency on the partitioning schema!)
					// also assume that the network is consistent and neighbor really is in some partition of the network
					indexInVertices := indexStringInSlice(neighbor, vertices)	// can optimize by skipping the current partition
					foundIn := indexInVertices/partSize
					if foundIn >= concurrentCount {
						foundIn = concurrentCount - 1
					}
					if !intInSlice(foundIn, dependsOnList[partitionIdx]) {
						dependsOnList[partitionIdx] = append(dependsOnList[partitionIdx], foundIn)
					}
					if !intInSlice(partitionIdx, dependencyToList[foundIn]) {
						dependencyToList[foundIn] = append(dependencyToList[foundIn], partitionIdx)
					}
				}
			}
			// assign the node in the partition to either the list of nodes with one or more neighbors outside the partition of the list of nodes with
			// all neighbors inside the partition
			if hasExternalDependencies {
				externals[partitionIdx] = append(externals[partitionIdx], nodeId)
			} else {
				internals[partitionIdx] = append(internals[partitionIdx], nodeId)
			}
		}
	}
	currentIteration := make([] int, concurrentCount)

	canIGoChannel := make(chan IterationMessage, 2*concurrentCount)	// multiplexed channel for goroutines to ask if they can proceed
	permissionStatus := make([]bool, concurrentCount)				// true if a goroutine is awaiting permission to proceed to the next iteration
	goChannels := make([]chan bool, concurrentCount)				// one channel per goroutine to signal proceed with processing, dependencies complete

	// Allocate a global map of node names to maps of labels observed
	// Each observation is the node index and the number of times that label was observed.
	// Yes, this flies in the face of the Go pattern of passing copies rather than working on one structure.
	// However, partitions will need to access labels from nodes outside their partition, and the concurrency scheme
	// ensures out of partition access is read-only AND synchronized such that dependency labels are correct before they are accessed.
	nodeLabelMemory := new(sync.Map)
	InitLabels(&vertices, 0, order - 1, nodeLabelMemory)
	for i:= 0; i < concurrentCount; i++ {
		currentIteration[i] = 0
		permissionStatus[i] = false
		goChannels[i] = make(chan bool, 1)

		//go PartitionSLPA(i, G, &vertices, partitionBoundaries[i], externals[i], internals[i], seed, iterations, nodeLabelMemory, canIGoChannel, goChannels[i])
		go PartitionSLPA(i, G, &vertices, &externals[i], &internals[i], seed, iterations, nodeLabelMemory, canIGoChannel, goChannels[i])
	}


	activeRoutines := concurrentCount
	for ; activeRoutines > 0; {
		select {
			// goroutine is asking for permission to proceed to the next iteration or is reporting finished
			case permissionMsg := <-canIGoChannel:
				currentIteration[permissionMsg.RoutineId] = permissionMsg.IterationNumber
				// if the iteration is -1, the goroutine is finished, so collect the results
				if currentIteration[permissionMsg.RoutineId] == -1 {
					//fmt.Println(fmt.Sprintf("Goroutine %d ending", permissionMsg.RoutineId))
					activeRoutines--
				}
				permissionStatus[permissionMsg.RoutineId] = true

				// if the dependencies of this partition are ready, signal ok and change permission status to false (not pending)
				if !DependenciesNotReady(permissionMsg.IterationNumber, dependsOnList[permissionMsg.RoutineId], currentIteration) || activeRoutines == 1 {
					goChannels[permissionMsg.RoutineId] <- true
					permissionStatus[permissionMsg.RoutineId] = false
				}

				// range over the list of partitions dependent on the requesting partition and signal them if they are active, pending, and this makes them ready
				for _, partitionIdx := range dependencyToList[permissionMsg.RoutineId] {
					if currentIteration[partitionIdx] != -1 && permissionStatus[partitionIdx] && !DependenciesNotReady(currentIteration[partitionIdx], dependsOnList[partitionIdx], currentIteration) {
						goChannels[partitionIdx] <- true
						permissionStatus[partitionIdx] = false
					}
				}

			default:
				time.Sleep(5 * time.Millisecond)
		}
	}
	close(canIGoChannel)
	for i:=0; i < concurrentCount; i++ {
		close(goChannels[i])
	}
	return PostProcess(nodeLabelMemory, threshold, minCommunitySize)
}


// Function executed as a concurrent goroutine
// Performs SLPA labelling for a partition of nodes
// Returns a list of nodes and their observed labels
// Vertices is passed by reference to avoid copying very large arrays
func PartitionSLPA(routineID int, G *Core.Network, vertices *[]string, externals *[]string, internals *[]string, seed int64, iterations int, nodeLabels *sync.Map, askChannel chan<- IterationMessage, waitChannel <-chan bool) {
	externalIndices := make([]int, len(*externals)) // indices of the external nodes relative to the start of partition

	r := rand.New(rand.NewSource(seed))
	for i := 0; i < len(*externals); i++ {
		//externalIndices[i] = indexStringInSlice(externals[i], partition)
		externalIndices[i] = indexStringInSlice((*externals)[i], *vertices)
	}
	internalIndices := make([]int, len(*internals)) // indices of the internal nodes relative to the start of partition
	for i := 0; i < len(*internals); i++ {
		//internalIndices[i] = indexStringInSlice(internals[i], partition)
		internalIndices[i] = indexStringInSlice((*internals)[i], *vertices)
	}
	for i := 0; i < iterations; i++ {
		DoOneIteration(vertices, G, &externalIndices, nodeLabels, r)

		if i < iterations - 1 {
			permissionSlip := IterationMessage{RoutineId: routineID, IterationNumber: i + 1}
			askChannel <- permissionSlip
			<-waitChannel

			// proceed with internal dependencies
			DoOneIteration(vertices, G, &internalIndices, nodeLabels, r)
		} else {

			// proceed with internal dependencies
			DoOneIteration(vertices, G, &internalIndices, nodeLabels, r)

			// wait to send termination as we don't want to drop out of the control loop in main
			terminationSlip := IterationMessage{RoutineId: routineID, IterationNumber: -1}
			askChannel <- terminationSlip
		}
	}
}

func DoOneIteration(nodes *[]string, G *Core.Network, indices *[]int, nodeLabels *sync.Map, r *rand.Rand) {
	// rand.Perm does a pseudo-random permutation of the digits [0..len(indices)], so convert to the actual node index
	for _, i := range rand.Perm(len(*indices)) {
		nodeID := (*nodes)[(*indices)[i]]
		labelsSeen := make(map[int] int)
		neighbors := G.GetNeighbors(nodeID)

		if len(neighbors) == 0 {
			continue
		}
		for neighbor, _ := range neighbors {
			labelMap, ok := nodeLabels.Load(neighbor)
			if ok {
				m := labelMap.(*sync.Map)
				dist := NewMultinomialLabels(m, r)
				label := dist.NextSample()
				count, ok := labelsSeen[label]
				if ok {
					labelsSeen[label] = count + 1

				} else {
					labelsSeen[label] = 1
				}
			}
		}

		maxLabel := MaxLabel(labelsSeen, r)

		listenerMap, ok := nodeLabels.Load(nodeID)
		m := listenerMap.(*sync.Map)
		if ok {
			count, ok := m.Load(maxLabel)
			if ok {
				m.Store(maxLabel, count.(int)+1)
			} else {
				m.Store(maxLabel, 1)
			}
			nodeLabels.Store(nodeID, m)
		}
	}
}

func InitLabels(nodes *[]string, partitionStart int, partitionEnd int, observations *sync.Map) {
	for idx := partitionStart; idx <= partitionEnd; idx++ {
		label := idx + partitionStart
		labels := new(sync.Map)
		labels.Store(label, 1)
		observations.Store((*nodes)[idx], labels)
	}
}

// for a map of label observations, return the label seen most often
// This label is also the ordinal index of the node in the graphs sorted list of nodes
// If the labels are tied, make a random selection.
func MaxLabel(labelsSeen map[int]int, r *rand.Rand) int {
	biggestV := 0
	biggestK := -1
	lastV := -1
	same := true
	labels := make([]int, 0, len(labelsSeen))
	for k, v := range labelsSeen {
		if v > biggestV {
			biggestV = v
			biggestK = k
		}
		if v != lastV && lastV != -1 {
			same = false
		}
		lastV = v
		labels = append(labels, k)
	}
	if same && len(labels) > 1 {
		return labels[r.Intn(len(labelsSeen))]
	} else {
		return biggestK
	}
}

func SumLabels(labels []LabelObservation) int {
	retVal := 0
	for _, label := range labels {
		//retVal += label.Value
		retVal += label.Value.(int)
	}
	return retVal

}

// Removes any label which does not clear the threshold (percentage of total observations), then
// constructs a map of labels and their associated node ids
func PostProcess(masterLabelMap *sync.Map, threshold float64, minSize int) map[int][]string {
	masterLabelMap.Range(func(k, v interface{}) bool {
		ApplyThreshold(v.(*sync.Map), threshold)
		return true
	})

	communities := make(map[int][]string)
	// iterate through the master map of label maps and make the label the key, then append the located
	// node id to the list associated with the label
	masterLabelMap.Range(func(key, ls interface{}) bool {
		ls.(*sync.Map).Range(func(k, v interface{}) bool {
			list, ok := communities[k.(int)]
			if ok {
				list = append(list, key.(string))
				communities[k.(int)] = list
			} else {
				community := make([]string, 0, 5)
				community = append(community, key.(string))
				communities[k.(int)] = community
			}

			return true
		})
		return true
	})
	if minSize > 1 {
		deletes := make([]int,0, len(communities))
		for label, cmty := range communities {
			if len(cmty) < minSize {
				deletes = append(deletes, label)
			}
		}
		for _, label := range deletes {
			delete(communities, label)
		}
	}
	return communities
}

func ApplyThreshold(labelset *sync.Map, threshold float64) {
	observed := make([]LabelObservation,0)
	labelset.Range(func(k, v interface{}) bool {
		observation := LabelObservation{Key: k.(int), Value: v.(int)}
		observed = append(observed, observation)
		return true
	})

	sum := SumLabels(observed)
	for _, observation := range observed {

		if observation.Value.(int) < int(float64(sum)*threshold) {
			labelset.Delete(observation.Key.(int))
		}
	}

}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func intInSlice(i int, list []int) bool{
	for _, b := range list {
		if b == i {
			return true
		}
	}
	return false
}

func indexStringInSlice(a string, list []string) int {
	for i, b := range list {
		if b == a {
			return i
		}
	}
	return -1
}

func DependenciesNotReady(myiteration int, deps []int, iterations []int) bool {
	for _, dep := range deps {
		if myiteration != -1 && iterations[dep] != -1 && Abs(myiteration - iterations[dep]) > 1 {
			return true
		}
	}
	return false
}

func Abs(num int) int {
	if num < 0 {
		return -num
	} else {
		return num
	}

}