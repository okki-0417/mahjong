package tile

// DoraCount counts the dora among tiles: one for each tile that is the dora
// of an indicator, plus one for each red five. A red five that is also
// indicated counts twice. How the count turns into han (none for a yakuman)
// is left to scoring.
func DoraCount(tiles, indicators []Tile) int {
	count := 0
	for _, t := range tiles {
		for _, indicator := range indicators {
			if t.SameKind(indicator.Dora()) {
				count++
			}
		}
		if t.IsRed() {
			count++
		}
	}
	return count
}
