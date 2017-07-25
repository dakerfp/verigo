package testdata

// type Posedge bool // XXX
// type Negedge bool // XXX

type Mux22 struct {
	A   int  `input`
	B   int  `input`
	Sel bool `input`

	Output int `output`
}

func (m Mux22) Output() int {
	if m.Sel {
		return m.B
	} else {
		return m.A
	}
}

type Forward struct {
	In int `input`

	Fwd int `output`
}

func (f Forward) Fwd() int {
	return f.In
}

type Sync struct {
	Rst bool `input`
	Clk bool `input`
}

type SumItUp struct {
	Sync

	A     int  `input`
	Start bool `input`

	Sum int `output`
}

func (sit SumItUp) Sum(Clk Posedge) int {
	if sit.Start {
		return sit.A
	} else {
		return sit.Sum + sit.A
	}
}

type ALU struct {
	Op byte `input`
	A  uint `input`
	B  uint `input`

	Result uint `output`
}

func (alu ALU) Result() uint {
	switch alu.Op {
	case 0:
		return alu.A + alu.B
	case 1:
		return alu.A - alu.B
	case 2:
		return alu.A * alu.B
	case 3:
		return alu.A / alu.B
	default:
		return alu.A
	}
}
