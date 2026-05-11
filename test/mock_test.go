package test

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

import (
	"fmt"
	"testing"

	"github.com/anviod/gos7"
)

// ==================== Mock-based Unit Tests (No PLC Required) ====================

// newMockClient creates a gos7.Client backed by MockClientHandler
func newMockClient() gos7.Client {
	handler := gos7.NewMockClientHandler()
	return gos7.NewClient(handler)
}

// TestMockAGReadWriteDB tests reading and writing Data Blocks using mock PLC
func TestMockAGReadWriteDB(t *testing.T) {
	client := newMockClient()

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

// TestMockAGReadWriteMB tests reading and writing Memory Bits using mock PLC
func TestMockAGReadWriteMB(t *testing.T) {
	client := newMockClient()

	testStart := 100
	testSize := 8
	testBuffer := []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x11, 0x22}

	err := client.AGWriteMB(testStart, testSize, testBuffer)
	if err != nil {
		t.Fatalf("AGWriteMB failed: %v", err)
	}

	readBuffer := make([]byte, testSize)
	err = client.AGReadMB(testStart, testSize, readBuffer)
	if err != nil {
		t.Fatalf("AGReadMB failed: %v", err)
	}

	for i := range testBuffer {
		if readBuffer[i] != testBuffer[i] {
			t.Errorf("AGReadWriteMB: byte %d mismatch - expected 0x%02X, got 0x%02X", i, testBuffer[i], readBuffer[i])
		}
	}
}

// TestMockAGReadWriteEB tests reading Process Inputs using mock PLC
func TestMockAGReadWriteEB(t *testing.T) {
	client := newMockClient()

	testStart := 0
	testSize := 8
	testBuffer := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	// Write then read
	err := client.AGWriteEB(testStart, testSize, testBuffer)
	if err != nil {
		t.Fatalf("AGWriteEB failed: %v", err)
	}

	readBuffer := make([]byte, testSize)
	err = client.AGReadEB(testStart, testSize, readBuffer)
	if err != nil {
		t.Fatalf("AGReadEB failed: %v", err)
	}

	for i := range testBuffer {
		if readBuffer[i] != testBuffer[i] {
			t.Errorf("AGReadWriteEB: byte %d mismatch - expected 0x%02X, got 0x%02X", i, testBuffer[i], readBuffer[i])
		}
	}
}

// TestMockAGReadWriteAB tests reading and writing Process Outputs using mock PLC
func TestMockAGReadWriteAB(t *testing.T) {
	client := newMockClient()

	testStart := 0
	testSize := 4
	testBuffer := []byte{0x01, 0x02, 0x04, 0x08}

	err := client.AGWriteAB(testStart, testSize, testBuffer)
	if err != nil {
		t.Fatalf("AGWriteAB failed: %v", err)
	}

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

// TestMockAGReadWriteMulti tests multiple read/write operations using mock PLC
func TestMockAGReadWriteMulti(t *testing.T) {
	client := newMockClient()

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

	for i := range data1 {
		if readData1[i] != data1[i] {
			t.Errorf("AGReadWriteMulti: DB101 byte %d mismatch - expected %d, got %d", i, data1[i], readData1[i])
		}
		if readData2[i] != data2[i] {
			t.Errorf("AGReadWriteMulti: DB102 byte %d mismatch - expected %d, got %d", i, data2[i], readData2[i])
		}
	}
}

// TestMockReadAreas tests batch reading with automatic batching using mock PLC
func TestMockReadAreas(t *testing.T) {
	client := newMockClient()

	// First write some data to read back
	for i := 0; i < 5; i++ {
		data := make([]byte, 8)
		for j := range data {
			data[j] = byte(i*10 + j)
		}
		err := client.AGWriteDB(100+i, 0, 8, data)
		if err != nil {
			t.Fatalf("AGWriteDB failed: %v", err)
		}
	}

	items := make([]gos7.S7DataItem, 25)
	for i := range items {
		items[i] = gos7.S7DataItem{
			Area:     gos7.S7AreaDB,
			WordLen:  gos7.S7WLByte,
			DBNumber: 100 + (i % 5),
			Start:    0,
			Amount:   8,
			Data:     make([]byte, 8),
		}
	}

	err := client.ReadAreas(items)
	if err != nil {
		t.Fatalf("ReadAreas failed: %v", err)
	}
}

// TestMockWriteAreas tests batch writing with automatic batching using mock PLC
func TestMockWriteAreas(t *testing.T) {
	client := newMockClient()

	items := make([]gos7.S7DataItem, 25)
	for i := range items {
		data := make([]byte, 4)
		for j := range data {
			data[j] = byte((i*10 + j) & 0xFF)
		}
		items[i] = gos7.S7DataItem{
			Area:     gos7.S7AreaDB,
			WordLen:  gos7.S7WLByte,
			DBNumber: 100 + (i % 5),
			Start:    0,
			Amount:   4,
			Data:     data,
		}
	}

	err := client.WriteAreas(items)
	if err != nil {
		t.Fatalf("WriteAreas failed: %v", err)
	}
}

// TestMockReadWriteLargeData tests reading/writing data larger than PDU size
func TestMockReadWriteLargeData(t *testing.T) {
	client := newMockClient()

	// 1024 bytes - should be split into multiple PDU chunks
	testSize := 1024
	testBuffer := make([]byte, testSize)
	for i := range testBuffer {
		testBuffer[i] = byte(i & 0xFF)
	}

	err := client.AGWriteDB(100, 0, testSize, testBuffer)
	if err != nil {
		t.Fatalf("AGWriteDB large data failed: %v", err)
	}

	readBuffer := make([]byte, testSize)
	err = client.AGReadDB(100, 0, testSize, readBuffer)
	if err != nil {
		t.Fatalf("AGReadDB large data failed: %v", err)
	}

	for i := range testBuffer {
		if readBuffer[i] != testBuffer[i] {
			t.Errorf("Large data: byte %d mismatch - expected %d, got %d", i, testBuffer[i], readBuffer[i])
		}
	}
}

// TestMockMultiAreaReadWrite tests read/write across different memory areas
func TestMockMultiAreaReadWrite(t *testing.T) {
	client := newMockClient()

	// Write to DB
	dbData := []byte{0x01, 0x02, 0x03, 0x04}
	if err := client.AGWriteDB(1, 0, 4, dbData); err != nil {
		t.Fatalf("Write DB failed: %v", err)
	}

	// Write to Merkers
	mkData := []byte{0xAA, 0xBB}
	if err := client.AGWriteMB(0, 2, mkData); err != nil {
		t.Fatalf("Write MB failed: %v", err)
	}

	// Write to Outputs
	abData := []byte{0x55}
	if err := client.AGWriteAB(0, 1, abData); err != nil {
		t.Fatalf("Write AB failed: %v", err)
	}

	// Multi-read from different areas
	readDB := make([]byte, 4)
	readMK := make([]byte, 2)
	readAB := make([]byte, 1)

	items := []gos7.S7DataItem{
		{Area: gos7.S7AreaDB, WordLen: gos7.S7WLByte, DBNumber: 1, Start: 0, Amount: 4, Data: readDB},
		{Area: gos7.S7AreaMK, WordLen: gos7.S7WLByte, DBNumber: 0, Start: 0, Amount: 2, Data: readMK},
		{Area: gos7.S7AreaPA, WordLen: gos7.S7WLByte, DBNumber: 0, Start: 0, Amount: 1, Data: readAB},
	}

	err := client.AGReadMulti(items, 3)
	if err != nil {
		t.Fatalf("AGReadMulti across areas failed: %v", err)
	}

	// Verify DB data
	for i, v := range dbData {
		if readDB[i] != v {
			t.Errorf("Multi-area DB byte %d mismatch", i)
		}
	}
	for i, v := range mkData {
		if readMK[i] != v {
			t.Errorf("Multi-area MK byte %d mismatch", i)
		}
	}
	if readAB[0] != abData[0] {
		t.Errorf("Multi-area AB byte mismatch")
	}
}

// ==================== Mock-based Benchmark Tests ====================

// BenchmarkMockAGReadDB benchmarks reading from Data Block using mock PLC
func BenchmarkMockAGReadDB(b *testing.B) {
	client := newMockClient()
	buffer := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.AGReadDB(100, 0, 1024, buffer)
	}
}

// BenchmarkMockAGWriteDB benchmarks writing to Data Block using mock PLC
func BenchmarkMockAGWriteDB(b *testing.B) {
	client := newMockClient()
	buffer := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.AGWriteDB(100, 0, 1024, buffer)
	}
}

// BenchmarkMockAGReadWriteDB benchmarks a full read-write cycle using mock PLC
func BenchmarkMockAGReadWriteDB(b *testing.B) {
	client := newMockClient()
	writeBuffer := make([]byte, 64)
	readBuffer := make([]byte, 64)
	for i := range writeBuffer {
		writeBuffer[i] = byte(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.AGWriteDB(100, 0, 64, writeBuffer)
		_ = client.AGReadDB(100, 0, 64, readBuffer)
	}
}

// BenchmarkMockAGReadDB_Small benchmarks reading small data (16 bytes)
func BenchmarkMockAGReadDB_Small(b *testing.B) {
	client := newMockClient()
	buffer := make([]byte, 16)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.AGReadDB(100, 0, 16, buffer)
	}
}

// BenchmarkMockAGReadDB_Medium benchmarks reading medium data (256 bytes)
func BenchmarkMockAGReadDB_Medium(b *testing.B) {
	client := newMockClient()
	buffer := make([]byte, 256)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.AGReadDB(100, 0, 256, buffer)
	}
}

// BenchmarkMockAGReadDB_Large benchmarks reading large data (1024 bytes)
func BenchmarkMockAGReadDB_Large(b *testing.B) {
	client := newMockClient()
	buffer := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.AGReadDB(100, 0, 1024, buffer)
	}
}

// BenchmarkMockAGReadMulti benchmarks multiple read operations using mock PLC
func BenchmarkMockAGReadMulti(b *testing.B) {
	client := newMockClient()

	items := make([]gos7.S7DataItem, 10)
	for i := range items {
		items[i] = gos7.S7DataItem{
			Area:     gos7.S7AreaDB,
			WordLen:  gos7.S7WLByte,
			DBNumber: 100 + (i % 5),
			Start:    0,
			Amount:   10,
			Data:     make([]byte, 10),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.AGReadMulti(items, 10)
	}
}

// BenchmarkMockAGWriteMulti benchmarks multiple write operations using mock PLC
func BenchmarkMockAGWriteMulti(b *testing.B) {
	client := newMockClient()

	items := make([]gos7.S7DataItem, 10)
	for i := range items {
		items[i] = gos7.S7DataItem{
			Area:     gos7.S7AreaDB,
			WordLen:  gos7.S7WLByte,
			DBNumber: 100 + (i % 5),
			Start:    0,
			Amount:   10,
			Data:     make([]byte, 10),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.AGWriteMulti(items, 10)
	}
}

// BenchmarkMockReadAreas benchmarks batch reading with automatic batching using mock PLC
func BenchmarkMockReadAreas(b *testing.B) {
	client := newMockClient()

	items := make([]gos7.S7DataItem, 30)
	for i := range items {
		items[i] = gos7.S7DataItem{
			Area:     gos7.S7AreaDB,
			WordLen:  gos7.S7WLByte,
			DBNumber: 100 + (i % 5),
			Start:    0,
			Amount:   10,
			Data:     make([]byte, 10),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.ReadAreas(items)
	}
}

// BenchmarkMockWriteAreas benchmarks batch writing with automatic batching using mock PLC
func BenchmarkMockWriteAreas(b *testing.B) {
	client := newMockClient()

	items := make([]gos7.S7DataItem, 30)
	for i := range items {
		items[i] = gos7.S7DataItem{
			Area:     gos7.S7AreaDB,
			WordLen:  gos7.S7WLByte,
			DBNumber: 100 + (i % 5),
			Start:    0,
			Amount:   4,
			Data:     make([]byte, 4),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.WriteAreas(items)
	}
}

// BenchmarkMockCompareSingleVsMulti compares single vs multi read performance using mock PLC
func BenchmarkMockCompareSingleVsMulti(b *testing.B) {
	client := newMockClient()

	b.Run("SingleReads", func(b *testing.B) {
		buffer := make([]byte, 100)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for j := 0; j < 10; j++ {
				_ = client.AGReadDB(100+j, 0, 100, buffer)
			}
		}
	})

	b.Run("MultiRead", func(b *testing.B) {
		items := make([]gos7.S7DataItem, 10)
		for i := range items {
			items[i] = gos7.S7DataItem{
				Area:     gos7.S7AreaDB,
				WordLen:  gos7.S7WLByte,
				DBNumber: 100 + i,
				Start:    0,
				Amount:   10,
				Data:     make([]byte, 10),
			}
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = client.AGReadMulti(items, 10)
		}
	})
}

// BenchmarkMockParallelRead benchmarks parallel read operations using mock PLC
func BenchmarkMockParallelRead(b *testing.B) {
	client := newMockClient()

	b.RunParallel(func(pb *testing.PB) {
		buffer := make([]byte, 64)
		for pb.Next() {
			_ = client.AGReadDB(100, 0, 64, buffer)
		}
	})
}

// BenchmarkMockParallelWrite benchmarks parallel write operations using mock PLC
func BenchmarkMockParallelWrite(b *testing.B) {
	client := newMockClient()

	b.RunParallel(func(pb *testing.PB) {
		buffer := make([]byte, 64)
		for pb.Next() {
			_ = client.AGWriteDB(100, 0, 64, buffer)
		}
	})
}

// MockPerformanceReport prints a summary of mock performance metrics
func MockPerformanceReport() {
	fmt.Println("\n=== Mock PLC Performance Report ===")
	fmt.Println("Note: Run 'go test -bench=BenchmarkMock -benchmem ./test/...' to get actual metrics")
	fmt.Println("\nMock Benchmark Tests Available:")
	fmt.Println("  - BenchmarkMockAGReadDB          : Single DB read (1024 bytes)")
	fmt.Println("  - BenchmarkMockAGWriteDB         : Single DB write (1024 bytes)")
	fmt.Println("  - BenchmarkMockAGReadWriteDB     : Full read-write cycle (64 bytes)")
	fmt.Println("  - BenchmarkMockAGReadDB_Small    : Small read (16 bytes)")
	fmt.Println("  - BenchmarkMockAGReadDB_Medium   : Medium read (256 bytes)")
	fmt.Println("  - BenchmarkMockAGReadDB_Large    : Large read (1024 bytes)")
	fmt.Println("  - BenchmarkMockAGReadMulti       : Multi-item read (10 items x 100 bytes)")
	fmt.Println("  - BenchmarkMockAGWriteMulti      : Multi-item write (10 items x 100 bytes)")
	fmt.Println("  - BenchmarkMockReadAreas         : Batch read with auto-batching (30 items)")
	fmt.Println("  - BenchmarkMockWriteAreas        : Batch write with auto-batching (30 items)")
	fmt.Println("  - BenchmarkMockCompareSingleVsMulti : Single vs Multi comparison")
	fmt.Println("  - BenchmarkMockParallelRead      : Parallel read operations")
	fmt.Println("  - BenchmarkMockParallelWrite     : Parallel write operations")
	fmt.Println()
}
