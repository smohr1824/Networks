# Networks
Research code for working with graphs in Go, particularly concurrent algorithms

## Basic classes
Graphs are represented by the Network struct in the Core package.  Clusters are represented by map[int] []uint32, where the integer is a community label generated during community detection using SLPA, and each uint32 is a vertex id.  Graphs are loaded
via NetworkSerializer struct, also in Core. 

Multilayer networks are now supported by the library. Support for node and categorical coupling is provided.

### Serialization Format
The supported serialization format is GML. A streaming, tokenized approach is now supported providing some resiliancy in the face of variations in the use of whitespace (e.g., placement of opening and closing brackets). GML arrays are not yet supported.
Low level routines are available for extracting all properties of a list including unknown properties, but unknown properties 
are not retained. These routines exist to support fuzzy cognitive maps.  GML support is provided for both monolayer and multilayer networks and will be
the primary format for monolayer networks going forward. It is the only supported format for multilayer networks.

Multilayer networks are serialized using an unofficial extension of the published GML format. A multilayer GML document consists of the directed property, followed by one or more layer records.  Layer records contain the coordinates of the 
layer followed by the GML serialization of the graph making up the layer.  After all layers are written, zero or more edge records are written to capture explicit interlayer edges.  Each edge contains lists for the source, target, and weight of the edge. 
Unlike monolayer sources and targets, each node has id and coordinates properties in a list. The weight property is a simple property.

Network serialization of monolayer networks supports the following deprecated legacy format. Each line of a graph represents an edge adjacency list.  The first uint32 is the from vertex, followed by the delimiter character, followed by
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
ConcurrentBipartite tests a network for biparteness.  If successful, the two sets of vertices are returned as uint32[] where the uint32 is the vertex id.

# Fuzzy Cognitive Maps
The FCM namespace adds basic fuzzy cognitive map capability utilizing the Network class behind the scenes. 
The threshold function for map inference may be set to bivalent, trivalent, or logistic by specifying an enumerated type, or the user may implement a custom 
function by creating a function of the form float f(float32 sum). If no threshold function is specified, the map defaults to bivalent. Similarly, the user may 
select the classic or modified Kosko equation for map inference.  The default is classic. 
The Step method of the FuzzyCognitiveMap class performs one generation of inference using algorithmic methods, executing with 
 O(|V| + |E|) complexity.
 
 Multilayer fuzzy cognitive maps are supported, as well. Inference is as with monolayer fuzzy cognitive maps. 
 Influences are explicit, i.e., categorical coupling is not used. This has the effect of letting each layer execute as a loosely coupled subsystem (coupled only by explicit interlayer edges), thereby permitting insight into the behavior of 
  different components of the overall map. Resolution of the overall activation level of a concept is acheived by summing the elementary layer activation layers and applying the transfer function and update rule. Consequently, elementary layers are normalized components 
  of the multilayer map, i.e., the concept activation level of any elementary layer has influence equal to the concept activation level of any other layer. Explicit interlayer edges should be used when a particular elementary layer dominates a concept. 