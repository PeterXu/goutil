package util

import (
	"errors"
	"log"
	"net"
	"os"
	"strings"
	"syscall"
)

// NewNetConn return a new net.Conn object with caching function.
func NewNetConn(c net.Conn) *NetConn {
	return &NetConn{nil, c, c}
}

// NetConn extends net.Conn
type NetConn struct {
	cached   []byte
	nc       net.Conn
	net.Conn // most methods of net.Conn are embedded
}

func (c *NetConn) LocalAddr() net.Addr {
	return c.nc.LocalAddr()
}

func (c *NetConn) RemoteAddr() net.Addr {
	return c.nc.RemoteAddr()
}

func (c *NetConn) preload_(n int) error {
	if n <= 0 {
		return nil
	}
	hadLen := len(c.cached)
	if hadLen >= n {
		return nil
	} else {
		buf := make([]byte, n-hadLen)
		nret, err := c.nc.Read(buf)
		if err != nil {
			return err
		}
		c.cached = append(c.cached, buf[0:nret]...)
		if nret != len(buf) {
			return errors.New("[NetConn] no enough data")
		}
		return nil
	}
}

func (c *NetConn) Peek(n int) ([]byte, error) {
	err := c.preload_(n)
	if err != nil {
		return nil, err
	}
	return c.cached[0:n], nil
}

func (c *NetConn) Read(p []byte) (int, error) {
	need := Min(len(c.cached), len(p))
	if need > 0 {
		copy(p, c.cached[0:need])
		c.cached = c.cached[need:]
		return need, nil
	} else {
		return c.nc.Read(p)
	}
}

func (c *NetConn) Write(p []byte) (int, error) {
	return c.nc.Write(p)
}

func (c *NetConn) Close() error {
	return c.nc.Close()
}

// NetAddrString return a complete network string: "udp|tcp://host:port".
func NetAddrString(addr net.Addr) string {
	if strings.Contains(addr.String(), "://") {
		return addr.String()
	} else {
		return addr.Network() + "://" + addr.String()
	}
}

// SocketFD for system socket description.
type SocketFD interface {
	File() (f *os.File, err error)
}

// SetSocketReuseAddr to set socket with SO_REUSEADDR.
func SetSocketReuseAddr(sock SocketFD) {
	if file, err := sock.File(); err == nil {
		//log.Println("set reuse addr")
		syscall.SetsockoptInt(int(file.Fd()), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	}
}

// LocalIP tries to determine a non-loopback address for local machine
func LocalIP() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() {
			if ipnet.IP.To4() != nil || ipnet.IP.To16() != nil {
				return ipnet.IP, nil
			}
		}
	}
	return nil, nil
}

// LocalIPString to return a non-loopback address string for local machine.
func LocalIPString() string {
	ip, err := LocalIP()
	if err != nil {
		log.Println("Error determining local ip address. ", err)
		return ""
	}
	if ip == nil {
		log.Println("Could not determine local ip address")
		return ""
	}
	return ip.String()
}

// LookupIP looks up host using the local resolver.
// It returns a host's IPv4 address (non-loopback).
func LookupIP(host string) string {
	hostIp := host
	if ips, err := net.LookupIP(host); err == nil {
		for _, ip := range ips {
			if ip.IsGlobalUnicast() && (ip.To4() != nil) {
				hostIp = ip.String()
				break
			}
		}
	}
	return hostIp
}

// return a complete network string: "udp|tcp://host:port".
func AddrToString(addr net.Addr) string {
	if strings.Contains(addr.String(), "://") {
		return addr.String()
	} else {
		return addr.Network() + "://" + addr.String()
	}
}
