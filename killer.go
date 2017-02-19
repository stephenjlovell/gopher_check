//-----------------------------------------------------------------------------------
// ♛ GopherCheck ♛
// Copyright © 2014 Stephen J. Lovell
//-----------------------------------------------------------------------------------

package main

const (
	KILLER_COUNT = 3
)

type KEntry [4]Move // keep power of 2 array size

func (s *StackItem) StoreKiller(m Move) {
	killers := &s.killers
	switch m {
	case killers[0]:
		// no update needed.
	case killers[1]:
		killers[0], killers[1] = killers[1], killers[0]
	default:
		killers[0], killers[1], killers[2] = m, killers[0], killers[1]
	}
}

func (s *StackItem) IsKiller(m Move) bool {
	killers := &s.killers
	return m == killers[0] || m == killers[1] || m == killers[2]
}
