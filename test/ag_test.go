package test

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
	"fmt"
	"testing"
	"time"

	"github.com/anviod/gos7"
)

// ==================== Unit Tests for AG Automation Device ====================

// TestAGReadWriteDB tests reading and writing Data Blocks
func TestAGReadWriteDB(t *testing.T) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 30 * time.Second
	handler.IdleTimeout = 30 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		t.Skipf("Skipping test - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)

	// Test data
	testDB := 100
	testStart := 0
	testSize := 16
	testBuffer := make([]byte, testSize)
	for i := range testBuffer {
		testBuffer[i] = byte(i + 1)
	}

	// Write test data
	err := client.AGWriteDB(testDB, testStart, testSize, testBuffer)
	if err != nil {
		t.Fatalf("AGWriteDB failed: %v", err)
	}

	// Read back and verify
	readBuffer := make([]byte, testSize)
	err = client.AGReadDB(testDB, testStart, testSize, readBuffer)
	if err != nil {
		t.Fatalf("AGReadDB failed: %v", err)
	}

	for i := range testBuffer {
		if readBuffer[i] != testBuffer[i] {
			t.Errorf("AGReadWriteDB: byte %d mismatch - expected %d, got %d", i, testBuffer[i], readBuffer[i])
		}
	}
}

// TestAGReadWriteMB tests reading and writing Memory Bits (Merkers)
func TestAGReadWriteMB(t *testing.T) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 30 * time.Second
	handler.IdleTimeout = 30 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		t.Skipf("Skipping test - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)

	// Test data
	testStart := 100
	testSize := 8
	testBuffer := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x11, 0x22}

	// Write test data
	err := client.AGWriteMB(testStart, testSize, testBuffer)
	if err != nil {
		t.Fatalf("AGWriteMB failed: %v", err)
	}

	// Read back and verify
	readBuffer := make([]byte, testSize)
	err = client.AGReadMB(testStart, testSize, readBuffer)
	if err != nil {
		t.Fatalf("AGReadMB failed: %v", err)
	}

	for i := range testBuffer {
		if readBuffer[i] != testBuffer[i] {
			t.Errorf("AGReadWriteMB: byte %d mismatch - expected %d, got %d", i, testBuffer[i], readBuffer[i])
		}
	}
}

// TestAGReadWriteEB tests reading and writing Process Inputs (E)
func TestAGReadWriteEB(t *testing.T) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 30 * time.Second
	handler.IdleTimeout = 30 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		t.Skipf("Skipping test - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)

	// Test read (Inputs are typically read-only from PLC perspective)
	testStart := 0
	testSize := 8
	readBuffer := make([]byte, testSize)
	err := client.AGReadEB(testStart, testSize, readBuffer)
	if err != nil {
		t.Fatalf("AGReadEB failed: %v", err)
	}

	t.Logf("AGReadEB: read %d bytes from inputs starting at %d", testSize, testStart)
}

// TestAGReadWriteAB tests reading and writing Process Outputs (A)
func TestAGReadWriteAB(t *testing.T) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 30 * time.Second
	handler.IdleTimeout = 30 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		t.Skipf("Skipping test - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)

	// Test data
	testStart := 0
	testSize := 4
	testBuffer := []byte{0x01, 0x02, 0x04, 0x08}

	// Write test data
	err := client.AGWriteAB(testStart, testSize, testBuffer)
	if err != nil {
		t.Fatalf("AGWriteAB failed: %v", err)
	}

	// Read back and verify
	readBuffer := make([]byte, testSize)
	err = client.AGReadAB(testStart, testSize, readBuffer)
	if err != nil {
		t.Fatalf("AGReadAB failed: %v", err)
	}

	for i := range testBuffer {
		if readBuffer[i] != testBuffer[i] {
			t.Errorf("AGReadWriteAB: byte %d mismatch - expected %d, got %d", i, testBuffer[i], readBuffer[i])
		}
	}
}

// TestAGReadWriteTM tests reading and writing Timers
func TestAGReadWriteTM(t *testing.T) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 30 * time.Second
	handler.IdleTimeout = 30 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		t.Skipf("Skipping test - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)

	// Test data
	testStart := 0
	testAmount := 2
	testBuffer := []byte{0x10, 0x20} // Timer values

	// Write test data
	err := client.AGWriteTM(testStart, testAmount, testBuffer)
	if err != nil {
		t.Fatalf("AGWriteTM failed: %v", err)
	}

	// Read back and verify
	readBuffer := make([]byte, testAmount)
	err = client.AGReadTM(testStart, testAmount, readBuffer)
	if err != nil {
		t.Fatalf("AGReadTM failed: %v", err)
	}

	for i := range testBuffer {
		if readBuffer[i] != testBuffer[i] {
			t.Errorf("AGReadWriteTM: byte %d mismatch - expected %d, got %d", i, testBuffer[i], readBuffer[i])
		}
	}
}

// TestAGReadWriteCT tests reading and writing Counters
func TestAGReadWriteCT(t *testing.T) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 30 * time.Second
	handler.IdleTimeout = 30 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		t.Skipf("Skipping test - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)

	// Test data
	testStart := 0
	testAmount := 2
	testBuffer := []byte{0x0A, 0x0B} // Counter values

	// Write test data
	err := client.AGWriteCT(testStart, testAmount, testBuffer)
	if err != nil {
		t.Fatalf("AGWriteCT failed: %v", err)
	}

	// Read back and verify
	readBuffer := make([]byte, testAmount)
	err = client.AGReadCT(testStart, testAmount, readBuffer)
	if err != nil {
		t.Fatalf("AGReadCT failed: %v", err)
	}

	for i := range testBuffer {
		if readBuffer[i] != testBuffer[i] {
			t.Errorf("AGReadWriteCT: byte %d mismatch - expected %d, got %d", i, testBuffer[i], readBuffer[i])
		}
	}
}

// TestAGReadWriteMulti tests multiple read/write operations
func TestAGReadWriteMulti(t *testing.T) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 30 * time.Second
	handler.IdleTimeout = 30 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		t.Skipf("Skipping test - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)

	// Prepare test data
	data1 := make([]byte, 16)
	data2 := make([]byte, 16)
	for i := range data1 {
		data1[i] = byte(i + 1)
		data2[i] = byte(i + 100)
	}

	// Write multiple items
	writeItems := []gos7.S7DataItem{
		{
			Area:     gos7.S7AreaDB,
			WordLen:  gos7.S7WLByte,
			DBNumber: 101,
			Start:    0,
			Amount:   16,
			Data:     data1,
		},
		{
			Area:     gos7.S7AreaDB,
			WordLen:  gos7.S7WLByte,
			DBNumber: 102,
			Start:    0,
			Amount:   16,
			Data:     data2,
		},
	}

	err := client.AGWriteMulti(writeItems, 2)
	if err != nil {
		t.Fatalf("AGWriteMulti failed: %v", err)
	}

	// Read back multiple items
	readData1 := make([]byte, 16)
	readData2 := make([]byte, 16)
	readItems := []gos7.S7DataItem{
		{
			Area:     gos7.S7AreaDB,
			WordLen:  gos7.S7WLByte,
			DBNumber: 101,
			Start:    0,
			Amount:   16,
			Data:     readData1,
		},
		{
			Area:     gos7.S7AreaDB,
			WordLen:  gos7.S7WLByte,
			DBNumber: 102,
			Start:    0,
			Amount:   16,
			Data:     readData2,
		},
	}

	err = client.AGReadMulti(readItems, 2)
	if err != nil {
		t.Fatalf("AGReadMulti failed: %v", err)
	}

	// Verify
	for i := range data1 {
		if readData1[i] != data1[i] {
			t.Errorf("AGReadWriteMulti: DB101 byte %d mismatch", i)
		}
		if readData2[i] != data2[i] {
			t.Errorf("AGReadWriteMulti: DB102 byte %d mismatch", i)
		}
	}
}

// TestReadAreas tests batch reading with automatic batching
func TestReadAreas(t *testing.T) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 30 * time.Second
	handler.IdleTimeout = 30 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		t.Skipf("Skipping test - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)

	// Prepare test items (more than 20 to test automatic batching)
	items := make([]gos7.S7DataItem, 25)
	for i := range items {
		items[i] = gos7.S7DataItem{
			Area:     gos7.S7AreaDB,
			WordLen:  gos7.S7WLByte,
			DBNumber: 100 + (i % 5),
			Start:    i * 10,
			Amount:   8,
			Data:     make([]byte, 8),
		}
	}

	err := client.ReadAreas(items)
	if err != nil {
		t.Fatalf("ReadAreas failed: %v", err)
	}

	t.Logf("ReadAreas: successfully read %d items", len(items))
}

// TestWriteAreas tests batch writing with automatic batching
func TestWriteAreas(t *testing.T) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 30 * time.Second
	handler.IdleTimeout = 30 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		t.Skipf("Skipping test - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)

	// Prepare test items (more than 20 to test automatic batching)
	items := make([]gos7.S7DataItem, 25)
	for i := range items {
		data := make([]byte, 8)
		for j := range data {
			data[j] = byte((i * 10) + j)
		}
		items[i] = gos7.S7DataItem{
			Area:     gos7.S7AreaDB,
			WordLen:  gos7.S7WLByte,
			DBNumber: 100 + (i % 5),
			Start:    i * 10,
			Amount:   8,
			Data:     data,
		}
	}

	err := client.WriteAreas(items)
	if err != nil {
		t.Fatalf("WriteAreas failed: %v", err)
	}

	t.Logf("WriteAreas: successfully wrote %d items", len(items))
}

// TestReadS7Syntax tests reading with S7 syntax
func TestReadS7Syntax(t *testing.T) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 30 * time.Second
	handler.IdleTimeout = 30 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		t.Skipf("Skipping test - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)

	// Test various S7 syntax formats
	testCases := []struct {
		name     string
		variable string
	}{
		{"DB Byte", "DB100.DBB0"},
		{"DB Word", "DB100.DBW2"},
		{"DB DWord", "DB100.DBD4"},
		{"Memory Byte", "MB100"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buffer := make([]byte, 4)
			_, err := client.Read(tc.variable, buffer)
			if err != nil {
				t.Errorf("Read(%q) failed: %v", tc.variable, err)
			}
		})
	}
}

// ==================== Performance Tests for AG Automation Device ====================

// BenchmarkAGReadDB benchmarks reading from Data Block
func BenchmarkAGReadDB(b *testing.B) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 60 * time.Second
	handler.IdleTimeout = 60 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		b.Skipf("Skipping benchmark - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)
	buffer := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.AGReadDB(100, 0, 1024, buffer)
	}
}

// BenchmarkAGWriteDB benchmarks writing to Data Block
func BenchmarkAGWriteDB(b *testing.B) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 60 * time.Second
	handler.IdleTimeout = 60 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		b.Skipf("Skipping benchmark - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)
	buffer := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.AGWriteDB(100, 0, 1024, buffer)
	}
}

// BenchmarkAGReadMulti benchmarks multiple read operations
func BenchmarkAGReadMulti(b *testing.B) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 60 * time.Second
	handler.IdleTimeout = 60 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		b.Skipf("Skipping benchmark - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)

	// Prepare 10 items per request
	items := make([]gos7.S7DataItem, 10)
	for i := range items {
		items[i] = gos7.S7DataItem{
			Area:     gos7.S7AreaDB,
			WordLen:  gos7.S7WLByte,
			DBNumber: 100 + (i % 5),
			Start:    i * 100,
			Amount:   100,
			Data:     make([]byte, 100),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.AGReadMulti(items, 10)
	}
}

// BenchmarkAGWriteMulti benchmarks multiple write operations
func BenchmarkAGWriteMulti(b *testing.B) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 60 * time.Second
	handler.IdleTimeout = 60 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		b.Skipf("Skipping benchmark - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)

	// Prepare 10 items per request
	items := make([]gos7.S7DataItem, 10)
	for i := range items {
		items[i] = gos7.S7DataItem{
			Area:     gos7.S7AreaDB,
			WordLen:  gos7.S7WLByte,
			DBNumber: 100 + (i % 5),
			Start:    i * 100,
			Amount:   100,
			Data:     make([]byte, 100),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.AGWriteMulti(items, 10)
	}
}

// BenchmarkReadAreas benchmarks batch reading with automatic batching
func BenchmarkReadAreas(b *testing.B) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 60 * time.Second
	handler.IdleTimeout = 60 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		b.Skipf("Skipping benchmark - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)

	// Prepare 30 items to test automatic batching
	items := make([]gos7.S7DataItem, 30)
	for i := range items {
		items[i] = gos7.S7DataItem{
			Area:     gos7.S7AreaDB,
			WordLen:  gos7.S7WLByte,
			DBNumber: 100 + (i % 5),
			Start:    i * 50,
			Amount:   50,
			Data:     make([]byte, 50),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.ReadAreas(items)
	}
}

// BenchmarkWriteAreas benchmarks batch writing with automatic batching
func BenchmarkWriteAreas(b *testing.B) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 60 * time.Second
	handler.IdleTimeout = 60 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		b.Skipf("Skipping benchmark - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)

	// Prepare 30 items to test automatic batching
	items := make([]gos7.S7DataItem, 30)
	for i := range items {
		items[i] = gos7.S7DataItem{
			Area:     gos7.S7AreaDB,
			WordLen:  gos7.S7WLByte,
			DBNumber: 100 + (i % 5),
			Start:    i * 50,
			Amount:   50,
			Data:     make([]byte, 50),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.WriteAreas(items)
	}
}

// BenchmarkCompareSingleVsMulti compares single vs multi read performance
func BenchmarkCompareSingleVsMulti(b *testing.B) {
	handler := gos7.NewTCPClientHandler(tcpDevice, rack, slot)
	handler.Timeout = 60 * time.Second
	handler.IdleTimeout = 60 * time.Second
	defer handler.Close()

	if err := handler.Connect(); err != nil {
		b.Skipf("Skipping benchmark - unable to connect to PLC: %v", err)
	}

	client := gos7.NewClient(handler)

	// Test reading 10 separate DBs
	b.Run("Single Reads", func(b *testing.B) {
		buffer := make([]byte, 100)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j < 10; j++ {
				_ = client.AGReadDB(100+j, 0, 100, buffer)
			}
		}
	})

	b.Run("Multi Read", func(b *testing.B) {
		items := make([]gos7.S7DataItem, 10)
		for i := range items {
			items[i] = gos7.S7DataItem{
				Area:     gos7.S7AreaDB,
				WordLen:  gos7.S7WLByte,
				DBNumber: 100 + i,
				Start:    0,
				Amount:   100,
				Data:     make([]byte, 100),
			}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = client.AGReadMulti(items, 10)
		}
	})
}

// PerformanceReport prints a summary of performance metrics
func PerformanceReport() {
	fmt.Println("\n=== AG Automation Device Performance Report ===")
	fmt.Println("Note: Run 'go test -bench=. -benchmem' in test directory to get actual metrics")
	fmt.Println("\nBenchmark Tests Available:")
	fmt.Println("  - BenchmarkAGReadDB        : Single DB read performance")
	fmt.Println("  - BenchmarkAGWriteDB       : Single DB write performance")
	fmt.Println("  - BenchmarkAGReadMulti     : Multi-item read performance")
	fmt.Println("  - BenchmarkAGWriteMulti    : Multi-item write performance")
	fmt.Println("  - BenchmarkReadAreas       : Batch read with auto-batching")
	fmt.Println("  - BenchmarkWriteAreas      : Batch write with auto-batching")
	fmt.Println("  - BenchmarkCompareSingleVsMulti : Single vs Multi comparison")
	fmt.Println("\nUnit Tests Available:")
	fmt.Println("  - TestAGReadWriteDB        : DB read/write")
	fmt.Println("  - TestAGReadWriteMB        : Memory bits read/write")
	fmt.Println("  - TestAGReadWriteEB        : Inputs read")
	fmt.Println("  - TestAGReadWriteAB        : Outputs read/write")
	fmt.Println("  - TestAGReadWriteTM        : Timer read/write")
	fmt.Println("  - TestAGReadWriteCT        : Counter read/write")
	fmt.Println("  - TestAGReadWriteMulti     : Multi-item read/write")
	fmt.Println("  - TestReadAreas            : Batch read")
	fmt.Println("  - TestWriteAreas           : Batch write")
	fmt.Println("  - TestReadS7Syntax         : S7 syntax read")
	fmt.Println()
}