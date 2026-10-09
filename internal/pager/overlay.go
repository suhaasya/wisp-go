package pager

// Overlay holds staged edit markers for one result (LW-029).
type Overlay struct {
	Mod [][]int
	Ins []int
	Del []int
}

// OverlayWire is the PRD overlay object on window responses.
type OverlayWire struct {
	Mod [][]int `json:"mod"`
	Ins []int   `json:"ins"`
	Del []int   `json:"del"`
}

func (o Overlay) wire() OverlayWire {
	mod := o.Mod
	if mod == nil {
		mod = [][]int{}
	}
	ins := o.Ins
	if ins == nil {
		ins = []int{}
	}
	del := o.Del
	if del == nil {
		del = []int{}
	}
	return OverlayWire{Mod: mod, Ins: ins, Del: del}
}
