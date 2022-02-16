package goutil

var _ssrcMap map[uint32]bool

const (
	kMinSsrc uint32 = 0x0000ffff
	kMaxSsrc uint32 = 0xfffffff0
)

func init() {
	_ssrcMap = make(map[uint32]bool)
}

func _genSSRC(min, max uint32) uint32 {
	var ssrc uint32
	for ssrc <= min || ssrc >= max {
		high := (RandomUint32() << 16)
		low := RandomUint32()
		ssrc = (high + low) & max
	}
	return ssrc
}

func CreateSSRC() uint32 {
	for {
		ssrc := _genSSRC(kMinSsrc, kMaxSsrc)
		if _, ok := _ssrcMap[ssrc]; !ok {
			_ssrcMap[ssrc] = true
			return ssrc
		}
	}
}

func RegisterSSRC(ssrc uint32) {
	_ssrcMap[ssrc] = true
}

func ReturnSSRC(ssrc uint32) {
	delete(_ssrcMap, ssrc)
}
