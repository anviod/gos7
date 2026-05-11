# gos7
西门子 S7 协议的 Go 语言实现

概述
-------------------
多年来，商业和开源领域都有许多驱动程序/连接器支持连接到 S7 系列 PLC 设备。GoS7 填补了 S7 协议的空白，使用纯 Go（也称为 golang）实现。我们坚信低级通信应该使用接近二进制和内存的低级编程语言来实现。

支持的最低 Go 版本是 1.13。

功能特性
-------------------
AG (自动化设备):
*   读取/写入数据块 (DB) (已测试)
*   读取/写入标志位 (MB) (已测试)
*   读取/写入输入过程映像 (EB) (已测试)
*   读取/写入输出过程映像 (AB) (已测试)
*   读取/写入定时器 (TM) (已测试)
*   读取/写入计数器 (CT) (已测试)
*   多区域读取/写入 (已测试)
*   批量读取/写入区域 (ReadAreas/WriteAreas) (已测试)
*   获取块信息 (已测试)

PG (编程设备):
*   PLC 热启动/冷启动/停止
*   获取 PLC CPU 状态 (已测试)
*   列出 PLC 中可用的块 (已测试)
*   设置/清除会话密码
*   获取 CPU 保护级别和 CPU 订货号
*   获取 CPU/CP 信息 (已测试)
*   读取/写入 PLC 时钟

辅助工具:
*   对字节数组进行各种类型的读写操作：位/整数/字/双字/无符号整数等，实数，时间，计数器

支持的通信方式
-----------------
*   TCP
*   串口 (PPI, MPI) (开发中)

使用示例:
----------
以下是通过 TCP 连接 PLC 的简单示例：
```go
const (
	tcpDevice = "127.0.0.1"
	rack      = 0
	slot      = 2
)
// 创建 TCP 客户端处理器
handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
handler.Timeout = 200 * time.Second
handler.IdleTimeout = 200 * time.Second
handler.Logger = log.New(os.Stdout, "tcp: ", log.LstdFlags)
// 手动连接以便在一个连接会话中处理多个请求
handler.Connect()
defer handler.Close()
// 初始化客户端
client := gos7.NewClient(handler)
address := 2710
start := 8
size := 2
buffer := make([]byte, 255)
value := 100
// AGWriteDB 向 DB2710 写入值 100，从位置 8 开始，大小为 2（整数）
var helper gos7.Helper
helper.SetValueAt(buffer, 0, value)  
err := client.AGWriteDB(address, start, size, buffer)
buf := make([]byte, 255)
// AGReadDB 从 DB2710 读取，从位置 8 开始，大小为 2
err := client.AGReadDB(address, start, size, buf)
var s7 gos7.Helper
var result uint16
s7.GetValueAt(buf, 0, &result)	 

```

批量操作:
----------
批量操作允许在单个请求中读取/写入多个区域，通过减少网络往返次数显著提高性能。

```go
// 批量读取多个数据块
items := []gos7.S7DataItem{
    {Area: gos7.S7AreaDB, DBNumber: 1, Start: 0, Amount: 10, WordLen: gos7.S7WLByte, Data: make([]byte, 10)},
    {Area: gos7.S7AreaDB, DBNumber: 2, Start: 5, Amount: 20, WordLen: gos7.S7WLByte, Data: make([]byte, 20)},
    {Area: gos7.S7AreaMK, Start: 100, Amount: 8, WordLen: gos7.S7WLByte, Data: make([]byte, 8)},
}
err := client.ReadAreas(items)
// 数据现在可在每个 item 的 Data 字段中获取

// 批量写入多个数据块
writeItems := []gos7.S7DataItem{
    {Area: gos7.S7AreaDB, DBNumber: 1, Start: 0, Amount: 10, WordLen: gos7.S7WLByte, Data: []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
    {Area: gos7.S7AreaDB, DBNumber: 2, Start: 5, Amount: 4, WordLen: gos7.S7WLWord, Data: []byte{0x12, 0x34, 0x56, 0x78}},
}
err := client.WriteAreas(writeItems)
```

参考资料
----------
- libnodave http://libnodave.sourceforge.net/
- snap7 http://snap7.sourceforge.net/ 
- tarm serial library https://github.com/tarm/serial
- Simatic Open TCP/IP Communication via Industrial Ethernet from Siemens(doku)
- SIMATIC NET FDL-Programmierschnittstelle (doku)
- Elementary Data Types from Siemens (doku)

Simatic, Simatic S5, Simatic S7, S7-200, S7-300, S7-400, S7-1200, S7-1500 是西门子的注册商标

许可证
----------
https://opensource.org/licenses/BSD-3-Clause

Copyright (c) 2018, robinson