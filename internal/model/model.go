package model

import "fmt"

type Tile struct {
	Provider string
	Z        int
	X        int
	Y        int
}

func (t *Tile) String() string {
	return fmt.Sprintf("Provider: %s, Z:%d, X:%d, Y:%d", t.Provider, t.Z, t.X, t.Y)
}
