package tile

import "fmt"

// Wind is one of the four winds, used for the round wind and each seat's
// wind. The zero value is not a valid wind.
type Wind uint8

// The four winds in play order.
const (
	EastWind Wind = iota + 1
	SouthWind
	WestWind
	NorthWind
)

var windNames = map[Wind]string{EastWind: "east", SouthWind: "south", WestWind: "west", NorthWind: "north"}

var windsByName = func() map[string]Wind {
	m := make(map[string]Wind, len(windNames))
	for w, n := range windNames {
		m[n] = w
	}
	return m
}()

// ParseWind returns the wind for its lowercase English name.
func ParseWind(name string) (Wind, error) {
	w, ok := windsByName[name]
	if !ok {
		return 0, fmt.Errorf("tile: invalid wind %q", name)
	}
	return w, nil
}

// IsValid reports whether w is one of the four winds.
func (w Wind) IsValid() bool {
	return w >= EastWind && w <= NorthWind
}

// String returns the lowercase English name: east, south, west, north.
func (w Wind) String() string {
	if n, ok := windNames[w]; ok {
		return n
	}
	return fmt.Sprintf("Wind(%d)", uint8(w))
}

// Tile returns the wind's tile: East for EastWind, and so on.
func (w Wind) Tile() Tile {
	return East + Tile(w-EastWind)
}

// Next returns the wind that follows in play order, wrapping north to east.
func (w Wind) Next() Wind {
	if w == NorthWind {
		return EastWind
	}
	return w + 1
}
