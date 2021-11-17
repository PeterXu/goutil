package util

/*
#cgo CFLAGS: -Wno-deprecated -std=c99

#include "src/h264_parser.c"

typedef uint8_t* uint8ptr_t;
*/
import "C"

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"
)

/*
 *  0                   1                   2                   3
 *  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |V=2|P|X|  CC   |M|     PT      |       sequence number         |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |                           timestamp                           |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 * |           synchronization source (SSRC) identifier            |
 * +=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+=+
 * |            contributing source (CSRC) identifiers             |
 * |                             ....                              |
 * +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
 */

const (
	kRtpVersion      = 2
	kRtpHeaderLength = 12

	// the first byte
	kRtpVersionShift   = 6
	kRtpVersionMask    = 0x3
	kRtpPaddingShift   = 5
	kRtpPaddingMask    = 0x1
	kRtpExtensionShift = 4
	kRtpExtensionMask  = 0x1
	kRtpCcMask         = 0xF

	// the second byte
	kRtpMarkerShift = 7
	kRtpMarkerMask  = 0x1
	kRtpPtMask      = 0x7F

	// the 3th - 12th bytes
	kRtpSeqNumOffset    = 2
	kRtpTimestampOffset = 4
	kRtpSsrcOffset      = 8
	kRtpCsrcOffset      = 12
)

type RtpHeader struct {
	Version          uint8
	Padding          bool
	Extension        bool
	Marker           bool
	PayloadOffset    int
	PayloadType      uint8
	SequenceNumber   uint16
	Timestamp        uint32
	SSRC             uint32
	CSRC             []uint32
	ExtensionProfile uint16
	ExtensionPayload []byte

	RtpExtension     RtpExtension
	HeaderLength     uint32
	PaddingLength    uint32
	PayloadFrequency uint32
}

type RtpExtension struct {
	HasTransmissionTimeOffset  bool
	TransmissionTimeOffset     int32
	HasAbsoluteSendTime        bool
	AbsoluteSendTime           uint32
	HasTransportSequenceNumber bool
	TransportSequenceNumber    uint16

	// Audio Level includes both level in dBov and voiced/unvoiced bit. See:
	// https://datatracker.ietf.org/doc/draft-lennox-avt-rtp-audio-level-exthdr/
	HasAudioLevel bool
	VoiceActivity bool
	AudioLevel    uint8

	Has_video_timing  bool
	Has_frame_marking bool

	// For identification of a stream when ssrc is not signaled. See
	// https://tools.ietf.org/html/draft-ietf-avtext-rid-09
	// TODO(danilchap): Update url from draft to release version.
	Stream_id          string
	Repaired_stream_id string

	// For identifying the media section used to interpret this RTP packet. See
	// https://tools.ietf.org/html/draft-ietf-mmusic-sdp-bundle-negotiation-38
	Mid string
}

// Marshal serializes the header into bytes.
func (h *RtpHeader) Marshal() (buf []byte, err error) {
	buf = make([]byte, h.MarshalSize())
	if n, err := h.MarshalTo(buf); err != nil {
		return nil, err
	} else {
		return buf[:n], nil
	}
}

// MarshalSize returns the size of the header once marshaled.
func (h *RtpHeader) MarshalSize() int {
	// NOTE: Be careful to match the MarshalTo() method.
	size := kRtpHeaderLength + (len(h.CSRC) * 4)
	if h.Extension {
		size += 4 + len(h.ExtensionPayload)
	}
	return size
}

// Unmarshal parses the header
func (h *RtpHeader) Unmarshal(rawPacket []byte) error {
	if len(rawPacket) < kRtpHeaderLength {
		return fmt.Errorf("RTP header size insufficient: %d", len(rawPacket))
	}

	h.Version = rawPacket[0] >> kRtpVersionShift & kRtpVersionMask
	h.Padding = (rawPacket[0] >> kRtpPaddingShift & kRtpPaddingMask) > 0
	h.Extension = (rawPacket[0] >> kRtpExtensionShift & kRtpExtensionMask) > 0
	h.CSRC = make([]uint32, rawPacket[0]&kRtpCcMask)

	h.Marker = (rawPacket[1] >> kRtpMarkerShift & kRtpMarkerMask) > 0
	h.PayloadType = rawPacket[1] & kRtpPtMask

	h.SequenceNumber = binary.BigEndian.Uint16(rawPacket[kRtpSeqNumOffset : kRtpSeqNumOffset+2])
	h.Timestamp = binary.BigEndian.Uint32(rawPacket[kRtpTimestampOffset : kRtpTimestampOffset+4])
	h.SSRC = binary.BigEndian.Uint32(rawPacket[kRtpSsrcOffset : kRtpSsrcOffset+4])

	if h.Padding {
		idx := len(rawPacket) - 1
		h.PaddingLength = uint32(rawPacket[idx])
	}

	currOffset := kRtpCsrcOffset + (len(h.CSRC) * 4)
	if len(rawPacket) < currOffset {
		return fmt.Errorf("RTP header size insufficient; %d < %d", len(rawPacket), currOffset)
	}

	for i := range h.CSRC {
		offset := kRtpCsrcOffset + (i * 4)
		h.CSRC[i] = binary.BigEndian.Uint32(rawPacket[offset:])
	}

	if h.Extension {
		if len(rawPacket) < currOffset+4 {
			return fmt.Errorf("RTP header size insufficient for extension; %d < %d", len(rawPacket), currOffset)
		}

		h.ExtensionProfile = binary.BigEndian.Uint16(rawPacket[currOffset:])
		currOffset += 2
		extensionLength := int(binary.BigEndian.Uint16(rawPacket[currOffset:])) * 4
		currOffset += 2

		if len(rawPacket) < currOffset+extensionLength {
			return fmt.Errorf("RTP header size insufficient for extension length; %d < %d", len(rawPacket), currOffset+extensionLength)
		}

		h.ExtensionPayload = rawPacket[currOffset : currOffset+extensionLength]
		currOffset += len(h.ExtensionPayload)
	}
	h.PayloadOffset = currOffset
	h.HeaderLength = uint32(currOffset)

	return nil
}

func (h *RtpHeader) MarshalTo(buf []byte) (n int, err error) {
	size := h.MarshalSize()
	if size > len(buf) {
		return 0, io.ErrShortBuffer
	}

	// The first byte contains the version, padding bit, extension bit, and csrc size
	h.Version = kRtpVersion
	buf[0] = (h.Version << kRtpVersionShift) | uint8(len(h.CSRC))
	if h.Padding {
		buf[0] |= 1 << kRtpPaddingShift
	}

	if h.Extension {
		buf[0] |= 1 << kRtpExtensionShift
	}

	// The second byte contains the marker bit and payload type.
	buf[1] = h.PayloadType
	if h.Marker {
		buf[1] |= 1 << kRtpMarkerShift
	}

	binary.BigEndian.PutUint16(buf[kRtpSeqNumOffset:kRtpSeqNumOffset+2], h.SequenceNumber)
	binary.BigEndian.PutUint32(buf[kRtpTimestampOffset:kRtpTimestampOffset+4], h.Timestamp)
	binary.BigEndian.PutUint32(buf[kRtpSsrcOffset:kRtpSsrcOffset+4], h.SSRC)

	n = kRtpHeaderLength
	for _, csrc := range h.CSRC {
		binary.BigEndian.PutUint32(buf[n:n+4], csrc)
		n += 4
	}

	// Calculate the size of the header by seeing how many bytes we're written.
	// TODO This is a BUG but fixing it causes more issues.
	h.PayloadOffset = n

	if h.Extension {
		if len(h.ExtensionPayload)%4 != 0 {
			//the payload must be in 32-bit words.
			return 0, io.ErrShortBuffer
		}
		extSize := uint16(len(h.ExtensionPayload) / 4)
		binary.BigEndian.PutUint16(buf[n+0:n+2], h.ExtensionProfile)
		binary.BigEndian.PutUint16(buf[n+2:n+4], extSize)
		n += 4
		n += copy(buf[n:], h.ExtensionPayload)
	}

	return n, nil
}

type RtpPacket struct {
	RtpHeader
	Raw     []byte
	Payload []byte
}

func (p RtpPacket) String() string {
	out := "RTP PACKET:\n"
	out += fmt.Sprintf("\tVersion: %v\n", p.Version)
	out += fmt.Sprintf("\tMarker: %v\n", p.Marker)
	out += fmt.Sprintf("\tPayloadType: %d\n", p.PayloadType)
	out += fmt.Sprintf("\tSequenceNumber: %d\n", p.SequenceNumber)
	out += fmt.Sprintf("\tTimestamp: %d\n", p.Timestamp)
	out += fmt.Sprintf("\tSSRC: %d (0x%x)\n", p.SSRC, p.SSRC)
	out += fmt.Sprintf("\tPayloadLength: %d\n", len(p.Payload))
	out += fmt.Sprintf("\tPacketLength: %d\n", len(p.Raw))
	return out
}

// Unmarshal parses the Raw RTP data
func (p *RtpPacket) Unmarshal(rawPacket []byte) error {
	if err := p.RtpHeader.Unmarshal(rawPacket); err != nil {
		return err
	}

	p.Payload = rawPacket[p.PayloadOffset:]
	p.Raw = rawPacket
	return nil
}

// Marshal serializes the packet into bytes.
func (p *RtpPacket) Marshal() (buf []byte, err error) {
	buf = make([]byte, p.MarshalSize())
	if n, err := p.MarshalTo(buf); err != nil {
		return nil, err
	} else {
		return buf[:n], nil
	}
}

// MarshalTo serializes the packet and writes to the buffer.
func (p *RtpPacket) MarshalTo(buf []byte) (n int, err error) {
	n, err = p.RtpHeader.MarshalTo(buf)
	if err != nil {
		return 0, err
	}

	// Make sure the buffer is large enough to hold the packet.
	if n+len(p.Payload) > len(buf) {
		return 0, io.ErrShortBuffer
	}

	m := copy(buf[n:], p.Payload)
	p.Raw = buf[n : n+m]
	return n + m, nil
}

// MarshalSize returns the size of the packet once marshaled.
func (p *RtpPacket) MarshalSize() int {
	return p.RtpHeader.MarshalSize() + len(p.Payload)
}

func (p *RtpPacket) ParseVideo() *RtpVideo {
	nal_data := C.uint8ptr_t(unsafe.Pointer(&p.Raw[0]))
	nal_size := C.size_t(len(p.Raw))
	ntype := C.uint8_t(0)
	width := C.int(0)
	height := C.int(0)
	spsid := C.uint32_t(0)
	C.parse_rtp_video(nal_data, nal_size, &ntype, &width, &height, &spsid)
	return &RtpVideo{uint8(ntype), int(width), int(height), uint32(spsid)}
}

type RtpVideo struct {
	nalType uint8
	width   int
	height  int
	spsId   uint32
}

/// Some Common RTP Tools

func GetRtpPayloadType(buf []byte) uint8 {
	if len(buf) < kRtpHeaderLength {
		return 0
	}
	return buf[1] & kRtpPtMask
}

func GetRtpSeqNum(buf []byte) uint16 {
	if len(buf) < kRtpHeaderLength {
		return 0
	}
	offset := kRtpSeqNumOffset
	return binary.BigEndian.Uint16(buf[offset : offset+2])
}

func GetRtpTimestamp(buf []byte) uint32 {
	if len(buf) < kRtpHeaderLength {
		return 0
	}
	offset := kRtpTimestampOffset
	return binary.BigEndian.Uint32(buf[offset : offset+4])
}

func GetRtpSsrc(buf []byte) uint32 {
	if len(buf) < kRtpHeaderLength {
		return 0
	}
	offset := kRtpSsrcOffset
	return binary.BigEndian.Uint32(buf[offset : offset+4])
}

func GetRtpHeaderLen(buf []byte) int {
	if len(buf) < kMinRtpPacketLen {
		return 0
	}
	header := buf[0:]
	header_size := kMinRtpPacketLen + (int(header[0]&0xF) * 4)
	if len(buf) < header_size {
		return 0
	}
	if int(header[0]&0x10) != 0 {
		if len(buf) < (header_size + 4) {
			return 0
		}
		length := int(binary.BigEndian.Uint16(header[header_size+2 : header_size+4]))
		header_size += ((length + 1) * 4)
		if len(buf) < header_size {
			return 0
		}
	}
	return header_size
}

func SetRtpHeaderFlags(buf []byte, padding bool, extension bool, csrc_count int) bool {
	if csrc_count > 0x0F {
		return false
	}
	flags := 0
	flags |= (kRtpVersion << 6)
	flags |= (IfInt(padding, 1, 0) << 5)
	flags |= (IfInt(extension, 1, 0) << 4)
	flags |= csrc_count
	buf[0] = uint8(flags)
	return true
}

func SetRtpPayloadType(buf []byte, payloadType uint8) bool {
	if len(buf) < kRtpHeaderLength {
		return false
	}
	marker := (buf[1] >> kRtpMarkerShift & kRtpMarkerMask) > 0
	buf[1] = payloadType
	if marker {
		buf[1] |= 1 << kRtpMarkerShift
	}
	return true
}

func SetRtpSeqNum(buf []byte, seqNum uint16) bool {
	if len(buf) < kRtpHeaderLength {
		return false
	}
	offset := kRtpSeqNumOffset
	binary.BigEndian.PutUint16(buf[offset:offset+2], seqNum)
	return true
}

func SetRtpTimestamp(buf []byte, timestamp uint32) bool {
	if len(buf) < kRtpHeaderLength {
		return false
	}
	offset := kRtpTimestampOffset
	binary.BigEndian.PutUint32(buf[offset:offset+4], timestamp)
	return true
}

func SetRtpSsrc(buf []byte, ssrc uint32) bool {
	if len(buf) < kRtpHeaderLength {
		return false
	}
	offset := kRtpSsrcOffset
	binary.BigEndian.PutUint32(buf[offset:offset+4], ssrc)
	return true
}

// Logic Newer Seq(seq > prevSeq)
func IsNewerRtpSeq(seq, prevSeq uint16) bool {
	diff := uint16(seq - prevSeq)
	return diff != 0 && diff < 0x8000
}

// Logic Newer Timestamp
func IsNewerRtpTimestamp(ts, prevTs uint32) bool {
	diff := uint32(ts - prevTs)
	return diff != 0 && diff < 0x80000000
}

func LatestRtpSeq(seq, prevSeq uint16) uint16 {
	if IsNewerRtpSeq(seq, prevSeq) {
		return seq
	} else {
		return prevSeq
	}
}

func LatestRtpTimestamp(ts, prevTs uint32) uint32 {
	if IsNewerRtpTimestamp(ts, prevTs) {
		return ts
	} else {
		return prevTs
	}
}

// This returns 1 if RTP-SEQ(uint16) seq1 > seq2.
func CompareRtpSeq(seq1, seq2 uint16) int {
	diff := seq1 - seq2
	if diff != 0 {
		if diff < 0x8000 {
			return 1
		} else {
			return -1
		}
	} else {
		return 0
	}
}

func ComputeRtpSeqDistance(seq1, seq2 uint16) int {
	diff := uint16(seq1 - seq2)
	if diff <= 0x8000 {
		return int(diff)
	} else {
		return 65536 - int(diff)
	}
}

// This return true if RTP-SEQ(uint16) seqn between (start, start+size).
func IsRtpSeqInRange(seqn, start, size uint16) bool {
	var n int = int(seqn)
	var nh int = ((1 << 16) + n)
	var s int = int(start)
	var e int = s + int(size)
	return (s <= n && n < e) || (s <= nh && nh < e)
}
