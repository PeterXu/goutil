package util

type BitSet []bool

func NewBitSet(number int) BitSet {
	bs := make(BitSet, number)
	bs.ResetAll()
	return bs
}

// Returns the number of bits in the bitset that are set
func (b BitSet) Count() int {
	count := 0
	for _, v := range b {
		if v {
			count += 1
		}
	}
	return count
}

// Returns the number of bits in the bitset
func (b BitSet) Size() int {
	return len(b)
}

// Returns whether the bit at position pos is set
func (b BitSet) Test(pos int) bool {
	if pos >= 0 && pos < len(b) {
		return b[pos]
	} else {
		return false
	}
}

// Returns whether any of the bits is set
func (b BitSet) Any() bool {
	for _, v := range b {
		if v {
			return true
		}
	}
	return false
}

// Returns whether none of the bits is set
func (b BitSet) None() bool {
	return b.Count() == 0
}

// Returns whether all of the bits in the bitset are set
func (b BitSet) All() bool {
	return b.Count() == len(b)
}

// Sets the bit at position pos
func (b BitSet) Set(pos int) {
	if pos >= 0 && pos < len(b) {
		b[pos] = true
	}
}

// Sets all bits in the bitset
func (b BitSet) SetAll() {
	for i, _ := range b {
		b[i] = true
	}
}

// Resets the bit at position pos
func (b BitSet) Reset(pos int) {
	if pos >= 0 && pos < len(b) {
		b[pos] = false
	}
}

// Resets all bits in the bitset
func (b BitSet) ResetAll() {
	for i, _ := range b {
		b[i] = false
	}
}

// Flips the bit at position pos
func (b BitSet) Flip(pos int) {
	if pos >= 0 && pos < len(b) {
		b[pos] = !b[pos]
	}
}

// Flips all bits in the bitset
func (b BitSet) FlipAll() {
	for i, v := range b {
		b[i] = !v
	}
}
