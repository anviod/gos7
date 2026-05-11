package gos7

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
	"time"
)

// Client is the main interface for S7 PLC communication
// Client 是S7 PLC通信的主接口
// It provides methods for reading/writing data from/to PLC memory areas,
// as well as PLC control and system information functions.
// 它提供了从PLC内存区域读写数据的方法，以及PLC控制和系统信息功能。
type Client interface {
	/*************** AG API (Automatisationsgerät / Automation Device) ***************/
	// AGReadDB reads data block from PLC
	// AGReadDB 从PLC读取数据块
	AGReadDB(dbNumber int, start int, size int, buffer []byte) (err error)

	// AGWriteDB writes data block to PLC
	// AGWriteDB 向PLC写入数据块
	AGWriteDB(dbNumber int, start int, size int, buffer []byte) (err error)

	// AGReadMB reads Merkers (Memory Bits) from PLC
	// AGReadMB 从PLC读取标志位
	AGReadMB(start int, size int, buffer []byte) (err error)

	// AGWriteMB writes Merkers (Memory Bits) to PLC
	// AGWriteMB 向PLC写入标志位
	AGWriteMB(start int, size int, buffer []byte) (err error)

	// AGReadEB reads Process Inputs (E) from PLC
	// AGReadEB 从PLC读取输入过程映像
	AGReadEB(start int, size int, buffer []byte) (err error)

	// AGWriteEB writes Process Inputs (E) to PLC
	// AGWriteEB 向PLC写入输入过程映像
	AGWriteEB(start int, size int, buffer []byte) (err error)

	// AGReadAB reads Process Outputs (A) from PLC
	// AGReadAB 从PLC读取输出过程映像
	AGReadAB(start int, size int, buffer []byte) (err error)

	// AGWriteAB writes Process Outputs (A) to PLC
	// AGWriteAB 向PLC写入输出过程映像
	AGWriteAB(start int, size int, buffer []byte) (err error)

	// AGReadTM reads Timer values from PLC
	// AGReadTM 从PLC读取定时器值
	AGReadTM(start int, size int, buffer []byte) (err error)

	// AGWriteTM writes Timer values to PLC
	// AGWriteTM 向PLC写入定时器值
	AGWriteTM(start int, size int, buffer []byte) (err error)

	// AGReadCT reads Counter values from PLC
	// AGReadCT 从PLC读取计数器值
	AGReadCT(start int, size int, buffer []byte) (err error)

	// AGWriteCT writes Counter values to PLC
	// AGWriteCT 向PLC写入计数器值
	AGWriteCT(start int, size int, buffer []byte) (err error)

	// AGReadMulti reads multiple data items in a single request
	// AGReadMulti 在单个请求中读取多个数据项
	AGReadMulti(dataItems []S7DataItem, itemsCount int) (err error)

	// AGWriteMulti writes multiple data items in a single request
	// AGWriteMulti 在单个请求中写入多个数据项
	AGWriteMulti(dataItems []S7DataItem, itemsCount int) (err error)

	// DBFill fills a data block with a specific byte value
	// DBFill 使用指定字节值填充数据块
	DBFill(dbnumber int, fillchar int) error

	// DBGet reads an entire data block
	// DBGet 读取整个数据块
	DBGet(dbnumber int, usrdata []byte, size int) error

	// Read reads a variable using S7 syntax (e.g., "DB1.DBB0", "MB10")
	// Read 使用S7语法读取变量（例如 "DB1.DBB0", "MB10"）
	Read(variable string, buffer []byte) (value interface{}, err error)

	// GetAgBlockInfo retrieves block information from AG area
	// GetAgBlockInfo 从AG区域获取块信息
	GetAgBlockInfo(blocktype int, blocknum int) (info S7BlockInfo, err error)

	/*************** PG API (Programmiergerät / Programming Device) ***************/
	// PLCHotStart performs a hot start on the PLC CPU
	// PLCHotStart 对PLC CPU执行热启动
	PLCHotStart() error

	// PLCColdStart performs a cold start on the PLC CPU
	// PLCColdStart 对PLC CPU执行冷启动
	PLCColdStart() error

	// PLCStop stops the PLC CPU
	// PLCStop 停止PLC CPU
	PLCStop() error

	// PLCGetStatus returns the CPU status (running/stopped)
	// PLCGetStatus 返回CPU状态（运行/停止）
	PLCGetStatus() (status int, err error)

	// PGListBlocks lists all blocks in PLC
	// PGListBlocks 列出PLC中所有块
	PGListBlocks() (list S7BlocksList, err error)

	// SetSessionPassword sets the session password for PLC
	// SetSessionPassword 设置PLC会话密码
	SetSessionPassword(password string) error

	// ClearSessionPassword clears the session password
	// ClearSessionPassword 清除会话密码
	ClearSessionPassword() error

	// GetProtection returns CPU protection level information
	// GetProtection 返回CPU保护级别信息
	GetProtection() (protection S7Protection, err error)

	// GetOrderCode returns CPU order code information
	// GetOrderCode 返回CPU订货号信息
	GetOrderCode() (info S7OrderCode, err error)

	// GetCPUInfo returns CPU information
	// GetCPUInfo 返回CPU信息
	GetCPUInfo() (info S7CpuInfo, err error)

	// GetCPInfo returns CP (Communication Processor) information
	// GetCPInfo 返回CP（通信处理器）信息
	GetCPInfo() (info S7CpInfo, err error)

	// PGClockRead reads the PLC clock
	// PGClockRead 读取PLC时钟
	PGClockRead(datetime time.Time) error

	// PGClockWrite writes to the PLC clock
	// PGClockWrite 写入PLC时钟
	PGClockWrite() (dt time.Time, err error)

	/*************** Generic Area Operations ***************/
	// WriteArea writes to a generic memory area
	// WriteArea 向通用内存区域写入数据
	WriteArea(area int, dbnumber int, start int, amount int, wordlen int, buffer []byte) (err error)

	// ReadArea reads from a generic memory area
	// ReadArea 从通用内存区域读取数据
	ReadArea(area int, dbNumber int, start int, amount int, wordLen int, buffer []byte) (err error)

	// ReadAreas batch reads multiple data items with automatic batching
	// ReadAreas 批量读取多个数据项（自动分批处理）
	ReadAreas(items []S7DataItem) (err error)

	// WriteAreas batch writes multiple data items with automatic batching
	// WriteAreas 批量写入多个数据项（自动分批处理）
	WriteAreas(items []S7DataItem) (err error)
}
