package util

var _ssrcMap map[uint32]uint32

func init() {
	_ssrcMap = make(map[uint32]uint32)
}

func _genSSRC() uint32 {
	var ssrc uint32
	for ssrc <= 0x0000ffff || ssrc >= 0xfffffff0 {
		high := (RandomUint32() << 16)
		low := RandomUint32() & 0xfffffff0
		ssrc = high + low
	}
	return ssrc
}

func CreateSSRC() uint32 {
	for {
		ssrc := _genSSRC()
		if _, ok := _ssrcMap[ssrc]; !ok {
			_ssrcMap[ssrc] = 0
			return ssrc
		}
	}
	return 0
}

func RegisterSSRC(ssrc uint32) {
	_ssrcMap[ssrc] = 0
}

func ReturnSSRC(ssrc uint32) {
	delete(_ssrcMap, ssrc)
}
