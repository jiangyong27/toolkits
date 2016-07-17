package util

type BitMap struct {
	data     []uint64
	capacity uint64
}

func NewBitMap(capacity uint64) *BitMap {
	data_len := (capacity + 63) / 64
	return &BitMap{
		capacity: capacity,
		data:     make([]uint64, data_len, data_len),
	}
}

func (bm *BitMap) Set(pos uint64) {
	if pos >= bm.capacity {
		return
	}

	bm.data[pos/64] |= (0x01 << (pos % 64))
}

func (bm *BitMap) Clr(pos uint64) {
	if pos >= bm.capacity {
		return
	}

	bm.data[pos/64] &^= (0x01 << (pos % 64))

}

func (bm *BitMap) Get(pos uint64) bool {
	if pos >= bm.capacity {
		return false
	}
	index, offset := pos/64, pos%64
	return bm.data[index]&(0x01<<offset) == (0x01 << offset)
}
