package FuzzyCognitiveMap

type IMultilayerCognitiveConcept interface {
	LayerActivationLevels() []ILayerActivationLevel
	GetName() string
	SetName(name string)
	GetActivationLevel() float32
	SetActivationLevel(lvl float32)
}

type ILayerActivationLevel interface {
	Coordinates() string
	Level() float32
}

type layerActivationLevel struct {
	coords string
	level  float32
}

func (ml layerActivationLevel) Coordinates() string {
	return ml.coords
}

func (ml *layerActivationLevel) Level() float32 {
	return ml.level
}

func NewLayerActivationLevel(coordinates string, lvl float32) *layerActivationLevel {
	retVal := layerActivationLevel{coords: coordinates, level: lvl}
	return &retVal
}
