package tinyio

import (
	"net"
	"time"
	_ "unsafe"
)

// A Netdever is a network device driver for Tinygo; Tinygo's network device
// driver model.
type Netdever interface {

	// NetConnect device to IP network
	NetConnect() error

	// NetDisconnect device from IP network
	NetDisconnect()

	// NetNotify to register callback for network events
	// NetNotify(func(Event))

	// GetHostByName returns the IP address of either a hostname or IPv4
	// address in standard dot notation
	GetHostByName(name string) (net.IP, error)

	// GetHardwareAddr returns device MAC address
	GetHardwareAddr() (net.HardwareAddr, error)

	// GetIPAddr returns IP address assigned to device, either by DHCP or
	// statically
	GetIPAddr() (net.IP, error)

	// Socketer is Berkely Sockets-like interface
	Socketer
}

type AddressFamily int
type SockType int
type Protocol int
type SockAddr struct {
	port [2]byte // Network byte order
	ip   [4]byte // Network byte order
}
type SockFlags int
type SockOpt int
type SockOptLevel int
type Sockfd int

// Berkely Sockets-like interface.  See man page for socket(2), etc.
type Socketer interface {
	Socket(family AddressFamily, sockType SockType, protocol Protocol) (Sockfd, error)
	Bind(sockfd Sockfd, myaddr SockAddr) error
	Connect(sockfd Sockfd, servaddr SockAddr) error
	Listen(sockfd Sockfd, backlog int) error
	Accept(sockfd Sockfd, peer SockAddr) error
	Send(sockfd Sockfd, buff []byte, flags SockFlags, timeout time.Duration) (int, error)
	SendTo(sockfd Sockfd, buff []byte, flags SockFlags, to SockAddr,
		timeout time.Duration) (int, error)
	Recv(sockfd Sockfd, buff []byte, flags SockFlags, timeout time.Duration) (int, error)
	RecvFrom(sockfd Sockfd, buff []byte, flags SockFlags, from SockAddr,
		timeout time.Duration) (int, error)
	Close(sockfd Sockfd) error
	SetSockOpt(sockfd Sockfd, level SockOptLevel, opt SockOpt, value interface{}) error
}

func UseNetdever(dev Netdever) {
	UseNetdevDirect(netdeverWrapper{Netdever: dev})
}

//go:linkname UseNetdev net.useDev
func UseNetdevDirect(dev dev)

type netdeverWrapper struct {
	Netdever
}

func (w netdeverWrapper) Socket(family int, sockType uint8, protocol int) (fd uintptr, err error) {
	fde, err := w.Netdever.Socket(AddressFamily(family), SockType(sockType), Protocol(protocol))
	return uintptr(fde), err
}

func (w netdeverWrapper) Bind(sockfd uintptr, addr net.Addr) error {
	return w.Netdever.Bind(Sockfd(sockfd), netAddrToSockAddr(addr))
}
func (w netdeverWrapper) Connect(sockfd uintptr, servaddr net.Addr) error {
	return w.Netdever.Connect(Sockfd(sockfd), netAddrToSockAddr(servaddr))
}
func (w netdeverWrapper) Listen(sockfd uintptr, backlog int) error {
	return w.Netdever.Listen(Sockfd(sockfd), backlog)
}
func (w netdeverWrapper) Accept(sockfd uintptr, peer net.Addr) (uintptr, error) {
	err := w.Netdever.Accept(Sockfd(sockfd), netAddrToSockAddr(peer))
	return 0, err
}
func (w netdeverWrapper) Send(sockfd uintptr, buf []byte, flags uint16, timeout time.Duration) (int, error) {
	return w.Netdever.Send(Sockfd(sockfd), buf, SockFlags(flags), timeout)
}
func (w netdeverWrapper) Recv(sockfd uintptr, buf []byte, flags uint16, timeout time.Duration) (int, error) {
	return w.Netdever.Recv(Sockfd(sockfd), buf, SockFlags(flags), timeout)
}
func (w netdeverWrapper) Close(sockfd uintptr) error {
	return w.Netdever.Close(Sockfd(sockfd))
}
func (w netdeverWrapper) SetSockOpt(sockfd uintptr, level, opt int, optionValue any) error {
	return w.Netdever.SetSockOpt(Sockfd(sockfd), SockOptLevel(level), SockOpt(opt), optionValue)
}
func netAddrToSockAddr(addr net.Addr) SockAddr {
	return SockAddr{}
}

// dev drivers implement the net.dev interface.
//
// A Netdever is passed to the "net" package using netdev.Use().
//
// Just like a net.Conn, multiple goroutines may invoke methods on a Netdever
// simultaneously.
type dev interface {

	// NetConnect device to IP network
	NetConnect() error

	// NetDisconnect device from IP network
	NetDisconnect()

	// NetNotify to register callback for network events
	// NetNotify(func(Event))

	// GetHostByName returns the IP address of either a hostname or IPv4
	// address in standard dot notation
	GetHostByName(name string) (net.IP, error)

	// GetHardwareAddr returns device MAC address
	GetHardwareAddr() (net.HardwareAddr, error)

	// GetIPAddr returns IP address assigned to device, either by DHCP or
	// statically
	GetIPAddr() (net.IP, error)

	// Socketer is a Berkely Sockets-like interface
	socketer
}

// Berkely Sockets-like interface.  See man page for socket(2), etc.
//
// Multiple goroutines may invoke methods on a Socketer simultaneously.
type socketer interface {
	// # Socket Address family argument
	//
	// Socket address families specifies a communication domain:
	//  - AF_UNIX, AF_LOCAL(synonyms): Local communication For further information, see unix(7).
	//  - AF_INET: IPv4 Internet protocols.  For further information, see ip(7).
	//
	// # Socket type argument
	//
	// Socket types which specifies the communication semantics.
	//  - SOCK_STREAM: Provides sequenced, reliable, two-way, connection-based
	//  byte streams.  An out-of-band data transmission mechanism may be supported.
	//  - SOCK_DGRAM: Supports datagrams (connectionless, unreliable messages of
	//  a fixed maximum length).
	//
	// The type argument serves a second purpose: in addition to specifying a
	// socket type, it may include the bitwise OR of any of the following values,
	// to modify the behavior of socket():
	//  - SOCK_NONBLOCK: Set the O_NONBLOCK file status flag on the open file description.
	//
	// # Socket protocol argument
	//
	// The protocol specifies a particular protocol to be used with the
	// socket.  Normally only a single protocol exists to support a
	// particular socket type within a given protocol family, in which
	// case protocol can be specified as 0. However, it is possible
	// that many protocols may exist, in which case a particular
	// protocol must be specified in this manner.
	//
	// # Return value
	//
	// On success, a file descriptor for the new socket is returned. Quoting man pages:
	// "On error, -1 is returned, and errno is set to indicate the error." Since
	// this is not C we may use a error type native to Go to represent the error
	// ocurred which by itself not only notifies of an error but also provides
	// information on the error as a human readable string when calling the Error method.
	Socket(family int, sockType uint8, protocol int) (fd uintptr, err error)

	Bind(sockfd uintptr, addr net.Addr) error
	Connect(sockfd uintptr, servaddr net.Addr) error
	Listen(sockfd uintptr, backlog int) error
	Accept(sockfd uintptr, peer net.Addr) (uintptr, error)
	// # Flags argument
	//
	// The flags argument is formed by ORing one or more of the following values:
	//  - MSG_CMSG_CLOEXEC: Set the close-on-exec flag for the file descriptor. Unix.
	//  - MSG_DONTWAIT: Enables nonblocking operation. If call would block then returns error.
	//  - MSG_ERRQUEUE: (see manpage) his flag specifies that queued errors should be received
	//  from the socket error queue.
	//  - MSG_OOB: his flag requests receipt of out-of-band data that would not be received in the normal data stream.
	//  - MSG_PEEK: This flag causes the receive operation to return data from
	//  the beginning of the receive queue without removing that data from the queue.
	//  - MSG_TRUNC: Ask for real length of datagram even when it was longer than passed buffer.
	//  - MSG_WAITALL: This flag requests that the operation block until the full request is satisfied.
	Send(sockfd uintptr, buf []byte, flags uint16, timeout time.Duration) (int, error)
	Recv(sockfd uintptr, buf []byte, flags uint16, timeout time.Duration) (int, error)
	Close(sockfd uintptr) error
	// SetSockOpt manipulates options for the socket
	// referred to by the file descriptor sockfd.  Options may exist at
	// multiple protocol levels; they are always present at the
	// uppermost socket level.
	//
	// # Level argument
	//
	// When manipulating socket options, the level at which the option
	// resides and the name of the option must be specified.  To
	// manipulate options at the sockets API level, level is specified
	// as SOL_SOCKET.  To manipulate options at any other level the
	// protocol number of the appropriate protocol controlling the
	// option is supplied.  For example, to indicate that an option is
	// to be interpreted by the TCP protocol, level should be set to the
	// protocol number of TCP; see getprotoent(3).
	//
	// # Option argument
	//
	// The arguments optval and optlen are used to access option values
	// for setsockopt().  For getsockopt() they identify a buffer in
	// which the value for the requested option(s) are to be returned.
	// In Go we provide developers with an `any` interface to be able
	// to pass driver-specific configurations.
	SetSockOpt(sockfd uintptr, level, opt int, optionValue any) error
}
