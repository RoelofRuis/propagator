package propagator

import (
	"math"
)

// Probability is a floating point value indicating the chance that an index will be picked.
type Probability = float32

// Priority is a non-negative value indicating the priority with which an index will be picked.
// Lower values precede higher values, 0 being the lowest possible.
type Priority = uint32

type packedProbPrio int64

func packPriorityProbability(probability Probability, priority Priority) packedProbPrio {
	probabilityBits := math.Float32bits(probability)
	priorityBits := priority

	return packedProbPrio(uint64(probabilityBits)<<32 | uint64(priorityBits))
}

func unpackPriorityProbability(p packedProbPrio) (Probability, Priority) {
	probabilityBits := uint32(p >> 32)
	priorityBits := uint32(p & 0xFFFFFFFF)

	return math.Float32frombits(probabilityBits), priorityBits
}
