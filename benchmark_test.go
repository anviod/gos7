package gos7

import (
	"encoding/binary"
	"math"
	"testing"
)

// ============================================================================
// Helper.SetValueAt / GetValueAt Benchmarks
// These benchmarks verify the optimization of direct byte operations vs bytes.Buffer
// ============================================================================

// BenchmarkHelper_SetValueAt_Uint16 benchmarks the optimized SetValueAt for uint16
func BenchmarkHelper_SetValueAt_Uint16(b *testing.B) {
	var h Helper
	buf := make([]byte, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.SetValueAt(buf, 0, uint16(i))
	}
}

// BenchmarkHelper_SetValueAt_Uint32 benchmarks the optimized SetValueAt for uint32
func BenchmarkHelper_SetValueAt_Uint32(b *testing.B) {
	var h Helper
	buf := make([]byte, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.SetValueAt(buf, 0, uint32(i))
	}
}

// BenchmarkHelper_SetValueAt_Int64 benchmarks the optimized SetValueAt for int64
func BenchmarkHelper_SetValueAt_Int64(b *testing.B) {
	var h Helper
	buf := make([]byte, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.SetValueAt(buf, 0, int64(i))
	}
}

// BenchmarkHelper_SetValueAt_Float32 benchmarks the optimized SetValueAt for float32
func BenchmarkHelper_SetValueAt_Float32(b *testing.B) {
	var h Helper
	buf := make([]byte, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.SetValueAt(buf, 0, float32(i)*1.5)
	}
}

// BenchmarkHelper_GetValueAt_Uint16 benchmarks the optimized GetValueAt for uint16
func BenchmarkHelper_GetValueAt_Uint16(b *testing.B) {
	var h Helper
	buf := make([]byte, 64)
	binary.BigEndian.PutUint16(buf, 12345)
	var result uint16
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.GetValueAt(buf, 0, &result)
	}
}

// BenchmarkHelper_GetValueAt_Uint32 benchmarks the optimized GetValueAt for uint32
func BenchmarkHelper_GetValueAt_Uint32(b *testing.B) {
	var h Helper
	buf := make([]byte, 64)
	binary.BigEndian.PutUint32(buf, 123456789)
	var result uint32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.GetValueAt(buf, 0, &result)
	}
}

// BenchmarkHelper_GetValueAt_Int64 benchmarks the optimized GetValueAt for int64
func BenchmarkHelper_GetValueAt_Int64(b *testing.B) {
	var h Helper
	buf := make([]byte, 64)
	binary.BigEndian.PutUint64(buf, 1234567890123456789)
	var result int64
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.GetValueAt(buf, 0, &result)
	}
}

// BenchmarkHelper_GetValueAt_Float32 benchmarks the optimized GetValueAt for float32
func BenchmarkHelper_GetValueAt_Float32(b *testing.B) {
	var h Helper
	buf := make([]byte, 64)
	binary.BigEndian.PutUint32(buf, math.Float32bits(3.14))
	var result float32
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.GetValueAt(buf, 0, &result)
	}
}

// BenchmarkHelper_SetRealAt benchmarks the optimized SetRealAt
func BenchmarkHelper_SetRealAt(b *testing.B) {
	var h Helper
	buf := make([]byte, 64)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.SetRealAt(buf, 0, float32(i)*1.5)
	}
}

// BenchmarkHelper_GetRealAt benchmarks the optimized GetRealAt
func BenchmarkHelper_GetRealAt(b *testing.B) {
	var h Helper
	buf := make([]byte, 64)
	h.SetRealAt(buf, 0, 3.14)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.GetRealAt(buf, 0)
	}
}

// ============================================================================
// putUint24 / uint24 Benchmarks
// ============================================================================

// BenchmarkPutUint24 benchmarks the putUint24 helper
func BenchmarkPutUint24(b *testing.B) {
	buf := make([]byte, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		putUint24(buf, uint32(i))
	}
}

// BenchmarkUint24 benchmarks the uint24 helper
func BenchmarkUint24(b *testing.B) {
	buf := []byte{0x12, 0x34, 0x56}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = uint24(buf)
	}
}

// BenchmarkPutUint24_ManualCompare compares putUint24 with manual bit shifting
func BenchmarkPutUint24_ManualCompare(b *testing.B) {
	buf := make([]byte, 3)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		addr := uint32(i)
		buf[2] = byte(addr & 0x0FF)
		addr = addr >> 8
		buf[1] = byte(addr & 0x0FF)
		addr = addr >> 8
		buf[0] = byte(addr & 0x0FF)
	}
}

// ============================================================================
// dataToBlocks Benchmarks
// ============================================================================

// BenchmarkDataToBlocks benchmarks the optimized dataToBlocks
func BenchmarkDataToBlocks(b *testing.B) {
	// Create sample data with 50 blocks
	data := make([]byte, 200) // 50 blocks * 4 bytes each
	for i := range data {
		data[i] = byte(i % 256)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = dataToBlocks(data)
	}
}

// ============================================================================
// Buffer Pool Benchmarks (Tiered Pools)
// ============================================================================

// BenchmarkBufferPool_SmallGet benchmarks getting a small buffer from pool
func BenchmarkBufferPool_SmallGet(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bufPtr := smallBufferPool.Get().(*[]byte)
		_ = *bufPtr
		smallBufferPool.Put(bufPtr)
	}
}

// BenchmarkBufferPool_LargeGet benchmarks getting a large buffer from pool
func BenchmarkBufferPool_LargeGet(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bufPtr := largeBufferPool.Get().(*[]byte)
		_ = *bufPtr
		largeBufferPool.Put(bufPtr)
	}
}

// BenchmarkBufferPool_SmallAlloc benchmarks allocating a small buffer (no pool)
func BenchmarkBufferPool_SmallAlloc(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 512)
		_ = buf
	}
}

// BenchmarkBufferPool_LargeAlloc benchmarks allocating a large buffer (no pool)
func BenchmarkBufferPool_LargeAlloc(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, tcpMaxLength)
		_ = buf
	}
}

// ============================================================================
// WriteArea Request Building Benchmarks
// ============================================================================

// BenchmarkWriteArea_AppendPattern benchmarks the old nested append pattern
func BenchmarkWriteArea_AppendPattern(b *testing.B) {
	header := make([]byte, sizeHeaderWrite)
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Old pattern: nested append
		_ = append(header[:35], append(data, header[35:]...)...)
	}
}

// BenchmarkWriteArea_PreAllocCopy benchmarks the new pre-allocated copy pattern
func BenchmarkWriteArea_PreAllocCopy(b *testing.B) {
	header := make([]byte, sizeHeaderWrite)
	data := make([]byte, 100)
	for i := range data {
		data[i] = byte(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// New pattern: pre-allocate and copy
		complete := make([]byte, sizeHeaderWrite+len(data))
		copy(complete, header[:sizeHeaderWrite])
		copy(complete[sizeHeaderWrite:], data)
	}
}

// ============================================================================
// AGWriteMulti Telegram Building Benchmarks
// ============================================================================

// BenchmarkAGWriteMulti_AppendPattern benchmarks the old O(n²) append pattern
func BenchmarkAGWriteMulti_AppendPattern(b *testing.B) {
	headerLen := len(s7MultiWriteHeaderTelegram)
	paramItemLen := len(s7MultiWriteItemTelegram)
	itemsCount := 10

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s7Multi := make([]byte, headerLen)
		copy(s7Multi, s7MultiWriteHeaderTelegram)
		offset := headerLen
		for j := 0; j < itemsCount; j++ {
			paramItem := make([]byte, paramItemLen)
			copy(paramItem, s7MultiWriteItemTelegram)
			// O(n²) append pattern
			s7Multi = append(s7Multi[:offset], append(paramItem, s7Multi[offset:]...)...)
			offset += paramItemLen
		}
	}
}

// BenchmarkAGWriteMulti_PreAlloc benchmarks the new pre-allocated pattern
func BenchmarkAGWriteMulti_PreAlloc(b *testing.B) {
	headerLen := len(s7MultiWriteHeaderTelegram)
	paramItemLen := len(s7MultiWriteItemTelegram)
	itemsCount := 10

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		totalSize := headerLen + itemsCount*paramItemLen
		s7Multi := make([]byte, totalSize)
		copy(s7Multi, s7MultiWriteHeaderTelegram)
		offset := headerLen
		paramBuf := make([]byte, paramItemLen)
		for j := 0; j < itemsCount; j++ {
			copy(paramBuf, s7MultiWriteItemTelegram)
			copy(s7Multi[offset:], paramBuf)
			offset += paramItemLen
		}
	}
}

// ============================================================================
// Read Variable Parsing Benchmarks
// ============================================================================

// BenchmarkReadVariableParse benchmarks variable string parsing
func BenchmarkReadVariableParse(b *testing.B) {
	variables := []string{
		"DB1.DBB0",
		"DB10.DBW20",
		"DB5.DBD100",
		"DB1.DBX0.3",
		"MB10",
		"MW20",
		"EB0",
		"AB0",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := variables[i%len(variables)]
		_ = v[:2]
	}
}

// ============================================================================
// ARM/ Low Memory Optimized: Small Buffer Operations
// ============================================================================

// BenchmarkSmallBuffer_Pool benchmarks small buffer pooling for ARM (512B)
func BenchmarkSmallBuffer_Pool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bufPtr := getBuffer(512)
		buf := *bufPtr
		// Simulate small read
		buf[0] = 0x03
		buf[1] = 0x00
		putBuffer(bufPtr)
	}
}

// BenchmarkSmallBuffer_Alloc benchmarks small buffer allocation (no pool)
func BenchmarkSmallBuffer_Alloc(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 512)
		buf[0] = 0x03
		buf[1] = 0x00
	}
}

// BenchmarkLargeBuffer_Pool benchmarks large buffer pooling (2084B)
func BenchmarkLargeBuffer_Pool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bufPtr := getBuffer(tcpMaxLength)
		buf := *bufPtr
		buf[0] = 0x03
		buf[1] = 0x00
		putBuffer(bufPtr)
	}
}

// BenchmarkLargeBuffer_Alloc benchmarks large buffer allocation (no pool)
func BenchmarkLargeBuffer_Alloc(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, tcpMaxLength)
		buf[0] = 0x03
		buf[1] = 0x00
	}
}

// ============================================================================
// Memory Allocation Count Benchmarks (important for ARM/low-mem)
// ============================================================================

// BenchmarkAllocs_Helper_SetValueAt measures allocations for SetValueAt
func BenchmarkAllocs_Helper_SetValueAt(b *testing.B) {
	var h Helper
	buf := make([]byte, 64)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.SetValueAt(buf, 0, uint16(i))
	}
}

// BenchmarkAllocs_Helper_GetValueAt measures allocations for GetValueAt
func BenchmarkAllocs_Helper_GetValueAt(b *testing.B) {
	var h Helper
	buf := make([]byte, 64)
	var result uint16
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.GetValueAt(buf, 0, &result)
	}
}

// BenchmarkAllocs_BufferPool_Small measures allocations for small buffer pool usage
func BenchmarkAllocs_BufferPool_Small(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bufPtr := getBuffer(512)
		_ = *bufPtr
		putBuffer(bufPtr)
	}
}

// BenchmarkAllocs_BufferPool_Large measures allocations for large buffer pool usage
func BenchmarkAllocs_BufferPool_Large(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bufPtr := getBuffer(tcpMaxLength)
		_ = *bufPtr
		putBuffer(bufPtr)
	}
}

// BenchmarkAllocs_WriteArea_PreAlloc measures allocations for pre-allocated write
func BenchmarkAllocs_WriteArea_PreAlloc(b *testing.B) {
	header := make([]byte, sizeHeaderWrite)
	data := make([]byte, 100)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		complete := make([]byte, sizeHeaderWrite+len(data))
		copy(complete, header[:sizeHeaderWrite])
		copy(complete[sizeHeaderWrite:], data)
	}
}

// BenchmarkAllocs_AGWriteMulti_PreAlloc measures allocations for pre-allocated multi-write
func BenchmarkAllocs_AGWriteMulti_PreAlloc(b *testing.B) {
	headerLen := len(s7MultiWriteHeaderTelegram)
	paramItemLen := len(s7MultiWriteItemTelegram)
	itemsCount := 10

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		totalSize := headerLen + itemsCount*paramItemLen
		s7Multi := make([]byte, totalSize)
		copy(s7Multi, s7MultiWriteHeaderTelegram)
		offset := headerLen
		paramBuf := make([]byte, paramItemLen)
		for j := 0; j < itemsCount; j++ {
			copy(paramBuf, s7MultiWriteItemTelegram)
			copy(s7Multi[offset:], paramBuf)
			offset += paramItemLen
		}
	}
}
