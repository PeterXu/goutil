package util

import (
	//"crypto/rsa"
	"crypto/sha256"
	//"crypto/tls"
	"crypto/x509"
	//"encoding/hex"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

const kSdpOwner string = "xrtc"
const kSdpCname string = "xrtc_endpoint"
const kAudioStreamLabel string = "stream_audio_label"
const kAudioTrackLabel string = "track_audio_label"

var kNewlineChar = []byte{'\n'}
var kSpaceChar = []byte{' '}

// MediaType m=audio/video/application
type MediaType int

// These are different media types
const (
	kMediaNil   MediaType = 0
	kMediaAudio MediaType = 1 << (iota - 1)
	kMediaVideo
	kMediaApplication
)

type SdpMediaAttrs struct {
	Ssrcs   map[uint32]*SdpSsrc
	Ptypes  map[uint8]*SdpPtype
	Extmaps map[int]*SdpExtmap
}

func NewSdpMediaAttrs() *SdpMediaAttrs {
	return &SdpMediaAttrs{
		Ssrcs:   make(map[uint32]*SdpSsrc),
		Ptypes:  make(map[uint8]*SdpPtype),
		Extmaps: make(map[int]*SdpExtmap),
	}
}

// RtpMap
type SdpPtype struct {
	Ptype     uint8
	AptPtype  uint8
	Codec     string
	Channels  int
	Frequency int
}

func (sp SdpPtype) String() string {
	return fmt.Sprintf("ptype:%d_%d_%s_%dch_%dhz",
		sp.Ptype, sp.AptPtype, sp.Codec, sp.Channels, sp.Frequency)
}

// SSRC pair: main/rtx
// a=ssrc
// a=fmtp:. apt=rtx
type SdpSsrc struct {
	Main uint32
	Rtx  uint32
	Num  int
	Idx  int
}

func (ss SdpSsrc) String() string {
	return fmt.Sprintf("ssrc:%d_%d_idx%d", ss.Main, ss.Rtx, ss.Idx)
}

type SdpExtmap struct {
	Id  int
	Uri string
}

func (se SdpExtmap) String() string {
	return fmt.Sprintf("extmap:%d_%s", se.Id, se.Uri)
}

// Media Direction
type SdpMediaDirection int

// These are different sdp directions: sendrecv/sendonly/recvonly/inactive
const (
	kDirectionUnknown  SdpMediaDirection = 0
	kDirectionInactive SdpMediaDirection = iota
	kDirectionSendOnly
	kDirectionRecvOnly
	kDirectionSendRecv
)

// SDP a=msid-semantic
// a=msid-semantic:WMS *
// a=msid-semantic:WMS id1
type MsidSemantic struct {
	name  string
	msids []string
}

// SDP a=rtpmap
// a=rtpmap:111 opus/48000/2
// a=rtpmap:126 H264/90000
type RtpMapInfo struct {
	ptype     int
	codec     string
	frequency int
	channels  int
	misc      string
	apt_ptype int
}

func (r *RtpMapInfo) Clone() *RtpMapInfo {
	d := &RtpMapInfo{}
	d.ptype = r.ptype
	d.codec = r.codec
	d.frequency = r.frequency
	d.channels = r.channels
	d.misc = r.misc
	d.apt_ptype = r.apt_ptype
	return d
}

func (r RtpMapInfo) a_rtpmap() string {
	return Itoa(r.ptype) + " " + r.codec + "/" + Itoa(r.frequency)
}

func (r RtpMapInfo) a_rtxmap() string {
	return Itoa(r.apt_ptype) + " rtx/90000"
}

func (r RtpMapInfo) a_fmtp_apt() string {
	return Itoa(r.apt_ptype) + " apt=" + Itoa(r.ptype)
}

// NewFmtpInfo return a FmtpInfo object
// a=fmtp:111 maxplaybackrate=48000;stereo=1;useinbandfec=1
// a=fmtp:126 profile-level-id=42e01f;level-asymmetry-allowed=1;packetization-mode=1
// a=fmtp:101 0-15
func NewFmtpInfo(ptype int) *FmtpInfo {
	return &FmtpInfo{ptype: ptype, props: make(map[string]int)}
}

// SDP media format-specific parameters: a=fmtp
type FmtpInfo struct {
	ptype int
	props map[string]int
	misc  string
}

// SDP rtcp feedback: a=rtcp-fb
// a=rtcp-fb:126 nack
// a=rtcp-fb:126 nack pli
type RtcpFbInfo struct {
	ptype  int
	fbtype string
}

// SDP rtp-ext-header: a=extmap
// a=extmap:1 http://www.webrtc.org/experiments/rtp-hdrext/abs-send-time
// a=extmap:2/sendrecv urn:ietf:params:rtp-hdrext:toffset
type ExtMapInfo struct {
	id        int
	direction string
	uri       string
}

// SDP ssrc: a=ssrc
// a=ssrc:1081040086 cname:{name}
// a=ssrc:1081040086 cname:name
// a=ssrc:1081040086 msid:id1 id2
// a=ssrc:1081040086 mslabel:id1
// a=ssrc:1081040086 label:id2
type SsrcInfo struct {
	ssrc    uint32
	cname   string
	msids   []string
	label   string
	mslabel string
}

// SDP a=ssrc-group:FID
// a=ssrc-group:FID 1081040086 1081040087
type FidInfo struct {
	main uint32
	rtx  uint32
}

// SDP sctp: a=sctpmap
// a=sctpmap:5000 webrtc-datachannel 1024
type SctpInfo struct {
	port       int
	name       string
	number     int
	is_sctpmap bool
}

func NewMediaAttr(mtype, proto string) *MediaAttr {
	return &MediaAttr{mtype: mtype, proto: proto,
		fmtps:      make(map[int]*FmtpInfo),
		av_rtpmaps: make(map[string]*RtpMapInfo)}
}

// SDP media attribute lines
type MediaAttr struct {
	mtype            string            // m=
	proto            string            // m=
	ptypes           []string          // m=
	ice_ufrag        string            // a=ice-ufrag:..
	ice_pwd          string            // a=ice-pwd:..
	ice_options      string            // a=ice-options:..
	fingerprint      StringPair        // a=fingerprint:sha-256 ..
	setup            string            // a=setup:..
	direction        SdpMediaDirection // a=sendrecv/sendonly/recvonly
	mid              string            // a=mid:..
	msid             []*StringPair     // a=msid:{id1} {id2}
	rtcp_mux         bool              // a=rtcp-mux
	rtcp_rsize       bool              // a=rtcp-rsize
	rtpmaps          []*RtpMapInfo     // a=rtpmap:..
	fmtps            map[int]*FmtpInfo // a=fmtp:..
	rtcp_fbs         []*RtcpFbInfo     // a=rtcp-fb:..
	extmaps          []*ExtMapInfo     // a=extmap:..
	fid_ssrcs        []*FidInfo        // a=ssrc-group:FID ..
	ssrcs            []*SsrcInfo       // a=ssrc:..
	msids            []string          // a=msid:..
	sctp             *SctpInfo         // a=sctpmap: or a=sctp-port:
	max_message_size int               // a=max-message-size:
	candidates       []string          // a=candidate:
	maxptime         int

	// for anwser
	av_rtpmaps      map[string]*RtpMapInfo
	use_rtx         bool
	use_rtx_apt     bool
	use_rtx_fid     bool
	use_red_fec     bool
	use_red_rtx     bool
	use_red_rtx_apt bool
}

func (a *MediaAttr) GetSsrcs() *SdpSsrc {
	ssrc := &SdpSsrc{}
	if len(a.fid_ssrcs) > 0 {
		ssrc.Main = a.fid_ssrcs[0].main
		ssrc.Rtx = a.fid_ssrcs[0].rtx
		ssrc.Num = len(a.fid_ssrcs)
	} else if len(a.ssrcs) > 0 {
		ssrc.Main = a.ssrcs[0].ssrc
		ssrc.Num = len(a.ssrcs)
	} else {
		ssrc = nil
	}
	return ssrc
}

func (a *MediaAttr) GetExtmaps(attrs *SdpMediaAttrs) {
	for _, item := range a.extmaps {
		attrs.Extmaps[item.id] = &SdpExtmap{item.id, item.uri}
	}
}

func (a *MediaAttr) GetSsrc(attrs *SdpMediaAttrs) {
	if len(a.fid_ssrcs) > 0 {
		num := len(a.fid_ssrcs)
		idx := num - 1
		for _, item := range a.fid_ssrcs {
			attrs.Ssrcs[item.main] = &SdpSsrc{item.main, item.rtx, num, idx}
			idx -= 1
		}
	} else if len(a.ssrcs) > 0 {
		num := len(a.ssrcs)
		idx := num - 1
		for _, item := range a.ssrcs {
			attrs.Ssrcs[item.ssrc] = &SdpSsrc{item.ssrc, 0, num, idx}
			idx -= 1
		}
	}
}

func (a *MediaAttr) GetPtype(attrs *SdpMediaAttrs) {
	if len(a.rtpmaps) > 0 {
		for _, item := range a.rtpmaps {
			sdpPtype := &SdpPtype{
				Ptype:     uint8(item.ptype),
				AptPtype:  uint8(item.apt_ptype),
				Codec:     item.codec,
				Channels:  item.channels,
				Frequency: item.frequency,
			}

			// Check a=fmtp:rtx_ptype apt=main_ptype
			if item.codec == "rtx" {
				if fmtp, ok := a.fmtps[item.ptype]; ok {
					if main_ptype, had := fmtp.props["apt"]; had {
						sdpPtype.AptPtype = uint8(main_ptype) // main type
					}
				}
			} else {
				for apt_ptype, fmtp := range a.fmtps {
					if main_ptype, had := fmtp.props["apt"]; had {
						if item.ptype == main_ptype {
							sdpPtype.AptPtype = uint8(apt_ptype) // rtx type
							break
						}
					}
				}
			}

			attrs.Ptypes[uint8(item.ptype)] = sdpPtype
		}
	}
}

// SDP media lines
type MediaSdp struct {
	owner         string       // o=..
	source        string       // s=..
	ice_lite      bool         // a=ice-lite
	ice_options   string       // global a=ice-options:..
	fingerprint   StringPair   // global a=fingerprint:sha-256 ..
	group_bundles []string     // a=group:BUNDLE ..
	msid_semantic MsidSemantic // a=msid-sematic: ..
	audios        []*MediaAttr // m=audio ..
	videos        []*MediaAttr // m=video ..
	applications  []*MediaAttr // m=application ..
}

// parseSdp to parse SDP lines, return true if ok
func (m *MediaSdp) parseSdp(data []byte) bool {
	var mattr *MediaAttr
	lines := strings.Split(string(data), "\r\n")
	if len(lines) <= 1 {
		lines = strings.Split(string(data), "\n")
	}

	//log.Println("[sdp] parseSdp, lines=", len(lines))
	for item := range lines {
		line := []byte(lines[item])
		if len(line) <= 2 || line[1] != '=' {
			//log.Println("invalid sdp line: ", string(line))
			continue
		}

		switch line[0] {
		case 'v':
			// nop
		case 'o':
			fields := strings.SplitN(string(line[2:]), " ", 2)
			if len(fields) == 2 {
				attrs := strings.SplitN(fields[1], " ", 2)
				m.owner = attrs[0]
			}
		case 's':
			// nop
		case 't':
			// nop
		case 'm':
			fields := strings.Split(string(line[2:]), " ")
			if len(fields) >= 4 {
				mattr = NewMediaAttr(fields[0], fields[2])
				mattr.ptypes = append(mattr.ptypes, fields[3:]...)
			} else {
				mattr = NewMediaAttr(fields[0], "")
			}
			if fields[0] == "audio" {
				m.audios = append(m.audios, mattr)
			} else if fields[0] == "video" {
				m.videos = append(m.videos, mattr)
			} else if fields[0] == "application" {
				m.applications = append(m.applications, mattr)
			} else {
				break
			}
		case 'c':
			// nop
		case 'a':
			m.parseSdp_a(line, mattr)
		default:
		}
	}
	return true
}

// parseSdp_a to parse SDP attribute: 'a='
func (m *MediaSdp) parseSdp_a(line []byte, media *MediaAttr) {
	fields := strings.SplitN(string(line[2:]), ":", 2)
	akey := fields[0]
	if len(fields) == 1 {
		if akey == "ice-lite" {
			m.ice_lite = true
			return
		}

		if media == nil {
			log.Println("[sdp] no valid media for line=", string(line[:]))
			return
		}

		if akey == "inactive" {
			media.direction = kDirectionInactive
		} else if akey == "sendonly" {
			media.direction = kDirectionSendOnly
		} else if akey == "recvonly" {
			media.direction = kDirectionRecvOnly
		} else if akey == "sendrecv" {
			media.direction = kDirectionSendRecv
		} else if akey == "rtcp-mux" {
			media.rtcp_mux = true
		} else if akey == "rtcp-rsize" {
			media.rtcp_rsize = true
		}
		return
	}

	if akey == "group" {
		attrs := strings.Split(fields[1], " ")
		//log.Println("[sdp] a=group:", attrs, len(attrs))
		if len(attrs) >= 1 {
			aval := strings.ToLower(attrs[0])
			switch aval {
			case "bundle":
				if len(attrs) >= 2 {
					m.group_bundles = append(m.group_bundles, attrs[1:]...)
				}
			default:
				log.Println("[sdp] unsupported attr - a=group:", aval)
			}
		}
		return
	}

	if akey == "msid-semantic" {
		attrs := strings.Split(fields[1], " ")
		if len(attrs) >= 1 {
			m.msid_semantic.name = attrs[0]
		}
		if len(attrs) >= 2 {
			props := attrs[1:]
			m.msid_semantic.msids = append(m.msid_semantic.msids, props...)
		}
		return
	}

	if media == nil {
		if akey == "ice-options" {
			m.ice_options = fields[1]
			return
		} else if akey == "fingerprint" {
			attrs := strings.SplitN(fields[1], " ", 2)
			if len(attrs) == 2 {
				m.fingerprint.First = attrs[0]
				m.fingerprint.Second = attrs[1]
			}
			return
		}

		log.Println("[sdp] no valid media for line=", string(line[:]))
		return
	}

	if akey == "rtcp" {
		// nop
	} else if akey == "ice-ufrag" {
		media.ice_ufrag = strings.TrimSpace(fields[1])
	} else if akey == "ice-pwd" {
		media.ice_pwd = strings.TrimSpace(fields[1])
	} else if akey == "ice-options" {
		media.ice_options = fields[1]
	} else if akey == "fingerprint" {
		attrs := strings.SplitN(fields[1], " ", 2)
		if len(attrs) == 2 {
			media.fingerprint.First = attrs[0]
			media.fingerprint.Second = attrs[1]
		}
	} else if akey == "setup" {
		media.setup = fields[1]
	} else if akey == "mid" {
		media.mid = fields[1]
	} else if akey == "rtpmap" {
		attrs := strings.SplitN(fields[1], " ", 2)
		if len(attrs) == 2 {
			rmap := &RtpMapInfo{ptype: Atoi(attrs[0])}
			props := strings.Split(attrs[1], "/")
			if len(props) >= 2 {
				rmap.codec = strings.ToLower(props[0])
				rmap.frequency = Atoi(props[1])
				if len(props) >= 3 {
					rmap.channels = Atoi(props[2])
				}
			} else {
				rmap.misc = attrs[1]
			}
			media.rtpmaps = append(media.rtpmaps, rmap)
		}
	} else if akey == "fmtp" {
		attrs := strings.SplitN(fields[1], " ", 2)
		if len(attrs) == 2 {
			fmtp := NewFmtpInfo(Atoi(attrs[0]))
			props := strings.Split(attrs[1], ";")
			for k := range props {
				kv := strings.Split(props[k], "=")
				if len(kv) == 2 {
					fmtp.props[kv[0]] = Atoi(kv[1])
				} else {
					fmtp.misc = props[k]
				}
			}
			media.fmtps[fmtp.ptype] = fmtp
		}
	} else if akey == "rtcp-fb" {
		attrs := strings.SplitN(fields[1], " ", 2)
		if len(attrs) == 2 {
			rtcpfb := &RtcpFbInfo{Atoi(attrs[0]), attrs[1]}
			media.rtcp_fbs = append(media.rtcp_fbs, rtcpfb)
		}
	} else if akey == "extmap" {
		attrs := strings.SplitN(fields[1], " ", 2)
		if len(attrs) == 2 {
			extmap := &ExtMapInfo{id: Atoi(attrs[0]), uri: attrs[1]}
			keys := strings.Split(attrs[0], "/")
			if len(keys) >= 2 {
				extmap.direction = keys[1]
			}
			media.extmaps = append(media.extmaps, extmap)
		}
	} else if akey == "ssrc-group" {
		attrs := strings.SplitN(fields[1], " ", 2)
		if len(attrs) == 2 {
			if attrs[0] == "FID" {
				props := strings.Split(attrs[1], " ")
				if len(props) == 2 {
					fid := &FidInfo{Atou32(props[0]), Atou32(props[1])}
					media.fid_ssrcs = append(media.fid_ssrcs, fid)
				}
			} else if attrs[0] == "SIM" {
				// not support
			}
		}
	} else if akey == "ssrc" {
		attrs := strings.SplitN(fields[1], " ", 2)
		if len(attrs) == 2 {
			ssrc := &SsrcInfo{ssrc: Atou32(attrs[0])}
			props := strings.SplitN(attrs[1], ":", 2)
			if len(props) == 2 {
				if props[0] == "cname" {
					ssrc.cname = strings.Trim(props[0], "{}")
				} else if props[0] == "msid" {
					msids := strings.Split(props[1], " ")
					ssrc.msids = append(ssrc.msids, msids...)
				} else if props[0] == "mslabel" {
					ssrc.mslabel = props[1]
				} else if props[0] == "label" {
					ssrc.label = props[1]
				}
			}
			media.ssrcs = append(media.ssrcs, ssrc)
		}
	} else if akey == "msid" {
		msids := strings.Split(fields[1], " ")
		if len(msids) > 0 {
			media.msids = append(media.msids, msids...)
		}
	} else if akey == "sctpmap" {
		attrs := strings.Split(fields[1], " ")
		if len(attrs) >= 3 {
			media.sctp = &SctpInfo{Atoi(attrs[0]), attrs[1], Atoi(attrs[2]), true}

		}
	} else if akey == "sctp-port" {
		media.sctp = &SctpInfo{port: Atoi(fields[1])}
	} else if akey == "max-message-size" {
		media.max_message_size = Atoi(fields[1])
	} else if akey == "candidate" {
		media.candidates = append(media.candidates, string(line))
	} else if akey == "maxptime" {
		media.maxptime = Atoi(fields[1])
	} else {
		log.Println("[sdp] unsupported attr=", akey)
	}
}

// Media description (sdp offer/answer)
type MediaDesc struct {
	Sdp        MediaSdp
	haveAnswer bool

	// sdp answer
	av_agent        string
	av_unified_plan bool
	av_ice_ufrag    string
	av_ice_pwd      string
	av_fingerprint  StringPair // answer a=fingerprint:sha-256 ..
	av_ssrcs        map[uint32]uint32
}

func (m *MediaDesc) Parse(data []byte) bool {
	return m.Sdp.parseSdp(data)
}

func (m *MediaDesc) HaveAudio() bool {
	mt := m.GetMediaType()
	return (mt & kMediaAudio) != 0
}

func (m *MediaDesc) HaveVideo() bool {
	mt := m.GetMediaType()
	return (mt & kMediaVideo) != 0
}

func (m *MediaDesc) GetAudioAttrs() *SdpMediaAttrs {
	attrs := NewSdpMediaAttrs()
	for _, media := range m.Sdp.audios {
		media.GetSsrc(attrs)
		media.GetPtype(attrs)
		media.GetExtmaps(attrs)
	}
	return attrs
}

func (m *MediaDesc) GetVideoAttrs() *SdpMediaAttrs {
	attrs := NewSdpMediaAttrs()
	for _, media := range m.Sdp.videos {
		media.GetSsrc(attrs)
		media.GetPtype(attrs)
		media.GetExtmaps(attrs)
	}
	return attrs
}

func (m *MediaDesc) GetMediaType() MediaType {
	mt := kMediaNil
	if len(m.Sdp.audios) > 0 {
		mt |= kMediaAudio
	}
	if len(m.Sdp.videos) > 0 {
		mt |= kMediaVideo
	}
	if len(m.Sdp.applications) > 0 {
		mt |= kMediaApplication
	}
	return mt
}

func (m *MediaDesc) GetUfrag() string {
	mt := m.GetMediaType()
	if (mt & kMediaAudio) != 0 {
		return m.Sdp.audios[0].ice_ufrag
	} else if (mt & kMediaVideo) != 0 {
		return m.Sdp.videos[0].ice_ufrag
	} else if (mt & kMediaApplication) != 0 {
		return m.Sdp.applications[0].ice_ufrag
	} else {
		log.Println("[desc] invalid media type = ", mt)
		return ""
	}
}

func (m *MediaDesc) GetPasswd() string {
	mt := m.GetMediaType()
	if (mt & kMediaAudio) != 0 {
		return m.Sdp.audios[0].ice_pwd
	} else if (mt & kMediaVideo) != 0 {
		return m.Sdp.videos[0].ice_pwd
	} else if (mt & kMediaApplication) != 0 {
		return m.Sdp.applications[0].ice_pwd
	} else {
		log.Println("[desc] invalid media type = ", mt)
		return ""
	}
}

func (m *MediaDesc) GetCandidates() []string {
	mt := m.GetMediaType()
	if (mt & kMediaAudio) != 0 {
		return m.Sdp.audios[0].candidates
	} else if (mt & kMediaVideo) != 0 {
		return m.Sdp.videos[0].candidates
	} else if (mt & kMediaApplication) != 0 {
		return m.Sdp.applications[0].candidates
	} else {
		log.Println("[desc] invalid media type = ", mt)
		return nil
	}
}

func (m *MediaDesc) CreateAnswer(agent string, certFile string) bool {
	switch agent {
	case FirefoxAgent:
		m.av_unified_plan = true
	case ChromeAgent:
		m.av_unified_plan = false
	default:
		return false
	}
	m.av_agent = agent
	m.av_ssrcs = make(map[uint32]uint32)

	// create ufrag/pwd
	m.av_ice_ufrag = "xrtc" + RandomString(12)
	m.av_ice_pwd = RandomString(24)

	// create fingerprint
	if cert, err := LoadX509Certificate(certFile); err == nil {
		sum := sha256.Sum256(cert.Raw)
		if len(sum) != sha256.Size {
			log.Println("[sdp] fail to sha256.Sum256")
			return false
		}
		prefix := ""
		fingerprint := ""
		for k := 0; k < sha256.Size; k++ {
			fingerprint += fmt.Sprintf("%s%02X", prefix, sum[k])
			prefix = ":"
		}
		// 32*2+31 = 95
		log.Println("[sdp] fingerprint=", len(fingerprint), fingerprint)
		m.av_fingerprint.First = "sha-256"
		m.av_fingerprint.Second = fingerprint
	} else {
		log.Println("[sdp] fail to load x509:", err)
		return false
	}

	// pre-select each m=audio
	log.Println("[sdp] check audios:", len(m.Sdp.audios))
	if len(m.Sdp.audios) > 0 {
		// priority: opus > pcmu > pcma
		codecs := []string{"opus", "pcmu", "pcma"}
		for i := range m.Sdp.audios {
			audio := m.Sdp.audios[i]
			for j := range audio.rtpmaps {
				rtpmap := audio.rtpmaps[j]
				if rtpmap.codec == "opus" && rtpmap.frequency == 48000 {
					audio.av_rtpmaps["opus"] = rtpmap.Clone()
				}
				if rtpmap.codec == "pcmu" {
					audio.av_rtpmaps["pcmu"] = rtpmap.Clone()
				}
				if rtpmap.codec == "pcma" {
					audio.av_rtpmaps["pcma"] = rtpmap.Clone()
				}
			}
			for _, codec := range codecs {
				if rtpmap, ok := audio.av_rtpmaps[codec]; ok {
					audio.av_rtpmaps["main"] = rtpmap.Clone()
					break
				}
			}
		}
	}

	// pre-select in each m=video
	log.Println("[sdp] check videos:", len(m.Sdp.videos))
	if len(m.Sdp.videos) > 0 {
		//codecs := []string{"h264-0", "h264-1", "h264-2"}

		for i := range m.Sdp.videos {
			have_h264 := false
			have_red := false
			have_fec := false

			have_rtx := false
			have_rtx_apt := false

			video := m.Sdp.videos[i]

			// select the first h264
			for j := range video.rtpmaps {
				rtpmap := video.rtpmaps[j]
				//log.Println("[sdp] check codec:", rtpmap.codec)
				switch rtpmap.codec {
				case "h264":
					if !have_h264 {
						have_h264 = true
						video.av_rtpmaps["main"] = rtpmap.Clone()
					}
				case "red":
					if !have_red {
						have_red = true
						video.av_rtpmaps[rtpmap.codec] = rtpmap.Clone()
					}
				case "ulpfec":
					if !have_fec {
						have_fec = true
						video.av_rtpmaps[rtpmap.codec] = rtpmap.Clone()
					}
				}
			}
			main, _ := video.av_rtpmaps["main"]
			if main == nil {
				continue
			}
			//log.Println("[sdp] check main rtpmap:", main)

			for j := range video.rtpmaps {
				rtpmap := video.rtpmaps[j]
				if rtpmap.codec == "rtx" {
					if fmtp, ok := video.fmtps[rtpmap.ptype]; ok {
						//log.Println("[sdp] check rtpmap:", rtpmap, fmtp)
						if ptype, _ := fmtp.props["apt"]; ptype == main.ptype {
							have_rtx = true
							have_rtx_apt = true
							main.apt_ptype = rtpmap.ptype
							video.av_rtpmaps[rtpmap.codec] = rtpmap.Clone()
							break
						}
					}
				}
			}

			have_rtx_fid := (len(video.fid_ssrcs) > 0)
			if have_h264 {
				// hardcode to select supported features
				video.use_rtx = have_rtx
				video.use_rtx_apt = have_rtx_apt
				video.use_rtx_fid = have_rtx_fid
				video.use_red_fec = false
				video.use_red_rtx = false
				video.use_red_rtx_apt = false
			}
		}
	}

	m.haveAnswer = true
	return true
}

func (m *MediaDesc) ParseDrection(direction SdpMediaDirection) string {
	switch direction {
	case kDirectionSendRecv:
		return "sendrecv"
	case kDirectionRecvOnly:
		return "sendonly"
	case kDirectionSendOnly:
		return "recvonly"
	case kDirectionInactive:
		return "inactive"
	}
	return ""
}

func (m *MediaDesc) GetAudioCodec() string {
	if m.haveAnswer {
		for j := range m.Sdp.audios {
			audio := m.Sdp.audios[j]
			if rtpmap, ok := audio.av_rtpmaps["main"]; ok {
				return rtpmap.codec
			}
		}
	}
	return ""
}

func (m *MediaDesc) GetVideoCodec() string {
	if m.haveAnswer {
		for j := range m.Sdp.videos {
			video := m.Sdp.videos[j]
			if rtpmap, ok := video.av_rtpmaps["main"]; ok {
				return rtpmap.codec
			}
		}
	}
	return ""
}

func (m *MediaDesc) AnswerSdp() string {
	var prefix []string
	prefix = append(prefix, "v=0")
	prefix = append(prefix, "o="+kSdpOwner+" 123456789 2 IN IP4 127.0.0.1")
	prefix = append(prefix, "s=-")
	prefix = append(prefix, "t=0 0")

	bundles := "a=group:BUNDLE"
	semantics := "a=msid-semantic:WMS"
	log.Println("[desc] all bundles: ", m.Sdp.group_bundles)

	var body []string
	for i := range m.Sdp.group_bundles {
		bundle := m.Sdp.group_bundles[i]
		bundles += " " + bundle
		log.Println("[desc] one media bundle=", bundle)

		// check m=audio
		for j := range m.Sdp.audios {
			audio := m.Sdp.audios[j]
			if audio.mid == bundle {
				rtpmap, _ := audio.av_rtpmaps["main"]
				mline := "m=audio 1 " + audio.proto
				if rtpmap != nil {
					mline += " " + Itoa(rtpmap.ptype)
				}
				mline += " 126" // add telephone-event payload-type
				body = append(body, mline)
				body = append(body, "c=IN IP4 0.0.0.0")
				//body = append(body, "a=rtcp:1 IN IP4 0.0.0.0")
				body = append(body, "a=ice-ufrag:"+m.av_ice_ufrag)
				body = append(body, "a=ice-pwd:"+m.av_ice_pwd)
				body = append(body, "a=fingerprint:"+m.av_fingerprint.ToString(" "))
				body = append(body, "a=setup:passive")
				// set a=extmap:
				for k := range audio.extmaps {
					extmap := audio.extmaps[k]
					if strings.Contains(extmap.uri, "ssrc-audio-level") {
						aextmap := "a=extmap:" + Itoa(extmap.id) + " " + extmap.uri
						body = append(body, aextmap)
					}
				}
				// set audio a=sendrecv
				if rtpmap == nil {
					body = append(body, "a=inactive")
				} else {
					if adir := m.ParseDrection(audio.direction); len(adir) > 0 {
						body = append(body, "a="+adir)
					}
				}
				body = append(body, "a=mid:"+bundle)
				body = append(body, "a=rtcp-mux")
				if rtpmap != nil {
					artpmap := "a=rtpmap:" + rtpmap.a_rtpmap()
					if rtpmap.channels > 0 {
						artpmap += "/" + Itoa(rtpmap.channels)
					}
					body = append(body, artpmap)

					if rtpmap.codec == "opus" {
						afmtp := "a=fmtp:" + Itoa(rtpmap.ptype) + " minptime=20;useinbandfec=1;usedtx=0"
						body = append(body, afmtp)
						// ???: always use 20
						//body = append(body, "a=maxptime:"+Itoa(audio.maxptime))
						body = append(body, "a=maxptime:20")
					}
				}
				body = append(body, "a=rtpmap:126 telephone-event/8000")

				semantics += " " + kAudioStreamLabel
				body = append(body, "a=ssrc:1 cname:"+kSdpCname)
				if !m.av_unified_plan {
					body = append(body, "a=ssrc:1 msid:"+kAudioStreamLabel+" "+kAudioTrackLabel)
					body = append(body, "a=ssrc:1 mslabel:"+kAudioStreamLabel)
					body = append(body, "a=ssrc:1 label:"+kAudioTrackLabel)
				} else {
					body = append(body, "a=msid:"+kAudioStreamLabel+" "+kAudioTrackLabel)
				}
			}
		}

		// check m=application
		for j := range m.Sdp.applications {
			app := m.Sdp.applications[j]
			if app.mid == bundle {
				body = append(body, "m=application 9 "+app.proto+" "+app.ptypes[0])
				body = append(body, "c=IN IP4 0.0.0.0")
				body = append(body, "a=ice-ufrag:"+m.av_ice_ufrag)
				body = append(body, "a=ice-pwd:"+m.av_ice_pwd)
				body = append(body, "a=fingerprint:"+m.av_fingerprint.ToString(" "))
				body = append(body, "a=setup:passive")

				if adir := m.ParseDrection(app.direction); len(adir) > 0 {
					body = append(body, "a="+adir)
				}
				body = append(body, "a=mid:"+bundle)
				if app.sctp.is_sctpmap {
					body = append(body, "a=sctpmap:"+Itoa(app.sctp.port)+" "+app.sctp.name+" "+Itoa(app.sctp.number))
				} else {
					body = append(body, "a=sctp-port:"+Itoa(app.sctp.port))
				}
			}
		}

		// check m=video
		for j := range m.Sdp.videos {
			video := m.Sdp.videos[j]
			if video.mid == bundle {
				rtpmap, _ := video.av_rtpmaps["main"]
				redmap, _ := video.av_rtpmaps["red"]
				fecmap, _ := video.av_rtpmaps["ulpfec"]

				mline := "m=video 1 " + video.proto
				if rtpmap != nil {
					mline += " " + Itoa(rtpmap.ptype)
					if video.use_rtx && video.use_rtx_apt {
						mline += " " + Itoa(rtpmap.apt_ptype)
					}
				}
				if video.use_red_fec {
					if redmap != nil && fecmap != nil {
						mline += " " + Itoa(redmap.ptype)
						if video.use_red_rtx && video.use_red_rtx_apt {
							mline += " " + Itoa(redmap.apt_ptype)
						}
						mline += " " + Itoa(fecmap.ptype)
					}
				}
				body = append(body, mline)

				body = append(body, "c=IN IP4 0.0.0.0")
				//body = append(body, "b=AS:1500") // refine
				body = append(body, "a=ice-ufrag:"+m.av_ice_ufrag)
				body = append(body, "a=ice-pwd:"+m.av_ice_pwd)
				body = append(body, "a=fingerprint:"+m.av_fingerprint.ToString(" "))
				body = append(body, "a=setup:passive")

				// set a=extmap:
				for k := range video.extmaps {
					var aextmap string
					extmap := video.extmaps[k]
					if strings.Contains(extmap.uri, "urn:ietf:params:rtp-hdrext:toffset") {
						aextmap = "a=extmap:" + Itoa(extmap.id) + " " + extmap.uri
					} else if strings.Contains(extmap.uri, "rtp-hdrext/abs-send-time") {
						aextmap = "a=extmap:" + Itoa(extmap.id) + " " + extmap.uri
					} else if strings.Contains(extmap.uri, "holmer-rmcat-transport-wide-cc-extensions") {
						aextmap = "a=extmap:" + Itoa(extmap.id) + " " + extmap.uri
					} else if strings.Contains(extmap.uri, "ietf-avtext-framemarking") {
						aextmap = "a=extmap:" + Itoa(extmap.id) + " " + extmap.uri
					}
					if len(aextmap) > 0 {
						body = append(body, aextmap)
					}
				}

				if rtpmap == nil {
					body = append(body, "a=inactive")
				} else {
					if adir := m.ParseDrection(video.direction); len(adir) > 0 {
						body = append(body, "a="+adir)
					}
				}
				body = append(body, "a=mid:"+bundle)
				body = append(body, "a=rtcp-mux")
				if rtpmap != nil {
					body = append(body, "a=rtpmap:"+rtpmap.a_rtpmap())
					if video.use_rtx {
						body = append(body, "a=rtcp-fb:"+Itoa(rtpmap.ptype)+" nack")
					}
					body = append(body, "a=rtcp-fb:"+Itoa(rtpmap.ptype)+" nack pli")
					//body = append(body, "a=rtcp-fb:"+Itoa(rtpmap.ptype)+" ccm fir")
					body = append(body, "a=rtcp-fb:"+Itoa(rtpmap.ptype)+" goog-remb")
					body = append(body, "a=fmtp:"+Itoa(rtpmap.ptype)+" level-asymmetry-allowed=1;packetization-mode=1;profile-level-id=42e01f")
					// rtx payload: rtx-rtp for chrome and raw-rtp for firefox
					if video.use_rtx && video.use_rtx_apt {
						body = append(body, "a=rtpmap:"+rtpmap.a_rtxmap())
						body = append(body, "a=fmtp:"+rtpmap.a_fmtp_apt())
					}
				}
				if redmap != nil && fecmap != nil {
					if video.use_red_fec {
						body = append(body, "a=rtpmap:"+redmap.a_rtpmap())
						body = append(body, "a=rtpmap:"+fecmap.a_rtpmap())
						if video.use_red_rtx && video.use_red_rtx_apt {
							body = append(body, "a=rtpmap:"+redmap.a_rtxmap())
							body = append(body, "a=fmtp:"+redmap.a_fmtp_apt())
						}
					}
				}

				// ssrc template which will be processed in client
				//   keyword: <main_ssrc>, <rtx_ssrc>
				var main_ssrc, rtx_ssrc uint32
				if main_ssrc > 0 && rtx_ssrc > 0 {
					if video.use_rtx_fid {
						body = append(body, "a=ssrc-group:FID main_ssrc rtx_ssrc")
					}
					body = append(body, "a=ssrc:main_ssrc cname:"+kSdpCname)
					if !m.av_unified_plan {
						body = append(body, "a=ssrc:main_ssrc msid:stream_video_label track_video_label")
						body = append(body, "a=ssrc:main_ssrc mslabel:stream_video_label")
						body = append(body, "a=ssrc:main_ssrc label:track_video_label")
					} else {
						body = append(body, "a=msid:{stream_video_label} {track_video_label}")
					}
					if video.use_rtx_fid {
						body = append(body, "a=ssrc:rtx_ssrc cname:"+kSdpCname)
						if !m.av_unified_plan {
							body = append(body, "a=ssrc:rtx_ssrc msid:stream_video_label track_video_label")
							body = append(body, "a=ssrc:rtx_ssrc mslabel:stream_video_label")
							body = append(body, "a=ssrc:rtx_ssrc label:track_video_label")
						}
					}
					semantics += " stream_video_label"
				}
			}
		}
	}

	prefix = append(prefix, bundles, semantics)
	sdp := append(prefix, body...)
	return strings.Join(sdp, "\r\n")
}

// UpdateSdpCandidates to replace sdp candidates with new.
func UpdateSdpCandidates(data []byte, candidates []string) []byte {
	if len(candidates) == 0 {
		return data
	}

	sp := "\r\n"
	lines := strings.Split(string(data), "\r\n")
	if len(lines) <= 1 {
		sp = "\n"
		lines = strings.Split(string(data), "\n")
	}

	var newMline bool
	var hadCandidate bool
	var sdp []string

	//log.Println("[sdp] replace candidates, sdp lines=", len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "m=") {
			if newMline && !hadCandidate {
				sdp = append(sdp, candidates...)
				sdp = append(sdp, "a=end-of-candidates")
			}
			newMline = true
			hadCandidate = false
			sdp = append(sdp, line)
		} else if strings.HasPrefix(line, "a=candidate:") {
			// drop it
		} else if strings.HasPrefix(line, "a=end-of-candidates") {
			hadCandidate = true
			sdp = append(sdp, candidates...)
			sdp = append(sdp, line)
		} else if len(line) > 2 {
			sdp = append(sdp, line)
		} else {
			if newMline && !hadCandidate {
				sdp = append(sdp, candidates...)
				sdp = append(sdp, "a=end-of-candidates")
			}
			sdp = append(sdp, line)
		}
	}

	return []byte(strings.Join(sdp, sp))
}

// GetSdpCandidates to parse candidates from sdp
func GetSdpCandidates(data []byte) []string {
	lines := strings.Split(string(data), "\r\n")
	if len(lines) <= 1 {
		lines = strings.Split(string(data), "\n")
	}

	var candidates []string
	//log.Println("[sdp] replace candidates, sdp lines=", len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, "a=candidate:") {
			candidates = append(candidates, line)
			// skip
		} else if strings.HasPrefix(line, "a=end-of-candidates") {
			break
		}
	}
	return candidates
}

// a=candidate:1 1 udp 2113937151 192.168.1.1 5000 typ host
// a=candidate:2 1 tcp 1518280447 192.168.1.1 443 typ host tcptype passive
type Candidate struct {
	Foundation  string
	ComponentId int    // 1-256, e.g., RTP-1, RTCP-2
	Transport   string // udp/tcp
	Priority    int    // 1-(2^31 - 1)
	RelAddr     string // raddr
	RelPort     string // rport
	CandType    string // typ host/srflx/prflx/relay
	NetType     string // network type
}

func ParseCandidates(lines []string) []Candidate {
	var cands []Candidate
	for _, line := range lines {
		if !strings.HasPrefix(line, "a=candidate:") {
			continue
		}
		items := strings.Split(line, " ")
		if len(items) < 8 {
			log.Println("[sdp] invalid sdp candidate:", line)
			continue
		}

		foundation := ""
		if heads := strings.Split(items[0], ":"); len(heads) == 2 {
			foundation = heads[1]
		}

		// typ host/srflx/prflx/relay
		candType := items[6] + " " + items[7]

		// tcptype passive
		netType := ""
		if len(items) >= 9 {
			netType = strings.Join(items[8:], " ")
		}

		cand := Candidate{foundation, Atoi(items[1]), items[2], Atoi(items[3]), items[4], items[5], candType, netType}
		cands = append(cands, cand)
	}

	return cands
}

func LoadX509Certificate(certFile string) (*x509.Certificate, error) {
	if cf, err := ioutil.ReadFile(certFile); err != nil {
		return nil, err
	} else {
		cpb, _ := pem.Decode(cf)
		if crt, err := x509.ParseCertificate(cpb.Bytes); err != nil {
			return nil, err
		} else {
			return crt, nil
		}
	}
}
