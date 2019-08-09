# Networks
Research code for working with graphs in Go, particularly concurrent algorithms

## Basic classes
Graphs are represented by the Network struct in the Core package.  Clusters are represented by map[int] []uint32, where the integer is a community label generated during community detection using SLPA, and each uint32 is a vertex id.  Graphs are loaded
via NetworkSerializer struct, also in Core. 

### Serialization Format
The supported serialization format is GML. A streaming, tokenized approach is now supported providing some resiliancy in the face of variations in the use of whitespace (e.g., placement of opening and closing brackets). GML arrays are not yet supported.
Low level routines are available for extracting all properties of a list including unknown properties, but unknown properties 
are not retained. These routines exist to support fuzzy cognitive maps.  GML support will be extended as needed by the research project.

Network serialization supports the following deprecated legacy format. Each line of a graph represents an edge adjacency list.  The first uint32 is the from vertex, followed by the delimiter character, followed by
the to vertex, followed by the delimiter and the edge weight.  Edge weights are floats.  Graphs are assumed to be directed, unless the 
file is loaded with the directed parameter of LoadNetwork set to false.

# Community detection algorithms 
Presently, the Algorithms package implements the following community detection algorithms:

1. Speaker-Listener Propagation Algorithm (SLPA)

SLPA is described in Xie, Jierui and Szymanski, Boleslaw, Towards Linear Time Overlapping Community Detection in Social Networks, Proceedings of the Pacific-Asia Conference on Knowledge Discovery and Data Mining, :25-36, 2012.
The concurrent version is adapted from Kuzmin, Konstantin, Chen, Mingming, and Szymanski, Boleslaw, Parallelizing SLPA for Scalable Overlapping Community Detection, Scientific Programming, 2015

Due to the nature of the concurrency primitives in Go and the controller architecture selecting, there is less parallelization than prescribed by Kuzmin in his algorithm.  However, speed-up is observed.

Additional algorithm implementations are planned.

# Other Algorithms
ConcurrentBipartite tests a network for bipartness.  If successful, the two sets of vertices are returned as uint32[] where the uint32 is the vertex id.
