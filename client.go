package gos7

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

const (
	// Area ID - 区域标识符
	s7areape = 0x81 // Process Inputs / 过程输入区 (I)
	s7areapa = 0x82 // Process Outputs / 过程输出区 (Q)
	s7areamk = 0x83 // Merkers / 标志位区 (M)
	s7areadb = 0x84 // Data Block / 数据块 (DB)
	s7areact = 0x1C // Counters / 计数器区 (C)
	s7areatm = 0x1D // Timers / 定时器区 (T)

	// Word Length - 数据类型长度
	s7wlbit     = 0x01 // Bit (inside a word) / 位
	s7wlbyte    = 0x02 // Byte (8 bit) / 字节
	s7wlChar    = 0x03 // Char / 字符
	s7wlword    = 0x04 // Word (16 bit) / 字
	s7wlint     = 0x05 // Int / 整数
	s7wldword   = 0x06 // Double Word (32 bit) / 双字
	s7wldint    = 0x07 // DInt / 双整数
	s7wlreal    = 0x08 // Real (32 bit float) / 浮点数
	s7wlcounter = 0x1C // Counter (16 bit) / 计数器
	s7wltimer   = 0x1D // Timer (16 bit) / 定时器
)

// Exported area and word length constants for S7DataItem usage
const (
	// Area ID (exported) - 区域标识符
	S7AreaPE = s7areape // Process Inputs / 过程输入区 (I)
	S7AreaPA = s7areapa // Process Outputs / 过程输出区 (Q)
	S7AreaMK = s7areamk // Merkers / 标志位区 (M)
	S7AreaDB = s7areadb // Data Block / 数据块 (DB)
	S7AreaCT = s7areact // Counters / 计数器区 (C)
	S7AreaTM = s7areatm // Timers / 定时器区 (T)

	// Word Length (exported) - 数据类型长度
	S7WLBit     = s7wlbit     // Bit (inside a word) / 位
	S7WLByte    = s7wlbyte    // Byte (8 bit) / 字节
	S7WLChar    = s7wlChar    // Char / 字符
	S7WLWord    = s7wlword    // Word (16 bit) / 字
	S7WLInt     = s7wlint     // Int / 整数
	S7WLDWord   = s7wldword   // Double Word (32 bit) / 双字
	S7WLDInt    = s7wldint    // DInt / 双整数
	S7WLReal    = s7wlreal    // Real (32 bit float) / 浮点数
	S7WLCounter = s7wlcounter // Counter (16 bit) / 计数器
	S7WLTimer   = s7wltimer   // Timer (16 bit) / 定时器

	// PLC Status - PLC状态
	s7CpuStatusUnknown = 0 // Unknown / 未知
	s7CpuStatusRun     = 8 // Running / 运行中
	s7CpuStatusStop    = 4 // Stopped / 停止

	// PLC Status (exported) - PLC状态
	S7CpuStatusUnknown = s7CpuStatusUnknown // Unknown / 未知
	S7CpuStatusRun     = s7CpuStatusRun     // Running / 运行中
	S7CpuStatusStop    = s7CpuStatusStop    // Stopped / 停止

	// Header Size - 协议头大小
	sizeHeaderRead  int = 31 // Header Size when Reading / 读取操作的协议头大小
	sizeHeaderWrite int = 35 // Header Size when Writing / 写入操作的协议头大小

	// Result Transport Size - 结果传输类型
	tsResBit   = 3 // Bit result / 位结果
	tsResByte  = 4 // Byte result / 字节结果
	tsResInt   = 5 // Int result / 整数结果
	tsResReal  = 7 // Real result / 浮点结果
	tsResOctet = 9 // Octet result / 八位组结果
)

// PDULength variable stores PDU length after connection
// PDULength 变量用于存储连接后的PDU长度

// PDUProvider is an optional interface that provides the negotiated PDU length.
// Implementations that are not *TCPClientHandler should implement this interface
// to allow the client to determine PDU boundaries.
//
// PDUProvider 是一个可选接口，提供协商后的 PDU 长度。
// 非 *TCPClientHandler 的实现应实现此接口，以允许客户端确定 PDU 边界。
type PDUProvider interface {
	GetPDULength() int
}

// ClientHandler is the interface that groups the Packager and Transporter methods.
// ClientHandler 接口组合了 Packager 和 Transporter 方法
type ClientHandler interface {
	Packager
	Transporter
}

// client is the internal implementation of the Client interface
// client 是 Client 接口的内部实现
type client struct {
	packager    Packager
	transporter Transporter
}

// getPDU retrieves the negotiated PDU length from the transporter.
// It first checks for *TCPClientHandler (the standard implementation),
// then falls back to the PDUProvider interface for custom implementations.
//
// getPDU 从传输器获取协商后的 PDU 长度。
// 它首先检查 *TCPClientHandler（标准实现），
// 然后回退到 PDUProvider 接口（自定义实现）。
func (mb *client) getPDU() int {
	if tt, ok := interface{}(mb.transporter).(*TCPClientHandler); ok {
		return tt.PDULength
	}
	if pp, ok := interface{}(mb.transporter).(PDUProvider); ok {
		return pp.GetPDULength()
	}
	return 480 // default PDU size
}

// NewClient creates a new S7 client with given backend handler.
// NewClient 创建一个新的 S7 客户端，使用给定的后端处理器
func NewClient(handler ClientHandler) Client {
	return &client{packager: handler, transporter: handler}
}

// NewClient2 creates a new S7 client with separate packager and transporter.
// NewClient2 创建一个新的 S7 客户端，使用独立的 packager 和 transporter
func NewClient2(packager Packager, transporter Transporter) Client {
	return &client{packager: packager, transporter: transporter}
}

// AGReadDB implements the Client interface - Read Data Block from PLC
// AGReadDB 实现 Client 接口 - 从PLC读取数据块
func (mb *client) AGReadDB(dbnumber int, start int, size int, buffer []byte) (err error) {
	return mb.readArea(s7areadb, dbnumber, start, size, s7wlbyte, buffer)
}

// AGWriteDB implements the Client interface - Write Data Block to PLC
// AGWriteDB 实现 Client 接口 - 向PLC写入数据块
func (mb *client) AGWriteDB(dbNumber int, start int, size int, buffer []byte) (err error) {
	return mb.writeArea(s7areadb, dbNumber, start, size, s7wlbyte, buffer)
}

// AGReadMB implements the Client interface - Read Merkers (Memory Bits) from PLC
// AGReadMB 实现 Client 接口 - 从PLC读取标志位
func (mb *client) AGReadMB(start int, size int, buffer []byte) (err error) {
	return mb.readArea(s7areamk, 0, start, size, s7wlbyte, buffer)
}

// AGWriteMB implements the Client interface - Write Merkers (Memory Bits) to PLC
// AGWriteMB 实现 Client 接口 - 向PLC写入标志位
func (mb *client) AGWriteMB(start int, size int, buffer []byte) (err error) {
	return mb.writeArea(s7areamk, 0, start, size, s7wlbyte, buffer)
}

// AGReadEB implements the Client interface - Read Process Inputs (E) from PLC
// AGReadEB 实现 Client 接口 - 从PLC读取输入过程映像
func (mb *client) AGReadEB(start int, size int, buffer []byte) (err error) {
	return mb.readArea(s7areape, 0, start, size, s7wlbyte, buffer)
}

// AGWriteEB implements the Client interface - Write Process Inputs (E) to PLC
// AGWriteEB 实现 Client 接口 - 向PLC写入输入过程映像
func (mb *client) AGWriteEB(start int, size int, buffer []byte) (err error) {
	return mb.writeArea(s7areape, 0, start, size, s7wlbyte, buffer)
}

// AGReadAB implements the Client interface - Read Process Outputs (A) from PLC
// AGReadAB 实现 Client 接口 - 从PLC读取输出过程映像
func (mb *client) AGReadAB(start int, size int, buffer []byte) (err error) {
	return mb.readArea(s7areapa, 0, start, size, s7wlbyte, buffer)
}

// AGWriteAB implements the Client interface - Write Process Outputs (A) to PLC
// AGWriteAB 实现 Client 接口 - 向PLC写入输出过程映像
func (mb *client) AGWriteAB(start int, size int, buffer []byte) (err error) {
	return mb.writeArea(s7areapa, 0, start, size, s7wlbyte, buffer)
}

// AGReadTM implements the Client interface - Read Timer values from PLC
// AGReadTM 实现 Client 接口 - 从PLC读取定时器值
func (mb *client) AGReadTM(start int, amount int, buffer []byte) (err error) {
	sbuffer := make([]byte, amount*2)
	err = mb.readArea(s7areatm, 0, start, amount, s7wltimer, sbuffer)
	if err == nil {
		for c := 0; c < amount; c++ {
			buffer[c] = byte(uint16(sbuffer[c*2+1])<<8 + uint16(sbuffer[c*2]))
		}
	}
	return err
}

// AGWriteTM implements the Client interface - Write Timer values to PLC
// AGWriteTM 实现 Client 接口 - 向PLC写入定时器值
func (mb *client) AGWriteTM(start int, amount int, buffer []byte) (err error) {
	sbuffer := make([]byte, amount*2)
	for c := 0; c < amount; c++ {
		sbuffer[c*2+1] = byte((uint(buffer[c]) & uint(0xFF00)) >> 8)
		sbuffer[c*2] = byte(buffer[c] & 0x00FF)
	}
	err = mb.writeArea(s7areatm, 0, start, amount, s7wltimer, sbuffer)
	return err
}

// AGReadCT implements the Client interface - Read Counter values from PLC
// AGReadCT 实现 Client 接口 - 从PLC读取计数器值
func (mb *client) AGReadCT(start int, amount int, buffer []byte) (err error) {
	sbuffer := make([]byte, amount*2)
	err = mb.readArea(s7areact, 0, start, amount, s7wlcounter, sbuffer)
	if err == nil {
		for c := 0; c < amount; c++ {
			buffer[c] = byte(uint(sbuffer[c*2+1])<<8 + uint(sbuffer[c*2]))
		}
	}
	return err
}

// AGWriteCT implements the Client interface - Write Counter values to PLC
// AGWriteCT 实现 Client 接口 - 向PLC写入计数器值
func (mb *client) AGWriteCT(start int, amount int, buffer []byte) (err error) {
	sbuffer := make([]byte, amount*2)
	for c := 0; c < amount; c++ {
		sbuffer[c*2+1] = byte((uint(buffer[c]) & uint(0xFF00)) >> 8)
		sbuffer[c*2] = byte(buffer[c] & 0x00FF)
	}
	err = mb.writeArea(s7areact, 0, start, amount, s7wlcounter, sbuffer)
	return err
}

// readArea reads data from a generic PLC memory area and stores result in buffer
// readArea 从PLC通用内存区域读取数据并将结果存储到缓冲区
// Parameters:
//
//	area - Memory area type (s7areadb/s7areamk/s7areape/s7areapa/s7areact/s7areatm)
//	area - 内存区域类型
//	dbNumber - Data block number (only used for DB area)
//	dbNumber - 数据块编号（仅用于DB区域）
//	start - Start address within the area
//	start - 区域内的起始地址
//	amount - Number of elements to read
//	amount - 要读取的元素数量
//	wordLen - Data type (s7wlbit/s7wlbyte/s7wlword/s7wldword/s7wlreal/s7wlcounter/s7wltimer)
//	wordLen - 数据类型
//	buffer - Output buffer to store the read data
//	buffer - 存储读取数据的输出缓冲区
func (mb *client) readArea(area int, dbNumber int, start int, amount int, wordLen int, buffer []byte) (err error) {
	var address, numElements, maxElements, totElements, sizeRequested int
	offset := 0
	wordSize := 1

	// Adjust word length for counters and timers
	// 为计数器和定时器调整数据类型
	if area == s7areact {
		wordLen = s7wlcounter
	}
	if area == s7areatm {
		wordLen = s7wltimer
	}

	// Calculate word size based on data type
	// 根据数据类型计算字大小
	wordSize = dataSizeByte(wordLen)
	if wordSize == 0 {
		return fmt.Errorf("%s", ErrorText(errIsoInvalidDataSize))
	}

	// Special handling for bit operations (only 1 bit at a time)
	// 位操作的特殊处理（一次只能传输1位）
	if wordLen == s7wlbit {
		amount = 1 // Only 1 bit can be transferred at time
	} else {
		// Convert amount to bytes for non-counter/timer types
		// 对于非计数器/定时器类型，将数量转换为字节
		if wordLen != s7wlcounter && wordLen != s7wltimer {
			amount = amount * wordSize
			wordSize = 1
			wordLen = s7wlbyte
		}
	}

	// Get PDU length from transporter
	// 从传输器获取PDU长度
	pduLength := mb.getPDU()

	// Calculate max elements per request based on PDU size
	// 根据PDU大小计算每次请求的最大元素数
	maxElements = (pduLength - 18) / wordSize // 18 = Reply telegram header
	totElements = amount

	// Process in chunks if amount exceeds maxElements per request
	// 如果数量超过每次请求的最大元素数，则分块处理
	for totElements > 0 && err == nil {
		numElements = totElements
		if numElements > maxElements {
			numElements = maxElements
		}

		sizeRequested = numElements * wordSize

		// Setup the request telegram
		// 设置请求报文
		requestData := make([]byte, sizeHeaderRead)
		copy(requestData[0:], s7ReadWriteTelegram[0:])
		request := NewProtocolDataUnit(requestData)

		// Set area type
		// 设置区域类型
		request.Data[27] = byte(area)

		// Set DB number if accessing DB area
		// 如果访问DB区域，设置DB编号
		if area == s7areadb {
			binary.BigEndian.PutUint16(request.Data[25:], uint16(dbNumber))
		}

		// Adjust address format based on data type
		// 根据数据类型调整地址格式
		if wordLen == s7wlbit || wordLen == s7wlcounter || wordLen == s7wltimer {
			address = start
			request.Data[22] = byte(wordLen)
		} else {
			address = start << 3 // Convert to bit address
		}

		// Set number of elements
		// 设置元素数量
		binary.BigEndian.PutUint16(request.Data[23:], uint16(numElements))

		// Encode address into 3 bytes
		// 将地址编码为3字节
		request.Data[30] = byte(address & 0x0FF)
		address = address >> 8
		request.Data[29] = byte(address & 0x0FF)
		address = address >> 8
		request.Data[28] = byte(address & 0x0FF)

		// Send request and get response
		// 发送请求并获取响应
		var response *ProtocolDataUnit
		response, sendError := mb.send(&request)
		err = sendError

		// Process response
		// 处理响应
		if err == nil {
			if size := len(response.Data); size < 25 {
				err = fmt.Errorf("%s'%v'", ErrorText(errIsoInvalidDataSize), len(response.Data))
			} else {
				if response.Data[21] != 0xFF {
					// Check for CPU error
					// 检查CPU错误
					err = fmt.Errorf("%s", ErrorText(CPUError(uint(response.Data[21]))))
				} else {
					// Copy response data to buffer
					// 将响应数据复制到缓冲区
					copy(buffer[offset:offset+sizeRequested], response.Data[25:25+sizeRequested])
					offset += sizeRequested
				}
			}
		}

		// Update counters for next iteration
		// 更新计数器以进行下一次迭代
		totElements -= numElements
		start += numElements * wordSize
	}
	return
}

// writeArea writes data to a generic PLC memory area
// writeArea 向PLC通用内存区域写入数据
// Parameters:
//
//	area - Memory area type (s7areadb/s7areamk/s7areape/s7areapa/s7areact/s7areatm)
//	area - 内存区域类型
//	dbnumber - Data block number (only used for DB area)
//	dbnumber - 数据块编号（仅用于DB区域）
//	start - Start address within the area
//	start - 区域内的起始地址
//	amount - Number of elements to write
//	amount - 要写入的元素数量
//	wordlen - Data type (s7wlbit/s7wlbyte/s7wlword/s7wldword/s7wlreal/s7wlcounter/s7wltimer)
//	wordlen - 数据类型
//	buffer - Input buffer containing data to write
//	buffer - 包含要写入数据的输入缓冲区
func (mb *client) writeArea(area int, dbnumber int, start int, amount int, wordlen int, buffer []byte) (err error) {
	var address, numElements, maxElements, totElements, dataSize, isoSize, length int
	offset := 0
	wordSize := 1

	// Adjust word length for counters and timers
	// 为计数器和定时器调整数据类型
	if area == s7areact {
		wordlen = s7wlcounter
	}
	if area == s7areatm {
		wordlen = s7wltimer
	}

	// Calculate word size based on data type
	// 根据数据类型计算字大小
	wordSize = dataSizeByte(wordlen)
	if wordSize == 0 {
		return fmt.Errorf("%s", ErrorText(errIsoInvalidDataSize))
	}

	// Special handling for bit operations (only 1 bit at a time)
	// 位操作的特殊处理（一次只能传输1位）
	if wordlen == s7wlbit {
		amount = 1 // Only 1 bit can be transferred at time
	} else {
		// Convert amount to bytes for non-counter/timer types
		// 对于非计数器/定时器类型，将数量转换为字节
		if wordlen != s7wlcounter && wordlen != s7wltimer {
			amount = amount * wordSize
			wordSize = 1
			wordlen = s7wlbyte
		}
	}

	// Get PDU length from transporter
	// 从传输器获取PDU长度
	pduLength := mb.getPDU()
	maxElements = (pduLength - 35) / wordSize // 35 = Reply telegram header
	totElements = amount

	// Process in chunks if amount exceeds maxElements per request
	// 如果数量超过每次请求的最大元素数，则分块处理
	for totElements > 0 && err == nil {
		numElements = totElements
		if numElements > maxElements {
			numElements = maxElements
		}
		dataSize = numElements * wordSize
		isoSize = sizeHeaderWrite + dataSize

		// Setup the request telegram
		// 设置请求报文
		requestData := make([]byte, sizeHeaderWrite)
		copy(requestData[0:], s7ReadWriteTelegram[0:])

		request := NewProtocolDataUnit(requestData)

		// Set whole telegram size
		// 设置整个报文大小
		binary.BigEndian.PutUint16(request.Data[2:], uint16(isoSize))

		// Set data length (data + 4 bytes header)
		// 设置数据长度（数据+4字节头部）
		length = dataSize + 4
		binary.BigEndian.PutUint16(request.Data[15:], uint16(length))

		// Set function code: 0x05 = Write
		// 设置功能码：0x05 = 写入
		request.Data[17] = byte(0x05)

		// Set area type
		// 设置区域类型
		request.Data[27] = byte(area)
		if area == s7areadb {
			binary.BigEndian.PutUint16(request.Data[25:], uint16(dbnumber))
		}

		// Adjust address format based on data type
		// 根据数据类型调整地址格式
		if wordlen == s7wlbit || wordlen == s7wlcounter || wordlen == s7wltimer {
			address = start
			length = dataSize
			request.Data[22] = byte(wordlen)
		} else {
			address = start << 3
			length = dataSize << 3
		}

		// Set number of elements
		// 设置元素数量
		binary.BigEndian.PutUint16(request.Data[23:], uint16(numElements))

		// Encode address into 3 bytes
		// 将地址编码为3字节
		request.Data[30] = byte(address & 0x0FF)
		address = address >> 8
		request.Data[29] = byte(address & 0x0FF)
		address = address >> 8
		request.Data[28] = byte(address & 0x0FF)

		// Set transport size based on data type
		// 根据数据类型设置传输大小
		switch wordlen {
		case s7wlbit:
			request.Data[32] = tsResBit
			break
		case s7wlcounter:
		case s7wltimer:
			request.Data[32] = tsResOctet
			break
		default:
			request.Data[32] = tsResByte // byte/word/dword etc.
			break
		}

		// Set data length field
		// 设置数据长度字段
		binary.BigEndian.PutUint16(request.Data[33:], uint16(length))

		// Append data to request
		// 将数据附加到请求中
		request.Data = append(request.Data[:35], append(buffer[offset:offset+dataSize], request.Data[35:]...)...)

		// Send request and get response
		// 发送请求并获取响应
		response, sendError := mb.send(&request)
		err = sendError

		// Check response
		// 检查响应
		if err == nil {
			if length = len(response.Data); length == 22 {
				if response.Data[21] != byte(0xFF) {
					err = fmt.Errorf("%s", ErrorText(CPUError(uint(response.Data[21]))))
				}
			} else {
				err = fmt.Errorf("%s", ErrorText(errIsoInvalidPDU))
			}
		}

		// Update counters for next iteration
		// 更新计数器以进行下一次迭代
		offset += dataSize
		totElements -= numElements
		start += numElements * wordSize
	}
	return
}

// Read reads a PLC variable using S7 syntax and returns the value
// Read 使用S7语法读取PLC变量并返回值
// S7 syntax examples:
// - DB1.DBB0     - Data Block 1, Byte 0
// - DB1.DBW2     - Data Block 1, Word at offset 2
// - DB1.DBD4     - Data Block 1, DWord at offset 4
// - DB1.DBX0.0   - Data Block 1, Bit 0 of Byte 0
// - MB10        - Memory Bit 10
// - T5          - Timer 5
// - C10         - Counter 10
//
// Parameters:
//
//	variable - S7 syntax variable string
//	variable - S7语法变量字符串
//	buffer - Buffer for intermediate data storage
//	buffer - 中间数据存储缓冲区
//
// Returns:
//
//	value - Read value as interface{}
//	value - 读取的值
//	err - Error if any
//	err - 错误信息
func (mb *client) Read(variable string, buffer []byte) (value interface{}, err error) {
	variable = strings.ToUpper(variable)              // Convert to uppercase / 转换为大写
	variable = strings.Replace(variable, " ", "", -1) // Remove spaces / 移除空格

	if variable == "" {
		err = fmt.Errorf("input variable is empty, variable should be S7 syntax")
		return
	}

	// Parse S7 variable syntax
	// 解析S7变量语法
	switch valueArea := variable[0:2]; valueArea {
	case "EB": // Input byte / 输入字节
	case "EW": // Input word / 输入字
	case "ED": // Input double-word / 输入双字
	case "AB": // Output byte / 输出字节
	case "AW": // Output word / 输出字
	case "AD": // Output double-word / 输出双字
	case "MB": // Memory byte / 标志位字节
	case "MW": // Memory word / 标志位字
	case "MD": // Memory double-word / 标志位双字
	case "DB": // Data Block / 数据块
		dbArray := strings.Split(variable, ".")
		if len(dbArray) < 2 {
			err = fmt.Errorf("DB Area read variable should not be empty")
			return
		}
		dbNo, _ := strconv.ParseInt(string(string(dbArray[0])[2:]), 10, 16)
		dbIndex, _ := strconv.ParseInt(string(string(dbArray[1])[3:]), 10, 16)
		dbType := string(dbArray[1])[0:3]

		switch dbType {
		case "DBB": // Byte / 字节
			err = mb.AGReadDB(int(dbNo), int(dbIndex), 1, buffer)
			value = buffer[0]
			return
		case "DBW": // Word / 字
			err = mb.AGReadDB(int(dbNo), int(dbIndex), 2, buffer)
			value = binary.BigEndian.Uint16(buffer[0:])
			return
		case "DBD": // Double Word / 双字
			err = mb.AGReadDB(int(dbNo), int(dbIndex), 4, buffer)
			value = binary.BigEndian.Uint32(buffer[0:])
			return
		case "DBX": // Bit / 位
			mBit, _ := strconv.ParseInt(string(string(dbArray[2])[0:]), 10, 16)
			if mBit > 7 || mBit < 0 {
				err = fmt.Errorf("DB read bit is invalid")
				return
			}
			err = mb.AGReadDB(int(dbNo), int(dbIndex), 1, buffer)
			mask := []byte{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80}
			value = buffer[0] & mask[mBit]
			return
		default:
			err = fmt.Errorf("error when parsing DB type")
			return
		}
	default:
		switch otherArea := variable[0:1]; otherArea {
		case "E": // Input / 输入
		case "I": // Input / 输入
		case "A": // Output / 输出
		case "0": // Output / 输出
		case "M": // Memory / 标志位
		case "T": // Timer / 定时器
			startByte, _ := strconv.ParseInt(string(variable[1:]), 10, 16)
			err = mb.AGReadTM(int(startByte), 1, buffer)
			if err != nil {
				return
			}
			helper := Helper{}
			helper.GetValueAt(buffer, 0, value)
			return
		case "Z": // Counter / 计数器
		case "C": // Counter / 计数器
			startByte, _ := strconv.ParseInt(string(variable[1:]), 10, 16)
			err = mb.AGReadCT(int(startByte), 1, buffer)
			if err != nil {
				return
			}
			helper := Helper{}
			helper.GetValueAt(buffer, 0, value)
			return
		default:
			err = fmt.Errorf("error when parsing area")
			return
		}
	}
	return
}

// send sends a PDU request to the PLC and returns the response
// send 向PLC发送PDU请求并返回响应
// This method handles the low-level communication:
// 1. Sends the request through the transporter
// 2. Verifies the response using the packager
// 3. Creates and returns a ProtocolDataUnit response
//
// 该方法处理底层通信：
// 1. 通过传输器发送请求
// 2. 使用打包器验证响应
// 3. 创建并返回ProtocolDataUnit响应
func (mb *client) send(request *ProtocolDataUnit) (response *ProtocolDataUnit, err error) {
	// Send request through transporter
	// 通过传输器发送请求
	dataResponse, err := mb.transporter.Send(request.Data)
	if err != nil {
		return
	}

	// Verify response integrity
	// 验证响应完整性
	if err = mb.packager.Verify(request.Data, dataResponse); err != nil {
		return
	}

	// Check for empty response
	// 检查空响应
	if dataResponse == nil || len(dataResponse) == 0 {
		err = fmt.Errorf("s7: response data is empty")
		return
	}

	// Create response PDU
	// 创建响应PDU
	response = &ProtocolDataUnit{
		Data: dataResponse,
	}

	// Check for protocol errors
	// 检查协议错误
	err = responseError(response)
	return response, err
}

// responseError extracts and returns S7 protocol error from response
// responseError 从响应中提取并返回S7协议错误
func responseError(response *ProtocolDataUnit) error {
	s7Error := &S7Error{}
	if response.Data != nil && len(response.Data) > 0 {
		switch int(response.Data[1]) {
		case 1:
		case 7:
			s7Error.High = response.Data[2]
			s7Error.Low = response.Data[3]
			break
		case 2:
		case 3:
			s7Error.High = response.Data[10]
			s7Error.Low = response.Data[11]
			break
		default:
			return nil
		}
	}
	return s7Error
}

// dataSize to number of byte accordingly
func dataSizeByte(wordLength int) int {
	switch wordLength {
	case s7wlbit:
		return 1
	case s7wlbyte:
		return 1
	case s7wlChar:
		return 1
	case s7wlword:
		return 2
	case s7wlint:
		return 2
	case s7wlcounter:
		return 2
	case s7wltimer:
		return 2
	case s7wldword:
		return 4
	case s7wldint:
		return 4
	case s7wlreal:
		return 4
	default:
		return 0
	}

}

// add api WriteArea
func (mb *client) WriteArea(area int, dbnumber int, start int, amount int, wordlen int, buffer []byte) (err error) {
	return mb.writeArea(area, dbnumber, start, amount, wordlen, buffer)
}

// add api ReadArea
func (mb *client) ReadArea(area int, dbNumber int, start int, amount int, wordLen int, buffer []byte) (err error) {
	return mb.readArea(area, dbNumber, start, amount, wordLen, buffer)
}

// ReadAreas reads multiple S7 data items from PLC in batches.
// It splits the items into batches of up to 20 items and performs
// multi-read operations for each batch.
//
// ReadAreas 从PLC批量读取多个S7数据项。
// 将数据项按最多20个一批分组，对每批执行多区域读取操作。
//
// Parameters:
//
//	items - Slice of S7DataItem specifying the areas and addresses to read
//	items - 指定要读取的区域和地址的S7DataItem切片
func (mb *client) ReadAreas(items []S7DataItem) (err error) {
	itemsCount := len(items)
	if itemsCount == 0 {
		return
	}
	maxBatchSize := 20
	for i := 0; i < itemsCount; i += maxBatchSize {
		end := i + maxBatchSize
		if end > itemsCount {
			end = itemsCount
		}
		batch := items[i:end]
		err = mb.AGReadMulti(batch, len(batch))
		if err != nil {
			return
		}
	}
	return
}

// WriteAreas writes multiple S7 data items to PLC in batches.
// It splits the items into batches of up to 20 items and performs
// multi-write operations for each batch.
//
// WriteAreas 向PLC批量写入多个S7数据项。
// 将数据项按最多20个一批分组，对每批执行多区域写入操作。
//
// Parameters:
//
//	items - Slice of S7DataItem specifying the areas, addresses and data to write
//	items - 指定要写入的区域、地址和数据的S7DataItem切片
func (mb *client) WriteAreas(items []S7DataItem) (err error) {
	itemsCount := len(items)
	if itemsCount == 0 {
		return
	}
	maxBatchSize := 20
	for i := 0; i < itemsCount; i += maxBatchSize {
		end := i + maxBatchSize
		if end > itemsCount {
			end = itemsCount
		}
		batch := items[i:end]
		err = mb.AGWriteMulti(batch, len(batch))
		if err != nil {
			return
		}
	}
	return
}
