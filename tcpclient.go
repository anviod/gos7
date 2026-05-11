package gos7

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
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
func (h *TCPClientHandler) GetPDULength() int {
	return h.PDULength
}

// NewTCPClientHandler allocates a new TCPClientHandler.
func NewTCPClientHandler(address string, rack int, slot int) *TCPClientHandler {
	h := &TCPClientHandler{}
	h.Address = address
	h.Timeout = tcpTimeout
	h.IdleTimeout = tcpIdleTimeout
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

// tcpTransporter implements Transporter interface.
type tcpTransporter struct {
	// Connect string
	Address string
	// Connect & Read timeout
	Timeout time.Duration
	// Idle timeout to close the connection
	IdleTimeout time.Duration
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
	mb.mu.Lock()
	defer mb.mu.Unlock()
	// Set timer to close when idle
	mb.lastActivity = time.Now()
	mb.startCloseTimer()
	// Set write and read timeout
	var timeout time.Time
	if mb.Timeout > 0 {
		timeout = mb.lastActivity.Add(mb.Timeout)
	}
	if mb.conn == nil {
		err = fmt.Errorf("Connection to address %s is null", mb.Address)
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
	data := make([]byte, tcpMaxLength)
	length := 0
	for !done && err == nil {
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
	response = data[0:length]
	mb.logf("s7: received % x\n", response)
	return
}

// Connect establishes a new connection to the address in Address.
// Connect and Close are exported so that multiple requests can be done with one session
func (mb *tcpTransporter) Connect() error {
	// mb.mu.Lock()
	// defer mb.mu.Unlock()

	return mb.connect()
}
func (mb *tcpTransporter) tcpConnect() error {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	if mb.conn == nil {
		dialer := net.Dialer{Timeout: mb.Timeout}
		conn, err := dialer.Dial("tcp", mb.Address)
		if err != nil {
			if conn != nil {
				_ = conn.Close()
			}
			return err
		}
		mb.conn = conn
	}
	return nil
}
func (mb *tcpTransporter) connect() error {
	//first stage: TCP connection
	err := mb.tcpConnect()
	if err != nil {
		return err
	}
	//second stage: ISOTCP (ISO 8073) Connection
	err = mb.isoConnect()
	if err != nil {
		if mb.conn != nil {
			_ = mb.conn.Close()
		}
		return err
	}
	// Third stage : S7 protocol data unit negotiation
	return mb.negotiatePduLength()

}

func (mb *tcpTransporter) isoConnect() error {
	msg := make([]byte, len(isoConnectionRequestTelegram))
	copy(msg, isoConnectionRequestTelegram)
	msg[16] = mb.localTSAPHigh
	msg[17] = mb.localTSAPLow
	msg[20] = mb.remoteTSAPHigh
	msg[21] = mb.remoteTSAPLow

	// Sends the connection request telegram
	response, err := mb.Send(msg)
	if size := len(response); size == 22 {
		if mb.LastPDUType != byte(0xD0) { // 0xD0 = CC Connection confirm
			err = fmt.Errorf("errIsoConnect")
		}
	} else {
		err = fmt.Errorf("%s", ErrorText(errIsoInvalidPDU))
	}
	return err
}
func (mb *tcpTransporter) negotiatePduLength() error {
	// Set PDU Size Requested //lth
	pduSizePackage := make([]byte, len(s7PDUNegogiationTelegram))
	copy(pduSizePackage, s7PDUNegogiationTelegram)
	binary.BigEndian.PutUint16(pduSizePackage[23:], uint16(pduSizeRequested))
	// Sends the connection request telegram
	response, err := mb.Send(pduSizePackage)
	length := len(response)
	if length == 27 && response[17] == 0 && response[18] == 0 { // 20 = size of Negotiate Answer
		// Get PDU Size Negotiated
		mb.PDULength = int(binary.BigEndian.Uint16(response[25:]))
		if mb.PDULength <= 0 {
			err = fmt.Errorf("%s", ErrorText(errCliNegotiatingPDU))
		}
	} else {
		err = fmt.Errorf("%s", ErrorText(errCliNegotiatingPDU))
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

// Close closes current connection.
func (mb *tcpTransporter) Close() error {
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

func (mb *tcpTransporter) logf(format string, v ...interface{}) {
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
	idle := time.Now().Sub(mb.lastActivity)
	// 如果空闲时间超过 IdleTimeout，则关闭连接
	if idle >= mb.IdleTimeout {
		mb.logf("s7: closing connection due to idle timeout: %v", idle)
		mb.close()
	}
}

// reserve for future use, need to verify the request and response
func (mb *tcpPackager) Verify(request []byte, response []byte) (err error) {
	return
}
