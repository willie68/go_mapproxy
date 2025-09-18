package model

import "fmt"

type Tile struct {
	System string
	Z      int
	X      int
	Y      int
}

func (t *Tile) String() string {
	return fmt.Sprintf("System: %s, Z:%d, X:%d, Y:%d", t.System, t.Z, t.X, t.Y)
}
