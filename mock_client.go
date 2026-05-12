package gos7

import (
	"encoding/binary"
	"sync"
)

// MockClientHandler implements both Packager and Transporter interfaces
// to simulate a PLC connection without requiring actual hardware.
// It maintains an in-memory data store that responds to S7 protocol requests.
//
// MockClientHandler 实现了 Packager 和 Transporter 接口，
// 用于在没有真实 PLC 硬件的情况下模拟 PLC 连接。
// 它维护一个内存数据存储，用于响应 S7 协议请求。
type MockClientHandler struct {
	mu        sync.Mutex
	PDULength int
	data      []byte // simulated PLC memory: 1MB address space
}

// NewMockClientHandler creates a new mock PLC handler with default PDU length.
// NewMockClientHandler 创建一个新的模拟 PLC 处理器，使用默认 PDU 长度。
func NewMockClientHandler() *MockClientHandler {
	return &MockClientHandler{
		PDULength: 480,
		data:      make([]byte, 1024*1024), // 1MB simulated PLC memory
	}
}

// GetPDULength implements the PDUProvider interface
func (m *MockClientHandler) GetPDULength() int {
	return m.PDULength
}

// Verify implements the Packager interface
func (m *MockClientHandler) Verify(request []byte, response []byte) error {
	return nil
}

// Send implements the Transporter interface.
// It parses the S7 request telegram and returns a properly formatted response.
func (m *MockClientHandler) Send(request []byte) (response []byte, err error) {
	if len(request) < 19 {
		return nil, errInvalidRequest()
	}

	// Check S7 protocol ID (byte index 7 after TPKT+COTP)
	if request[7] != 0x32 {
		return nil, errInvalidRequest()
	}

	funcCode := request[17]
	itemCount := int(request[18])

	switch funcCode {
	case 0x04: // Read Var
		if itemCount > 1 {
			return m.handleMultiRead(request)
		}
		return m.handleRead(request)
	case 0x05: // Write Var
		if itemCount > 1 {
			return m.handleMultiWrite(request)
		}
		return m.handleWrite(request)
	}

	return m.buildErrorResponse(), nil
}

// handleRead handles a single read request
func (m *MockClientHandler) handleRead(request []byte) ([]byte, error) {
	if len(request) < 31 {
		return m.buildErrorResponse(), nil
	}

	area := int(request[27])
	dbNumber := int(binary.BigEndian.Uint16(request[25:27]))
	wordLen := int(request[22])
	amount := int(binary.BigEndian.Uint16(request[23:25]))

	// Calculate data start from 3-byte address (bytes 28-30)
	start := int(request[28])<<16 | int(request[29])<<8 | int(request[30])
	// For byte/word/dword types, address is in bits, convert to bytes
	if wordLen != 0x01 && wordLen != 0x1C && wordLen != 0x1D {
		start = start >> 3
	}

	// Calculate actual byte size
	wordSize := dataSizeByte(wordLen)
	if wordSize == 0 {
		wordSize = 1
	}
	dataBytes := amount * wordSize

	// Read from simulated memory
	data := m.readFromMemory(area, dbNumber, start, dataBytes)

	// Build response
	// Total = TPKT(4) + COTP(3) + S7Header(10) + ErrorClass(1) + ErrorCode(1) + Func(1) + ItemCount(1) + ReturnCode(1) + TransportSize(1) + DataLenBits(2) + Data + Padding
	padding := 0
	if len(data)%2 != 0 {
		padding = 1
	}
	respLen := 4 + 3 + 10 + 2 + 2 + 4 + len(data) + padding

	resp := make([]byte, respLen)
	idx := 0

	// TPKT header [0:4]
	resp[idx] = 0x03
	idx++ // [0]
	idx++ // [1] reserved
	binary.BigEndian.PutUint16(resp[idx:], uint16(respLen))
	idx += 2 // [2:4]

	// COTP [4:7]
	resp[idx] = 0x02
	idx++ // [4]
	resp[idx] = 0xF0
	idx++ // [5]
	resp[idx] = 0x80
	idx++ // [6]

	// S7 Header [7:17]
	resp[idx] = 0x32                                                    // Protocol ID
	idx++                                                               // [7]
	resp[idx] = 0x03                                                    // Message type: Ack Data
	idx++                                                               // [8]
	resp[idx] = 0x00                                                    // Reserved
	idx++                                                               // [9]
	resp[idx] = 0x00                                                    // Reserved
	idx++                                                               // [10]
	resp[idx] = request[11]                                             // PDU ref high (echo back)
	idx++                                                               // [11]
	resp[idx] = request[12]                                             // PDU ref low (echo back)
	idx++                                                               // [12]
	binary.BigEndian.PutUint16(resp[idx:], 0x0002)                      // Parameter length = 2
	idx += 2                                                            // [13:15]
	binary.BigEndian.PutUint16(resp[idx:], uint16(len(data)+4+padding)) // Data length
	idx += 2                                                            // [15:17]

	// Error class + Error code [17:19]
	resp[idx] = 0x00 // Error class = no error
	idx++            // [17]
	resp[idx] = 0x00 // Error code = no error
	idx++            // [18]

	// Parameter section [19:21]
	resp[idx] = 0x04 // Function: Read
	idx++            // [19]
	resp[idx] = 0x01 // Item count = 1
	idx++            // [20]

	// Data section - Item result [21:25+]
	resp[idx] = 0xFF                                            // Return code = success
	idx++                                                       // [21]
	resp[idx] = 0x04                                            // Transport size = byte/octet
	idx++                                                       // [22]
	binary.BigEndian.PutUint16(resp[idx:], uint16(len(data)*8)) // Size in bits
	idx += 2                                                    // [23:25]

	// Copy data
	copy(resp[idx:], data)
	idx += len(data)

	// Padding byte if odd
	if padding > 0 {
		resp[idx] = 0x00
	}

	return resp, nil
}

// handleWrite handles a single write request
// Response format (exactly 22 bytes):
//   [0-3]   TPKT
//   [4-6]   COTP
//   [7-16]  S7 Header
//   [17-18] Error class + Error code
//   [19-20] Function + Item count
//   [21]    Return code (0xFF = success)
func (m *MockClientHandler) handleWrite(request []byte) ([]byte, error) {
	if len(request) < 35 {
		return m.buildErrorResponse(), nil
	}

	area := int(request[27])
	dbNumber := int(binary.BigEndian.Uint16(request[25:27]))
	wordLen := int(request[22])
	amount := int(binary.BigEndian.Uint16(request[23:25]))

	// Calculate data start
	start := int(request[28])<<16 | int(request[29])<<8 | int(request[30])
	if wordLen != 0x01 && wordLen != 0x1C && wordLen != 0x1D {
		start = start >> 3
	}

	// Calculate actual byte size
	wordSize := dataSizeByte(wordLen)
	if wordSize == 0 {
		wordSize = 1
	}
	dataBytes := amount * wordSize

	// Extract write data (starts at byte 35 in the request)
	writeData := make([]byte, dataBytes)
	if len(request) > 35 {
		available := min(len(request)-35, dataBytes)
		copy(writeData, request[35:35+available])
	}

	// Write to simulated memory
	m.writeToMemory(area, dbNumber, start, writeData)

	// Build write response (exactly 22 bytes)
	resp := make([]byte, 22)

	// TPKT header [0:4]
	resp[0] = 0x03
	resp[1] = 0x00
	binary.BigEndian.PutUint16(resp[2:], 22)

	// COTP [4:7]
	resp[4] = 0x02
	resp[5] = 0xF0
	resp[6] = 0x80

	// S7 Header [7:17]
	resp[7] = 0x32                                // Protocol ID
	resp[8] = 0x03                                // Message type: Ack Data
	resp[9] = 0x00                                // Reserved
	resp[10] = 0x00                               // Reserved
	resp[11] = request[11]                        // PDU ref high (echo back)
	resp[12] = request[12]                        // PDU ref low (echo back)
	binary.BigEndian.PutUint16(resp[13:], 0x0002) // Parameter length = 2
	binary.BigEndian.PutUint16(resp[15:], 0x0000) // Data length = 0

	// Error class + Error code [17:19]
	resp[17] = 0x00 // Error class = no error
	resp[18] = 0x00 // Error code = no error

	// Parameter section [19:21]
	resp[19] = 0x05 // Function: Write
	resp[20] = 0x01 // Item count = 1

	// Return code [21]
	resp[21] = 0xFF // Success

	return resp, nil
}

// handleMultiRead handles a multi-read request
// Response layout (per AGReadMulti parsing):
//   [0-3]   TPKT
//   [4-6]   COTP
//   [7-16]  S7 Header
//   [17-18] CPU error (2 bytes, 0x0000 = no error)
//   [19]    Function (0x04 = Read)
//   [20]    Items read count
//   [21+]   Data items: ReturnCode(1) + TransportSize(1) + SizeInBits(2) + Data(n) + Padding
func (m *MockClientHandler) handleMultiRead(request []byte) ([]byte, error) {
	if len(request) < 19 {
		return m.buildErrorResponse(), nil
	}

	itemCount := int(request[18])

	// Parse each item parameter (12 bytes per item, starting at offset 19)
	type readItem struct {
		area     int
		dbNumber int
		start    int
		amount   int
		wordLen  int
	}
	items := make([]readItem, itemCount)

	for i := 0; i < itemCount; i++ {
		offset := 19 + i*12
		if offset+12 > len(request) {
			break
		}
		items[i].wordLen = int(request[offset+3])
		items[i].amount = int(binary.BigEndian.Uint16(request[offset+4 : offset+6]))
		items[i].dbNumber = int(binary.BigEndian.Uint16(request[offset+6 : offset+8]))
		items[i].area = int(request[offset+8])
		addr := int(request[offset+9])<<16 | int(request[offset+10])<<8 | int(request[offset+11])
		if items[i].wordLen != 0x01 && items[i].wordLen != 0x1C && items[i].wordLen != 0x1D {
			addr = addr >> 3
		}
		items[i].start = addr
	}

	// Build data section
	dataSection := []byte{}
	for i := 0; i < itemCount; i++ {
		wordSize := dataSizeByte(items[i].wordLen)
		if wordSize == 0 {
			wordSize = 1
		}
		size := items[i].amount * wordSize
		data := m.readFromMemory(items[i].area, items[i].dbNumber, items[i].start, size)

		// Item data: ReturnCode(1) + TransportSize(1) + SizeInBits(2) + Data
		itemData := make([]byte, 4+len(data))
		itemData[0] = 0xFF // Return code = success
		itemData[1] = 0x04 // Transport size = byte
		binary.BigEndian.PutUint16(itemData[2:], uint16(len(data)*8))
		copy(itemData[4:], data)
		dataSection = append(dataSection, itemData...)

		// Padding if odd
		if len(data)%2 != 0 {
			dataSection = append(dataSection, 0x00)
		}
	}

	// Build full response
	paramLen := 2 // function + item count
	dataLen := len(dataSection)
	respLen := 4 + 3 + 10 + 2 + paramLen + dataLen

	resp := make([]byte, respLen)
	idx := 0

	// TPKT [0:4]
	resp[idx] = 0x03
	idx += 2
	binary.BigEndian.PutUint16(resp[idx:], uint16(respLen))
	idx += 2

	// COTP [4:7]
	resp[idx] = 0x02
	idx++
	resp[idx] = 0xF0
	idx++
	resp[idx] = 0x80
	idx++

	// S7 Header [7:17]
	resp[idx] = 0x32
	idx++
	resp[idx] = 0x03 // Ack Data
	idx++
	resp[idx] = 0x00 // Reserved
	idx++
	resp[idx] = 0x00 // Reserved
	idx++
	resp[idx] = 0x00 // PDU ref high
	idx++
	resp[idx] = 0x00 // PDU ref low
	idx++
	binary.BigEndian.PutUint16(resp[idx:], uint16(paramLen))
	idx += 2
	binary.BigEndian.PutUint16(resp[idx:], uint16(dataLen))
	idx += 2

	// CPU error [17:19]
	resp[idx] = 0x00
	idx++
	resp[idx] = 0x00
	idx++

	// Parameter section [19:21]
	resp[idx] = 0x04 // Function: Read
	idx++
	resp[idx] = byte(itemCount)
	idx++

	// Data section
	copy(resp[idx:], dataSection)

	return resp, nil
}

// handleMultiWrite handles a multi-write request
// Response layout (per AGWriteMulti parsing):
//   [0-3]   TPKT
//   [4-6]   COTP
//   [7-16]  S7 Header
//   [17-18] CPU error (2 bytes, 0x0000 = no error)
//   [19]    Function (0x05 = Write)
//   [20]    Items written count
//   [21..21+itemsCount-1] Return code per item (0xFF = success)
func (m *MockClientHandler) handleMultiWrite(request []byte) ([]byte, error) {
	if len(request) < 19 {
		return m.buildErrorResponse(), nil
	}

	itemCount := int(request[18])
	paramLen := int(binary.BigEndian.Uint16(request[13:15]))

	// Parse parameter items (12 bytes per item, starting at offset 19)
	type writeItem struct {
		area     int
		dbNumber int
		start    int
		amount   int
		wordLen  int
	}
	items := make([]writeItem, itemCount)

	for i := 0; i < itemCount; i++ {
		offset := 19 + i*12
		if offset+12 > len(request) {
			break
		}
		items[i].wordLen = int(request[offset+3])
		items[i].amount = int(binary.BigEndian.Uint16(request[offset+4 : offset+6]))
		items[i].dbNumber = int(binary.BigEndian.Uint16(request[offset+6 : offset+8]))
		items[i].area = int(request[offset+8])
		addr := int(request[offset+9])<<16 | int(request[offset+10])<<8 | int(request[offset+11])
		if items[i].wordLen != 0x01 && items[i].wordLen != 0x1C && items[i].wordLen != 0x1D {
			addr = addr >> 3
		}
		items[i].start = addr
	}

	// Find data section start (after parameters)
	dataOffset := 19 + paramLen - 2 // subtract 2 for function+itemCount

	// Process each item's write data
	for i := 0; i < itemCount; i++ {
		if dataOffset+4 > len(request) {
			break
		}

		wordSize := dataSizeByte(items[i].wordLen)
		if wordSize == 0 {
			wordSize = 1
		}
		dataSize := items[i].amount * wordSize

		// Data item header: reserved(1) + transportSize(1) + dataLenInBits(2)
		dataOffset += 4 // skip 4-byte data header

		if dataOffset+dataSize <= len(request) {
			writeData := make([]byte, dataSize)
			copy(writeData, request[dataOffset:dataOffset+dataSize])
			m.writeToMemory(items[i].area, items[i].dbNumber, items[i].start, writeData)
		}

		dataOffset += dataSize
		if dataSize%2 != 0 {
			dataOffset++ // skip padding
		}
	}

	// Build write response
	resultLen := 21 + itemCount
	resp := make([]byte, resultLen)
	idx := 0

	// TPKT
	resp[idx] = 0x03
	idx += 2
	binary.BigEndian.PutUint16(resp[idx:], uint16(resultLen))
	idx += 2

	// COTP
	resp[idx] = 0x02
	idx++
	resp[idx] = 0xF0
	idx++
	resp[idx] = 0x80
	idx++

	// S7 Header
	resp[idx] = 0x32
	idx++
	resp[idx] = 0x03 // Ack Data
	idx++
	resp[idx] = 0x00 // Reserved
	idx++
	resp[idx] = 0x00 // Reserved
	idx++
	resp[idx] = 0x00 // PDU ref high
	idx++
	resp[idx] = 0x00 // PDU ref low
	idx++
	binary.BigEndian.PutUint16(resp[idx:], 0x0002) // Param length
	idx += 2
	binary.BigEndian.PutUint16(resp[idx:], uint16(itemCount)) // Data length = item count
	idx += 2

	// CPU error [17:19]
	resp[idx] = 0x00
	idx++
	resp[idx] = 0x00
	idx++

	// Function + item count [19:21]
	resp[idx] = 0x05 // Function: Write
	idx++
	resp[idx] = byte(itemCount)
	idx++

	// Return codes per item [21+]
	for i := 0; i < itemCount; i++ {
		resp[idx] = 0xFF // Success
		idx++
	}

	return resp, nil
}

// readFromMemory reads data from simulated PLC memory
func (m *MockClientHandler) readFromMemory(area int, dbNumber int, start int, size int) []byte {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]byte, size)
	memOffset := m.calcMemoryOffset(area, dbNumber, start)
	if memOffset >= 0 && memOffset+size <= len(m.data) {
		copy(result, m.data[memOffset:memOffset+size])
	}
	return result
}

// writeToMemory writes data to simulated PLC memory
func (m *MockClientHandler) writeToMemory(area int, dbNumber int, start int, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()

	memOffset := m.calcMemoryOffset(area, dbNumber, start)
	if memOffset >= 0 && memOffset+len(data) <= len(m.data) {
		copy(m.data[memOffset:], data)
	}
}

// calcMemoryOffset maps (area, dbNumber, start) to a flat memory offset
// Uses a simple layout with enough space for common DB numbers:
//   DB area:  base + (dbNumber % 16) * 32KB + start   (supports DB0-DB15, 32KB each)
//   PE area:  512KB + start
//   PA area:  576KB + start
//   MK area:  640KB + start
//   CT area:  704KB + start
//   TM area:  768KB + start
//   Other:    832KB + start
func (m *MockClientHandler) calcMemoryOffset(area int, dbNumber int, start int) int {
	switch area {
	case 0x84: // DB
		return (dbNumber%16)*32*1024 + start
	case 0x81: // PE (inputs)
		return 512*1024 + start
	case 0x82: // PA (outputs)
		return 576*1024 + start
	case 0x83: // MK (markers)
		return 640*1024 + start
	case 0x1C: // CT (counters)
		return 704*1024 + start
	case 0x1D: // TM (timers)
		return 768*1024 + start
	default:
		return 832*1024 + start
	}
}

// buildErrorResponse creates a minimal error response
func (m *MockClientHandler) buildErrorResponse() []byte {
	resp := make([]byte, 22)
	// TPKT
	resp[0] = 0x03
	binary.BigEndian.PutUint16(resp[2:], 22)
	// COTP
	resp[4] = 0x02
	resp[5] = 0xF0
	resp[6] = 0x80
	// S7 header
	resp[7] = 0x32
	resp[8] = 0x03 // Ack Data
	// Error indication
	resp[17] = 0x01 // Error class
	resp[18] = 0x01 // Error code
	resp[21] = 0x01 // Non-0xFF = error
	return resp
}

func errInvalidRequest() error {
	return &S7Error{High: 0x00, Low: 0x01}
}
