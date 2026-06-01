package gos7

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// 默认的 TCP 超时时间
	tcpTimeout     = 10 * time.Second // TCP 连接超时时间
	tcpIdleTimeout = 60 * time.Second // TCP 空闲超时时间
	tcpMaxLength   = 2084             // TCP 消息的最大长度
	// 消息相关
	pduSizeRequested = 480 // 请求的 PDU 大小
	isoTCP           = 102 // 默认的 ISO/TCP 端口
	isoHSize         = 7   // TPKT+COTP 头部大小
	minPduSize       = 16  // 最小的 PDU 大小
	// 客户端连接类型
	connectionTypePG    = 1 // 以 PG 方式连接到 例200 PLC 0x101 01.01
	connectionTypeOP    = 2 // 以 OP 方式连接到 	  PLC 0x201 02.01
	connectionTypeBasic = 3 // 基本连接			 		  0x301 03.01
)

// TCPClientHandler implements Packager and Transporter interface.
type TCPClientHandler struct {
	tcpPackager
	tcpTransporter
}

func (h *TCPClientHandler) LocalAddr() string {
	return h.tcpTransporter.LocalAddr()
}

// GetPDULength implements the PDUProvider interface
// GetPDULength 实现 PDUProvider 接口
func (h *TCPClientHandler) GetPDULength() int {
	return h.PDULength
}

// Connect establishes a new connection to the PLC.
// Connect 建立到 PLC 的新连接
func (h *TCPClientHandler) Connect() error {
	return h.tcpTransporter.Connect()
}

// ConnectContext establishes a new connection with context support for cancellation.
// This allows callers to cancel in-flight TCP dials via context (e.g., during graceful shutdown).
// ConnectContext 建立一个新连接，支持通过 context 进行取消操作
// 这允许调用者通过 context 取消正在进行的 TCP 连接（例如在优雅关闭期间）
func (h *TCPClientHandler) ConnectContext(ctx context.Context) error {
	return h.tcpTransporter.ConnectContext(ctx)
}

// NewTCPClientHandler allocates a new TCPClientHandler.
func NewTCPClientHandler(address string, rack int, slot int) *TCPClientHandler {
	h := &TCPClientHandler{}
	h.Address = address
	h.Timeout = tcpTimeout
	h.IdleTimeout = tcpIdleTimeout
	h.CloseTimeout = defaultCloseTimeout
	h.ConnectionType = connectionTypeBasic // Connect to the PLC with basic connection type
	h.PDULength = pduSizeRequested         // Set default PDU length
	remoteTSAP := uint16(h.ConnectionType)<<8 + (uint16(rack) * 0x20) + uint16(slot)
	h.setConnectionParameters(address, 0x0100, remoteTSAP)
	return h
}

// NewTCPClientHandlerWithConnectType allocates a new TCPClientHandler with connection type.
func NewTCPClientHandlerWithConnectType(address string, rack int, slot int, connectType int) *TCPClientHandler {
	h := &TCPClientHandler{}
	h.Address = address
	h.Timeout = tcpTimeout
	h.IdleTimeout = tcpIdleTimeout
	h.CloseTimeout = defaultCloseTimeout
	h.ConnectionType = connectType
	h.PDULength = pduSizeRequested // Set default PDU length
	remoteTSAP := uint16(h.ConnectionType)<<8 + (uint16(rack) * 0x20) + uint16(slot)
	h.setConnectionParameters(address, 0x0100, remoteTSAP)
	return h
}

// TCPClient creator for a TCP client with address, rack and slot, implement from interface client
func TCPClient(address string, rack int, slot int) Client {
	handler := NewTCPClientHandler(address, rack, slot)
	return NewClient(handler)
}

// TCPClientWithConnectType creator for a TCP client with address, rack, slot and connect type, implement from interface client
func TCPClientWithConnectType(address string, rack int, slot int, connectType int) Client {
	handler := NewTCPClientHandlerWithConnectType(address, rack, slot, connectType)
	return NewClient(handler)
}

// tcpPackager implements Packager interface.
type tcpPackager struct {
	//reserve for future use, this package should be pass into trans ID, pack ID
	//or somethingelse to verify the request and response
}

// bufferPool is a pool of reusable buffers for TCP send/receive operations.
// Uses tiered pooling: small buffers (512B) for typical PDU responses,
// large buffers (2084B) for maximum size. This reduces memory usage on ARM/low-mem devices.
var (
	// smallBufferPool pools 512-byte buffers for typical PDU operations (PDU size <= 480 + header)
	smallBufferPool = sync.Pool{
		New: func() any {
			buf := make([]byte, 512)
			return &buf
		},
	}
	// largeBufferPool pools 2084-byte buffers for maximum TCP message size
	largeBufferPool = sync.Pool{
		New: func() any {
			buf := make([]byte, tcpMaxLength)
			return &buf
		},
	}
)

// getBuffer returns a buffer from the appropriate pool based on needed size.
// getBuffer 根据所需大小从相应的池中获取缓冲区。
func getBuffer(size int) *[]byte {
	if size <= 512 {
		return smallBufferPool.Get().(*[]byte)
	}
	return largeBufferPool.Get().(*[]byte)
}

// putBuffer returns a buffer to the appropriate pool.
// putBuffer 将缓冲区归还到相应的池。
func putBuffer(bufPtr *[]byte) {
	if cap(*bufPtr) <= 512 {
		smallBufferPool.Put(bufPtr)
	} else {
		largeBufferPool.Put(bufPtr)
	}
}

// defaultCloseTimeout is the default timeout for graceful close waiting for peer to close.
const defaultCloseTimeout = 5 * time.Second

// closeState represents the TCP connection close state.
// It tracks the lifecycle of a connection from open to closed,
// including the half-closed state during graceful shutdown.
type closeState int

// closeState constants represent the possible states of a TCP connection during shutdown.
const (
	// closeStateUnknown indicates the connection state is unknown or not initialized.
	closeStateUnknown closeState = iota
	// closeStateOpen indicates the connection is open and active.
	closeStateOpen
	// closeStateHalfClosed indicates one direction of the connection has been closed
	// (typically after sending FIN), but the other direction may still be open.
	// This is the TCP half-close state.
	closeStateHalfClosed
	// closeStateClosed indicates the connection has been fully closed.
	closeStateClosed
)

// String returns a human-readable representation of the close state.
func (s closeState) String() string {
	switch s {
	case closeStateUnknown:
		return "unknown"
	case closeStateOpen:
		return "open"
	case closeStateHalfClosed:
		return "half_closed"
	case closeStateClosed:
		return "closed"
	default:
		return "invalid"
	}
}

// tcpTransporter implements Transporter interface.
type tcpTransporter struct {
	// Connect string
	Address string
	// Connect & Read timeout
	Timeout time.Duration
	// Idle timeout to close the connection
	IdleTimeout time.Duration
	// CloseTimeout is the timeout for graceful close waiting for peer to close.
	// Default is 5 seconds (defaultCloseTimeout). If set to 0, Close() performs
	// immediate close without graceful shutdown. Set to -1 to disable.
	CloseTimeout time.Duration
	// Transmission logger
	Logger *log.Logger

	// TCP connection
	mu           sync.Mutex
	conn         net.Conn
	closeTimer   *time.Timer
	lastActivity time.Time

	localTSAP, remoteTSAP uint16

	localTSAPHigh, localTSAPLow   byte
	remoteTSAPHigh, remoteTSAPLow byte
	ConnectionType                int
	LastPDUType                   byte

	PDULength int

	closeState closeState
}

func (mb *tcpTransporter) setConnectionParameters(address string, localTSAP uint16, remoteTSAP uint16) {
	locTSAP := localTSAP & 0x0000FFFF
	remTSAP := remoteTSAP & 0x0000FFFF
	if len(strings.Split(address, ":")) < 2 {
		mb.Address = address + ":" + strconv.Itoa(isoTCP) //ip:102
	} else {
		mb.Address = address
	}
	mb.localTSAPHigh = byte(locTSAP >> 8)
	mb.localTSAPLow = byte(locTSAP & 0x00FF)
	mb.remoteTSAPHigh = byte(remTSAP >> 8)
	mb.remoteTSAPLow = byte(remTSAP & 0x00FF)
}

// Send sends data to server and ensures response length is greater than header length.
func (mb *tcpTransporter) Send(request []byte) (response []byte, err error) {
	return mb.SendWithContext(context.Background(), request)
}

// SendWithContext sends data to server with context support for cancellation.
// SendWithContext 发送数据到服务器，支持通过 context 进行取消操作
func (mb *tcpTransporter) SendWithContext(ctx context.Context, request []byte) (response []byte, err error) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	// Check context before starting
	if err = ctx.Err(); err != nil {
		return
	}
	// Set timer to close when idle
	mb.lastActivity = time.Now()
	mb.startCloseTimer()
	// Set write and read timeout
	var timeout time.Time
	if mb.Timeout > 0 {
		timeout = mb.lastActivity.Add(mb.Timeout)
	}
	if mb.conn == nil {
		err = fmt.Errorf("send: connection to address %s is null", mb.Address)
		return
	}
	if err = mb.conn.SetDeadline(timeout); err != nil {
		return
	}
	// Send data
	mb.logf("s7: sending % x", request)
	if _, err = mb.conn.Write(request); err != nil {
		return
	}
	done := false
	// Use tiered pool: small buffer for typical PDU, large for max size
	bufPtr := getBuffer(tcpMaxLength)
	data := *bufPtr
	defer putBuffer(bufPtr)
	length := 0
	for !done && err == nil {
		// Check context cancellation between reads
		if err = ctx.Err(); err != nil {
			return
		}
		// Get TPKT (4 bytes)
		if _, err = io.ReadFull(mb.conn, data[:4]); err != nil {
			log.Printf("%T %+v", err, err)
			return
		}
		// Read length, ignore transaction & protocol id (4 bytes)
		length = int(binary.BigEndian.Uint16(data[2:]))
		if length == isoHSize {
			_, err = io.ReadFull(mb.conn, data[4:7])
			if err != nil { // Skip remaining 3 bytes and Done is still false
				return
			}
		} else {
			if length > pduSizeRequested+isoHSize || length < minPduSize {
				err = fmt.Errorf("s7: invalid pdu")
				return
			}
			done = true
		}
	}
	// Skip remaining 3 COTP bytes
	_, err = io.ReadFull(mb.conn, data[4:7])
	if err != nil {
		return
	}
	mb.LastPDUType = data[5] // Stores PDU Type, we need it
	// Receives the S7 Payload
	_, err = io.ReadFull(mb.conn, data[7:length])
	if err != nil {
		return
	}
	// Copy response data to a new slice before returning the buffer to pool.
	// response must not reference the pooled buffer after this function returns.
	response = make([]byte, length)
	copy(response, data[:length])
	mb.logf("s7: received % x\n", response)
	return
}

// Connect establishes a new connection to the address in Address.
// Connect and Close are exported so that multiple requests can be done with one session
// Connect 建立到地址的新连接
// Connect 和 Close 是导出的，以便可以在一个会话中执行多个请求
func (mb *tcpTransporter) Connect() error {
	return mb.ConnectContext(context.Background())
}

// ConnectContext establishes a new connection with context support for cancellation.
// This allows callers to cancel in-flight TCP dials via context (e.g., during graceful shutdown).
// ConnectContext 建立一个新连接，支持通过 context 进行取消操作
// 这允许调用者通过 context 取消正在进行的 TCP 连接（例如在优雅关闭期间）
func (mb *tcpTransporter) ConnectContext(ctx context.Context) error {
	return mb.connect(ctx)
}
func (mb *tcpTransporter) tcpConnect(ctx context.Context) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	if mb.conn == nil {
		// Check context before dialing
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("tcp connect: %w", err)
		}
		dialer := net.Dialer{Timeout: mb.Timeout}
		conn, err := dialer.DialContext(ctx, "tcp", mb.Address)
		if err != nil {
			return fmt.Errorf("tcp connect to %s: %w", mb.Address, err)
		}
		mb.conn = conn
	}
	return nil
}
func (mb *tcpTransporter) connect(ctx context.Context) error {
	//first stage: TCP connection
	err := mb.tcpConnect(ctx)
	if err != nil {
		return err
	}
	//second stage: ISOTCP (ISO 8073) Connection
	err = mb.isoConnect(ctx)
	if err != nil {
		mb.mu.Lock()
		mb.close()
		mb.mu.Unlock()
		return err
	}
	// Third stage : S7 protocol data unit negotiation
	err = mb.negotiatePduLength(ctx)
	if err != nil {
		mb.mu.Lock()
		mb.close()
		mb.mu.Unlock()
		return err
	}
	return nil
}

func (mb *tcpTransporter) isoConnect(ctx context.Context) error {
	// Check context before starting ISO connection
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("iso connect: %w", err)
	}

	msg := make([]byte, len(isoConnectionRequestTelegram))
	copy(msg, isoConnectionRequestTelegram)
	msg[16] = mb.localTSAPHigh
	msg[17] = mb.localTSAPLow
	msg[20] = mb.remoteTSAPHigh
	msg[21] = mb.remoteTSAPLow

	// Sends the connection request telegram
	response, err := mb.SendWithContext(ctx, msg)
	if err != nil {
		return fmt.Errorf("iso connect: %w", err)
	}
	if size := len(response); size == 22 {
		if mb.LastPDUType != byte(0xD0) { // 0xD0 = CC Connection confirm
			err = fmt.Errorf("iso connect: unexpected PDU type 0x%02x", mb.LastPDUType)
		}
	} else {
		err = fmt.Errorf("iso connect: %s (size=%d)", ErrorText(errIsoInvalidPDU), size)
	}
	return err
}
func (mb *tcpTransporter) negotiatePduLength(ctx context.Context) error {
	// Check context before starting PDU negotiation
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("pdu negotiation: %w", err)
	}

	// Set PDU Size Requested //lth
	pduSizePackage := make([]byte, len(s7PDUNegogiationTelegram))
	copy(pduSizePackage, s7PDUNegogiationTelegram)
	binary.BigEndian.PutUint16(pduSizePackage[23:], uint16(pduSizeRequested))
	// Sends the connection request telegram
	response, err := mb.SendWithContext(ctx, pduSizePackage)
	if err != nil {
		return fmt.Errorf("pdu negotiation: %w", err)
	}
	length := len(response)
	if length == 27 && response[17] == 0 && response[18] == 0 { // 20 = size of Negotiate Answer
		// Get PDU Size Negotiated
		mb.PDULength = int(binary.BigEndian.Uint16(response[25:]))
		if mb.PDULength <= 0 {
			err = fmt.Errorf("pdu negotiation: %s", ErrorText(errCliNegotiatingPDU))
		}
	} else {
		err = fmt.Errorf("pdu negotiation: %s (length=%d)", ErrorText(errCliNegotiatingPDU), length)
	}
	return err
}
func (mb *tcpTransporter) startCloseTimer() {
	if mb.IdleTimeout <= 0 {
		return
	}

	if mb.closeTimer == nil {
		mb.closeTimer = time.AfterFunc(mb.IdleTimeout, mb.closeIdle)
	} else {
		mb.closeTimer.Reset(mb.IdleTimeout)
	}
}

// Close closes the current connection with graceful handling of TCP half-close state.
// If CloseTimeout > 0 (default is 5 seconds), it will wait up to that duration for the peer
// to close their side. If the peer does not close within the timeout, the connection will
// be forcefully closed.
//
// For immediate forceful close without waiting, use CloseForce().
func (mb *tcpTransporter) Close() error {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	if mb.CloseTimeout != 0 {
		return mb.closeGracefully()
	}
	return mb.close()
}

// CloseForce closes the connection immediately without waiting for the peer to close.
// This bypasses graceful shutdown and forcibly terminates the connection.
// Use this when immediate termination is required, such as during emergency shutdown.
// Unlike Close(), this method ignores CloseTimeout and does not wait for peer acknowledgment.
func (mb *tcpTransporter) CloseForce() error {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	return mb.close()
}

// flush flushes pending data in the connection,
// returns io.EOF if connection is closed.
func (mb *tcpTransporter) flush(b []byte) (err error) {
	if err = mb.conn.SetReadDeadline(time.Now()); err != nil {
		return
	}
	// Timeout setting will be reset when reading
	if _, err = mb.conn.Read(b); err != nil {
		// Ignore timeout error
		if netError, ok := err.(net.Error); ok && netError.Timeout() {
			err = nil
		}
	}
	return
}

func (mb *tcpTransporter) LocalAddr() string {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	if mb.conn == nil {
		return ""
	}
	addr := mb.conn.LocalAddr()
	if addr == nil {
		return ""
	}
	return addr.String()
}

func (mb *tcpTransporter) logf(format string, v ...any) {
	if mb.Logger != nil {
		mb.Logger.Printf(format, v...)
	}
}

// closeLocked closes current connection. Caller must hold the mutex before calling this method.
func (mb *tcpTransporter) close() (err error) {
	if mb.conn != nil {
		err = mb.conn.Close()
		mb.conn = nil
	}
	return
}

// closeIdle 在连接空闲时间超过 IdleTimeout 时关闭连接。
// 该函数用于自动关闭空闲连接以及时释放资源。
// 它首先检查 IdleTimeout 设置是否大于 0，因为非正数表示不启用空闲超时功能。
// 然后计算自上次活动以来的时间（空闲时间）。如果这个时间超过了 IdleTimeout，它会记录关闭连接的原因并调用 close 方法来关闭连接。
// 该函数通过互斥锁保护，以确保在访问共享资源时的线程安全。
func (mb *tcpTransporter) closeIdle() {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	// 如果 IdleTimeout 不大于 0，则无需处理
	if mb.IdleTimeout <= 0 {
		return
	}
	// 计算空闲时间
	idle := time.Since(mb.lastActivity)
	// 如果空闲时间超过 IdleTimeout，则关闭连接
	if idle >= mb.IdleTimeout {
		mb.logf("s7: closing connection due to idle timeout: %v", idle)
		mb.close()
	}
}

// closeGracefully performs a graceful shutdown of the TCP connection.
// It handles the TCP half-close state by waiting for the peer to close their side
// before fully closing the connection.
//
// The graceful close process:
//  1. Transitions to half-closed state (closeStateHalfClosed)
//  2. Closes the local connection (sends FIN to peer)
//  3. Waits for peer to close their side or timeout (CloseTimeout)
//  4. If timeout expires, forces immediate close
//
// This ensures proper cleanup in scenarios where the peer may need time to
// process remaining data before closing their side of the connection.
func (mb *tcpTransporter) closeGracefully() (err error) {
	if mb.conn == nil {
		mb.closeState = closeStateClosed
		return nil
	}

	if mb.closeState == closeStateClosed {
		return nil
	}

	if mb.closeState == closeStateHalfClosed {
		mb.logf("s7: connection already in half-close state, performing final close")
		return mb.close()
	}

	mb.logf("s7: initiating graceful close, current state: %s", mb.closeState)

	conn := mb.conn
	mb.closeState = closeStateHalfClosed

	if mb.CloseTimeout <= 0 {
		mb.CloseTimeout = defaultCloseTimeout
	}

	conn.Close()

	peerClosedChan := make(chan struct{})
	go func() {
		abuf := smallBufferPool.Get().(*[]byte)
		defer smallBufferPool.Put(abuf)
		buf := *abuf
		for {
			conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			n, err := conn.Read(buf)
			if n > 0 {
				mb.logf("s7: discarded %d bytes during graceful close", n)
				continue
			}
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				break
			}
		}
		close(peerClosedChan)
	}()

	timeout := time.NewTimer(mb.CloseTimeout)
	select {
	case <-peerClosedChan:
		mb.logf("s7: peer closed connection gracefully")
		timeout.Stop()
	case <-timeout.C:
		mb.logf("s7: graceful close timeout (%v) exceeded, forcing close", mb.CloseTimeout)
	}

	return mb.close()
}

// waitForPeerClose waits for the peer to close their side of the connection
// or until the specified timeout expires.
//
// This method is useful for scenarios where you want to confirm the peer has
// finished sending data and closed their write direction before fully closing
// the connection.
//
// Returns nil if peer closed successfully or timeout occurred.
// Returns error if connection is nil or another error occurs.
func (mb *tcpTransporter) waitForPeerClose(timeout time.Duration) error {
	if mb.conn == nil {
		return fmt.Errorf("connection is nil")
	}

	if mb.closeState == closeStateClosed {
		return nil
	}

	peerClosedChan := make(chan error, 1)
	go func() {
		abuf := smallBufferPool.Get().(*[]byte)
		defer smallBufferPool.Put(abuf)
		buf := *abuf
		for {
			readErr := mb.conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			if readErr != nil {
				peerClosedChan <- readErr
				return
			}
			n, err := mb.conn.Read(buf)
			if n > 0 {
				mb.logf("s7: discarded %d bytes while waiting for peer close", n)
				continue
			}
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				peerClosedChan <- err
				return
			}
		}
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case err := <-peerClosedChan:
		if err != nil && err != io.EOF {
			return err
		}
		mb.logf("s7: peer closed connection (err=%v)", err)
		return nil
	case <-timer.C:
		return fmt.Errorf("timeout waiting for peer to close connection")
	}
}

// State returns the current close state of the connection.
func (mb *tcpTransporter) State() closeState {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	return mb.closeState
}

// IsClosed returns true if the connection is fully closed.
func (mb *tcpTransporter) IsClosed() bool {
	return mb.State() == closeStateClosed
}

// IsHalfClosed returns true if the connection is in half-closed state.
// A half-closed connection has had its write side closed (FIN sent)
// but may still receive data from the peer.
func (mb *tcpTransporter) IsHalfClosed() bool {
	return mb.State() == closeStateHalfClosed
}

// reserve for future use, need to verify the request and response
func (mb *tcpPackager) Verify(request []byte, response []byte) (err error) {
	return
}
