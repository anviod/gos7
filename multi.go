package gos7

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
	"encoding/binary"
	"fmt"
)

// S7DataItem represents a single data item for multiple read/write operations
// S7DataItem 表示多读写操作中的单个数据项
// Fields:
//   Area - Memory area type (s7areadb/s7areamk/s7areape/s7areapa/s7areact/s7areatm)
//   Area - 内存区域类型
//   WordLen - Data type (s7wlbit/s7wlbyte/s7wlword/s7wldword/s7wlreal/s7wlcounter/s7wltimer)
//   WordLen - 数据类型
//   DBNumber - Data block number (only for DB area)
//   DBNumber - 数据块编号（仅用于DB区域）
//   Start - Start address within the area
//   Start - 区域内的起始地址
//   Bit - Bit position (only for bit operations)
//   Bit - 位位置（仅用于位操作）
//   Amount - Number of elements to read/write
//   Amount - 要读写的元素数量
//   Data - Data buffer for input/output
//   Data - 输入/输出数据缓冲区
//   Error - Error message if operation fails
//   Error - 操作失败时的错误信息
type S7DataItem struct {
	Area     int
	WordLen  int
	DBNumber int
	Start    int
	Bit      int
	Amount   int
	Data     []byte
	Error    string
}

// AGWriteMulti writes multiple data items to PLC in a single request
// AGWriteMulti 在单个请求中向PLC写入多个数据项
// This method significantly reduces network round-trips when writing multiple variables
// 此方法在写入多个变量时显著减少网络往返次数
// Maximum 20 items per request (S7 protocol limitation)
// 每次请求最多20个项目（S7协议限制）
func (mb *client) AGWriteMulti(dataItems []S7DataItem, itemsCount int) (err error) {
	// Validate item count (max 20 per request)
	// 验证项目数量（每次请求最多20个）
	if itemsCount > 20 {
		err = fmt.Errorf("%s", ErrorText(errCliTooManyItems))
		return
	}

	// Initialize telegram with write header
	// 使用写入头初始化报文
	s7Multi := make([]byte, len(s7MultiWriteHeaderTelegram))
	copy(s7Multi, s7MultiWriteHeaderTelegram)

	// Calculate and set parameter length
	// 计算并设置参数长度
	parLength := itemsCount*len(s7MultiWriteItemTelegram) + 2
	binary.BigEndian.PutUint16(s7Multi[13:], uint16(parLength))
	s7Multi[18] = byte(itemsCount)

	// Build parameter section for each item
	// 为每个项目构建参数部分
	offset := len(s7MultiWriteHeaderTelegram)
	for i := 0; i < itemsCount; i++ {
		s7ParamItem := make([]byte, len(s7MultiWriteItemTelegram))
		copy(s7ParamItem, s7MultiWriteItemTelegram)

		// Set word length, area, amount, and DB number
		// 设置字长度、区域、数量和DB编号
		s7ParamItem[3] = byte(dataItems[i].WordLen)
		s7ParamItem[8] = byte(dataItems[i].Area)
		binary.BigEndian.PutUint16(s7ParamItem[4:], uint16(dataItems[i].Amount))
		binary.BigEndian.PutUint16(s7ParamItem[6:], uint16(dataItems[i].DBNumber))

		// Calculate address based on data type
		// 根据数据类型计算地址
		var addr int
		if dataItems[i].WordLen == s7wlbit || dataItems[i].WordLen == s7wlcounter || dataItems[i].WordLen == s7wltimer {
			addr = dataItems[i].Start
		} else {
			addr = dataItems[i].Start * 8
		}

		// Encode address into 3 bytes
		// 将地址编码为3字节
		s7ParamItem[11] = byte(addr & 0x0FF)
		addr = addr >> 8
		s7ParamItem[10] = byte(addr & 0x0FF)
		addr = addr >> 8
		s7ParamItem[9] = byte(addr & 0x0FF)

		// Insert parameter item into telegram
		// 将参数项插入报文
		s7Multi = append(s7Multi[:offset], append(s7ParamItem, s7Multi[offset:]...)...)
		offset += len(s7ParamItem)
	}

	// Build data section for each item
	// 为每个项目构建数据部分
	dataLength := 0
	for i := 0; i < itemsCount; i++ {
		s7ItemWrite := make([]byte, 1024)
		s7ItemWrite[0] = 0
		itemDataSize := 0

		// Determine transport size and data size based on word length
		// 根据字长度确定传输大小和数据大小
		switch dataItems[i].WordLen {
		case s7wlbit:
			s7ItemWrite[1] = tsResBit
			itemDataSize = dataItems[i].Amount
			binary.BigEndian.PutUint16(s7ItemWrite[2:], uint16(itemDataSize))
		case s7wlcounter, s7wltimer:
			s7ItemWrite[1] = tsResOctet
			itemDataSize = dataItems[i].Amount * 2
			binary.BigEndian.PutUint16(s7ItemWrite[2:], uint16(itemDataSize))
		case s7wlreal:
			s7ItemWrite[1] = tsResReal
			itemDataSize = dataItems[i].Amount * dataSizeByte(dataItems[i].WordLen)
			binary.BigEndian.PutUint16(s7ItemWrite[2:], uint16(itemDataSize))
		default:
			s7ItemWrite[1] = tsResByte
			itemDataSize = dataItems[i].Amount * dataSizeByte(dataItems[i].WordLen)
			binary.BigEndian.PutUint16(s7ItemWrite[2:], uint16(itemDataSize*8))
		}

		// Copy data to item buffer
		// 将数据复制到项目缓冲区
		copy(s7ItemWrite[4:4+itemDataSize], dataItems[i].Data)

		// Pad to even size if necessary
		// 如果需要，填充为偶数大小
		if itemDataSize%2 != 0 {
			s7ItemWrite[itemDataSize+4] = 0
			itemDataSize++
		}

		// Append item data to telegram
		// 将项目数据附加到报文
		s7Multi = append(s7Multi, s7ItemWrite[0:itemDataSize+4]...)
		offset = offset + itemDataSize + 4
		dataLength = dataLength + itemDataSize + 4
	}

	// Check PDU size limit
	// 检查PDU大小限制
	tt, _ := interface{}(mb.transporter).(*TCPClientHandler)
	if offset > tt.PDULength {
		err = fmt.Errorf("%s", ErrorText(errCliSizeOverPDU))
		return
	}

	// Set telegram sizes
	// 设置报文大小
	binary.BigEndian.PutUint16(s7Multi[2:], uint16(offset))
	binary.BigEndian.PutUint16(s7Multi[15:], uint16(dataLength))

	// Create and send request
	// 创建并发送请求
	request := NewProtocolDataUnit(s7Multi)
	response, err := mb.send(&request)

	if err == nil {
		// Check global operation result
		// 检查全局操作结果
		cpuErr := CPUError(uint(binary.BigEndian.Uint16(response.Data[17:])))
		if cpuErr != 0 {
			err = fmt.Errorf("%s", ErrorText(cpuErr))
			return
		}

		// Verify items written count
		// 验证写入的项目数量
		if itemsWritten := int(response.Data[20]); itemsWritten != itemsCount || itemsWritten > 20 {
			err = fmt.Errorf("%s", ErrorText(errCliInvalidPlcAnswer))
			return
		}

		// Check individual item errors
		// 检查每个项目的错误
		for i := 0; i < itemsCount; i++ {
			if response.Data[i+21] == 0xFF {
				dataItems[i].Error = ""
			} else {
				dataItems[i].Error = ErrorText(CPUError(uint(response.Data[i+21])))
			}
		}
	}
	return
}

// AGReadMulti reads multiple data items from PLC in a single request
// AGReadMulti 在单个请求中从PLC读取多个数据项
// This method significantly reduces network round-trips when reading multiple variables
// 此方法在读取多个变量时显著减少网络往返次数
// Maximum 20 items per request (S7 protocol limitation)
// 每次请求最多20个项目（S7协议限制）
func (mb *client) AGReadMulti(dataItems []S7DataItem, itemsCount int) (err error) {
	// Validate item count (max 20 per request)
	// 验证项目数量（每次请求最多20个）
	if itemsCount > 20 {
		err = fmt.Errorf("%s", ErrorText(errCliTooManyItems))
		return
	}

	// Initialize telegram with read header
	// 使用读取头初始化报文
	s7Item := make([]byte, 12)
	s7Multi := make([]byte, len(s7MultiReadHeaderTelegram))
	copy(s7Multi, s7MultiReadHeaderTelegram)

	// Set header parameters
	// 设置头参数
	binary.BigEndian.PutUint16(s7Multi[13:], uint16(itemsCount*len(s7Item)+2))
	s7Multi[18] = byte(itemsCount)

	// Build item section
	// 构建项目部分
	offset := 19
	for i := 0; i < itemsCount; i++ {
		copy(s7Item, s7MultiReadItemTelegram)

		// Set word length, amount, DB number, and area
		// 设置字长度、数量、DB编号和区域
		s7Item[3] = byte(dataItems[i].WordLen)
		binary.BigEndian.PutUint16(s7Item[4:], uint16(dataItems[i].Amount))
		if dataItems[i].Area == s7areadb {
			binary.BigEndian.PutUint16(s7Item[6:], uint16(dataItems[i].DBNumber))
		}
		s7Item[8] = byte(dataItems[i].Area)

		// Calculate address based on data type
		// 根据数据类型计算地址
		var addr int
		if dataItems[i].WordLen == s7wlcounter || dataItems[i].WordLen == s7wltimer {
			addr = dataItems[i].Start
		} else if dataItems[i].WordLen == s7wlbit {
			addr = dataItems[i].Start << 3
			addr += dataItems[i].Bit
		} else {
			addr = dataItems[i].Start * 8
		}

		// Encode address into 3 bytes
		// 将地址编码为3字节
		s7Item[11] = byte(addr & 0x0FF)
		addr = addr >> 8
		s7Item[10] = byte(addr & 0x0FF)
		addr = addr >> 8
		s7Item[9] = byte(addr & 0x0FF)

		// Append item to telegram
		// 将项目附加到报文
		s7Multi = append(s7Multi, s7Item...)
		offset += len(s7Item)
	}

	// Check PDU size limit
	// 检查PDU大小限制
	tt, _ := interface{}(mb.transporter).(*TCPClientHandler)
	if offset > tt.PDULength {
		err = fmt.Errorf("%s", ErrorText(errCliSizeOverPDU))
		return
	}

	// Set telegram size
	// 设置报文大小
	binary.BigEndian.PutUint16(s7Multi[2:], uint16(offset))

	// Create and send request
	// 创建并发送请求
	request := NewProtocolDataUnit(s7Multi)
	response, err := mb.send(&request)
	if err != nil {
		return
	}

	// Validate response length
	// 验证响应长度
	resLength := len(response.Data)
	if resLength < 22 {
		err = fmt.Errorf("%s", ErrorText(errIsoInvalidPDU))
		return
	}

	// Check global operation result
	// 检查全局操作结果
	cpuErr := CPUError(uint(binary.BigEndian.Uint16(response.Data[17:])))
	if cpuErr != 0 {
		err = fmt.Errorf("%s", ErrorText(cpuErr))
		return
	}

	// Verify items read count
	// 验证读取的项目数量
	itemsRead := int(response.Data[20])
	s7ItemRead := make([]byte, 1024)
	if itemsRead != itemsCount || itemsRead > 20 {
		err = fmt.Errorf("%s", ErrorText(errCliInvalidPlcAnswer))
		return
	}

	// Parse response data
	// 解析响应数据
	offset = 21
	for i := 0; i < itemsCount; i++ {
		copy(s7ItemRead[0:resLength-offset], response.Data[offset:resLength])

		if s7ItemRead[0] == 255 {
			// Success - extract data
			// 成功 - 提取数据
			itemSize := int(binary.BigEndian.Uint16(s7ItemRead[2:]))
			item1 := s7ItemRead[1]

			// Adjust size for byte-based types
			// 为基于字节的类型调整大小
			if item1 != tsResOctet && item1 != tsResReal && item1 != tsResBit {
				itemSize = itemSize >> 3
			}

			// Copy data to item buffer
			// 将数据复制到项目缓冲区
			copy(dataItems[i].Data[0:], s7ItemRead[4:4+itemSize])
			dataItems[i].Error = ""

			// Account for padding
			// 考虑填充
			if itemSize%2 != 0 {
				itemSize++
			}
			offset = offset + 4 + itemSize
		} else {
			// Error - record error message
			// 错误 - 记录错误信息
			dataItems[i].Error = ErrorText(CPUError(uint(s7ItemRead[0])))
			offset += 4
		}
	}

	return
}
