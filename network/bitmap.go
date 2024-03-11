package network

type Bitmap struct {
	Data []int64 `json:"data,omitempty"`
	Size int     `json:"size,omitempty"`
}

func newBitmap(size int) Bitmap {
	numWords := (size + 63) / 64
	bitmap := make([]int64, numWords)
	return Bitmap{
		bitmap, size,
	}
}

/*
set 将目标位设置成1
*/
func (b Bitmap) set(bitIndex int) {
	wordIndex := bitIndex / 64
	bitOffset := bitIndex % 64
	b.Data[wordIndex] |= 1 << bitOffset
}

/*
clear 将目标位置这只成0
*/
func (b Bitmap) clear(bitIndex int) {
	wordIndex := bitIndex / 64
	bitOffset := bitIndex % 64
	var mask int64 = ^(1 << bitOffset) // ^b 表示对b按位取反。
	b.Data[wordIndex] &= mask
}

/*
get 获取指定位置的值
*/
func (b Bitmap) get(bitIndex int) bool {
	wordIndex := bitIndex / 64
	bifOffset := bitIndex % 64
	return (b.Data[wordIndex] & (1 << bifOffset)) != 0
}

/*
len 获取位图的大小 (位数量)
*/
func (b Bitmap) len() int {
	return b.Size
}
