# Networks
Research code for working with graphs in Go, particularly concurrent algorithms

## Basic classes
Graphs are represented by the Network struct in the Core package.  Clusters are represented by map[int] []string, where the integer is a community label generated during community detection using SLPA.  Graphs are loaded
via NetworkSerializer struct, also in Core. 

### Serialization Format
Each line of a graph represents an edge adjacency list.  The first string is the from vertex, followed by the delimiter character, followed by,
the to vertext, followed by the delimiter and the edge weight.  Edge weights are integers.  Graphs are assumed to be directed, unless the 
file is loaded with the directed parameter of LoadNetwork set to false.  In that case, an edge is added for the reciprocal direction.

# Community detection algorithms 
Presently, the Algorithms package implements the following community detection algorithms:

1. Speaker-Listener Propagation Algorithm (SLPA)

SLPA is described in Xie, Jierui and Szymanski, Boleslaw, Towards Linear Time Overlapping Community Detection in Social Networks, Proceedings of the Pacific-Asiz Conference on Knowledge Discovery and Data Mining, :25-36, 2012.
The concurrent version is adapted from Kuzmin, Konstantin, Chen, Mingming, and Szymanski, Boleslaw, Parallelizing SLPA for Scalable Overlapping Community Detection, Scientific Programming, 2015

Due to the nature of the concurrency primitives in Go and the controller architecture selecting, there is less parallelization than prescribed by Kuzmin in his algorithm.  However, speed-up is observed.

Additional algorithm implementations are planned.
