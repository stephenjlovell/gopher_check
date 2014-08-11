package main

/*
static int lsb(unsigned long bitboard) {
  return (__builtin_ctzl(bitboard));
}
static int msb(unsigned long bitboard) {
  return (63-__builtin_clzl(bitboard));
}
static int pop_count(unsigned long bitboard) {
  return (__builtin_popcountl(bitboard));
}
*/
import (
	"C"
)

func lsb(b BB) int {
  return int(C.lsb(C.ulong(b)))
}

func msb(b BB) int {
  return int(C.msb(C.ulong(b)))
}

func furthest_forward(c uint8, b BB) int {
  if c == WHITE {
    return int(C.lsb(C.ulong(b)))  // LSB
  } else {
    return int(C.msb(C.ulong(b)))  // MSB
  }
}

func pop_count(b BB) int {
  return int(C.pop_count(C.ulong(b)))
}



