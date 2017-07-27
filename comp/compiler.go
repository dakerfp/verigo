package comp

type DataType int

const (
	NoType DataType = iota
	BoolType
	UintType
	IntType
)

var basicTypeToDataType = map[string]DataType{
	"bool": BoolType,
	"uint": UintType,
	"int":  IntType,
}

type Module struct {
	Name     string
	Inputs   map[string]DataType
	Outputs  map[string]DataType
	Internal map[string]DataType
	Signals  map[string]DataType
}

type Function struct {
	Name    string
	Inputs  map[string]DataType
	Outputs map[string]DataType
}
