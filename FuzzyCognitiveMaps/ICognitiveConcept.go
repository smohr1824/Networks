package FuzzyCognitiveMap

type ICognitiveConcept interface {
	GetName() string
	SetName(name string)
	GetActivationLevel() float32
	SetActivationLevel(level float32)
}
