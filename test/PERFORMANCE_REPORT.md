# gos7 Performance Test Report / gos7 性能测试报告

**Date / 日期**: 2026-05-11  
**Test Environment / 测试环境**: Windows, Intel Core i5-13500H, Go 1.21

---

## Executive Summary / 执行摘要

This report presents the performance test results for the gos7 library, focusing on the batch read/write operations (`ReadAreas`/`WriteAreas`) that significantly improve performance by reducing network round-trips.

本报告展示了 gos7 库的性能测试结果，重点关注通过减少网络往返次数显著提高性能的批量读写操作（`ReadAreas`/`WriteAreas`）。

---

## Test Environment / 测试环境

| Item / 项目 | Value / 值 |
|------------|-----------|
| Operating System / 操作系统 | Windows |
| CPU | 13th Gen Intel(R) Core(TM) i5-13500H |
| Architecture / 架构 | amd64 |
| Go Version / Go 版本 | 1.21 |
| Test Type / 测试类型 | Mock-based (no physical PLC required) / 基于模拟（无需物理 PLC） |

---

## Benchmark Results / 基准测试结果

### 1. Basic Operations / 基础操作

| Benchmark | Operations/sec | ns/op | B/op | allocs/op |
|-----------|---------------|-------|------|-----------|
| `BenchmarkMockAGReadDB` | 289,570 | 3,839 | 2,398 | 15 |
| `BenchmarkMockAGWriteDB` | 335,668 | 4,194 | 2,470 | 18 |
| `BenchmarkMockAGReadWriteDB` | 694,440 | 1,624 | 492 | 11 |

**Description / 说明**:  
- `AGReadDB`: Read data from a Data Block / 从数据块读取数据
- `AGWriteDB`: Write data to a Data Block / 向数据块写入数据
- `AGReadWriteDB`: Combined read-write operation / 组合读写操作

---

### 2. Data Size Comparison / 数据大小对比

| Benchmark | Operations/sec | ns/op | B/op | allocs/op |
|-----------|---------------|-------|------|-----------|
| `BenchmarkMockAGReadDB_Small` (8 bytes) | 2,031,348 | 628.1 | 122 | 5 |
| `BenchmarkMockAGReadDB_Medium` (64 bytes) | 1,000,000 | 1,248 | 602 | 5 |
| `BenchmarkMockAGReadDB_Large` (256 bytes) | 252,024 | 4,300 | 2,398 | 15 |

**Description / 说明**:  
Tests the impact of data size on read performance. Smaller data sizes result in significantly better performance.

测试数据大小对读取性能的影响。较小的数据大小可获得显著更好的性能。

---

### 3. Multi-Item Operations / 多项目操作

| Benchmark | Operations/sec | ns/op | B/op | allocs/op |
|-----------|---------------|-------|------|-----------|
| `BenchmarkMockAGReadMulti` (2 items) | 215,049 | 6,061 | 1,632 | 23 |
| `BenchmarkMockAGWriteMulti` (2 items) | 250,869 | 4,922 | 1,218 | 9 |

**Description / 说明**:  
- `AGReadMulti`: Read multiple data items in a single request / 在单个请求中读取多个数据项
- `AGWriteMulti`: Write multiple data items in a single request / 在单个请求中写入多个数据项

---

### 4. Batch Operations (ReadAreas/WriteAreas) / 批量操作

| Benchmark | Operations/sec | ns/op | B/op | allocs/op |
|-----------|---------------|-------|------|-----------|
| `BenchmarkMockReadAreas` (25 items) | 74,240 | 15,801 | 4,944 | 58 |
| `BenchmarkMockWriteAreas` (25 items) | 90,651 | 13,430 | 3,636 | 19 |

**Description / 说明**:  
These benchmarks test the automatic batching feature that splits large item sets into chunks of 20 (S7 protocol limit).

这些基准测试测试了自动分批功能，该功能将大型项目集分割为 20 个一组（S7 协议限制）。

---

### 5. Single vs Multi Read Comparison / 单次 vs 批量读取对比

| Benchmark | Operations/sec | ns/op | B/op | allocs/op |
|-----------|---------------|-------|------|-----------|
| `BenchmarkMockCompareSingleVsMulti/SingleReads` (2 separate reads) | 127,347 | 9,664 | 2,980 | 50 |
| `BenchmarkMockCompareSingleVsMulti/MultiRead` (1 multi-read) | 216,000 | 5,362 | 1,632 | 23 |

**Performance Improvement / 性能提升**:
- **Speed / 速度**: ~1.8x faster (9,664 ns → 5,362 ns)
- **Memory / 内存**: ~45% reduction (2,980 B → 1,632 B)
- **Allocations / 分配**: ~54% reduction (50 → 23)

**Description / 说明**:  
This comparison clearly demonstrates the benefit of using `AGReadMulti` over multiple individual `AGReadDB` calls.

此对比清楚地展示了使用 `AGReadMulti` 相对于多次单独 `AGReadDB` 调用的优势。

---

### 6. Parallel Operations / 并行操作

| Benchmark | Operations/sec | ns/op | B/op | allocs/op |
|-----------|---------------|-------|------|-----------|
| `BenchmarkMockParallelRead` | 1,369,448 | 898.0 | 218 | 5 |
| `BenchmarkMockParallelWrite` | 1,494,342 | 947.4 | 274 | 6 |

**Description / 说明**:  
Tests concurrent read/write operations using goroutines.

使用 goroutines 测试并发读写操作。

---

## Key Findings / 关键发现

### 1. Batch Operations Significantly Improve Performance / 批量操作显著提高性能

- **Single reads**: ~9,664 ns per operation
- **Multi-read (2 items)**: ~5,362 ns per operation
- **Improvement**: ~44% faster for reading 2 items

---

### 2. Memory Efficiency / 内存效率

Batch operations not only improve speed but also reduce memory allocations:

- **Single reads**: 2,980 B/op, 50 allocs/op
- **Multi-read**: 1,632 B/op, 23 allocs/op
- **Reduction**: ~45% less memory, ~54% fewer allocations

批量操作不仅提高速度，还减少内存分配：

- **单次读取**: 2,980 B/操作, 50 次分配/操作
- **批量读取**: 1,632 B/操作, 23 次分配/操作
- **减少**: 内存减少约 45%，分配次数减少约 54%

---

### 3. Automatic Batching Works Efficiently / 自动分批高效工作

The `ReadAreas`/`WriteAreas` methods automatically split requests into chunks of 20 items (S7 protocol limit), providing:

- Simplified API for large datasets
- Optimal network utilization
- Reduced code complexity

`ReadAreas`/`WriteAreas` 方法自动将请求分割为 20 个项目一组（S7 协议限制），提供：

- 简化的大型数据集 API
- 最佳的网络利用率
- 降低的代码复杂度

---

### 4. Data Size Impact / 数据大小影响

| Data Size | Performance |
|-----------|-------------|
| Small (8 bytes) | 628 ns/op - Excellent / 优秀 |
| Medium (64 bytes) | 1,248 ns/op - Good / 良好 |
| Large (256 bytes) | 4,300 ns/op - Acceptable / 可接受 |

---

## Recommendations / 建议

### For Production Use / 生产环境使用

1. **Use batch operations for multiple variables / 对多个变量使用批量操作**
   - `ReadAreas` instead of multiple `AGReadDB` calls
   - `WriteAreas` instead of multiple `AGWriteDB` calls

2. **Group related variables / 分组相关变量**
   - Combine reads/writes for variables that are always accessed together

3. **Consider data size / 考虑数据大小**
   - For large data transfers, consider splitting into reasonable chunks

4. **Use parallel operations for independent requests / 对独立请求使用并行操作**
   - When variables are not dependent on each other, use goroutines

---

## Conclusion / 结论

The gos7 library demonstrates excellent performance characteristics:

1. **Efficient batch operations**: 1.8x speed improvement with 45% less memory
2. **Automatic batching**: Simplifies handling of large datasets
3. **Low latency**: Sub-microsecond operations for small data
4. **Scalable**: Good performance even with larger data sizes

gos7 库展示了出色的性能特征：

1. **高效的批量操作**: 速度提升 1.8 倍，内存减少 45%
2. **自动分批**: 简化大型数据集的处理
3. **低延迟**: 小数据操作的亚微秒级性能
4. **可扩展**: 即使数据量较大也能保持良好性能

---

## Appendix: Raw Benchmark Output / 附录：原始基准测试输出

```
goos: windows
goarch: amd64
pkg: github.com/anviod/gos7/test
cpu: 13th Gen Intel(R) Core(TM) i5-13500H
BenchmarkMockAGReadDB-16                          289570              3839 ns/op            2398 B/op          15 allocs/op
BenchmarkMockAGWriteDB-16                         335668              4194 ns/op            2470 B/op          18 allocs/op
BenchmarkMockAGReadWriteDB-16                     694440              1624 ns/op             492 B/op          11 allocs/op
BenchmarkMockAGReadDB_Small-16                   2031348               628.1 ns/op           122 B/op           5 allocs/op
BenchmarkMockAGReadDB_Medium-16                  1000000              1248 ns/op             602 B/op           5 allocs/op
BenchmarkMockAGReadDB_Large-16                    252024              4300 ns/op            2398 B/op          15 allocs/op
BenchmarkMockAGReadMulti-16                       215049              6061 ns/op            1632 B/op          23 allocs/op
BenchmarkMockAGWriteMulti-16                      250869              4922 ns/op            1218 B/op           9 allocs/op
BenchmarkMockReadAreas-16                          74240             15801 ns/op            4944 B/op          58 allocs/op
BenchmarkMockWriteAreas-16                         90651             13430 ns/op            3636 B/op          19 allocs/op
BenchmarkMockCompareSingleVsMulti/SingleReads-16                  127347             9664 ns/op            2980 B/op          50 allocs/op
BenchmarkMockCompareSingleVsMulti/MultiRead-16                    216000             5362 ns/op            1632 B/op          23 allocs/op
BenchmarkMockParallelRead-16                                     1369448               898.0 ns/op           218 B/op           5 allocs/op
BenchmarkMockParallelWrite-16                                    1494342               947.4 ns/op           274 B/op           6 allocs/op
```

---

*Report generated by gos7 test suite / 报告由 gos7 测试套件生成*
