package gos7

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
import (
	"bytes"
	"context"
	"io"
	"net"
	"testing"
	"time"
)

func TestTCPTransporter(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		_, err = io.Copy(conn, conn)
		if err != nil {
			t.Error(err)
			return
		}
	}()
	client := &tcpTransporter{
		Address:     ln.Addr().String(),
		Timeout:     200 * time.Second,
		IdleTimeout: 100 * time.Millisecond,
	}
	req := []byte{0, 1, 0, 17, 0, 2, 1, 2, 0, 1, 0, 17, 0, 2, 1, 2, 2} //lengh 17, > MinPduSize

	err = client.tcpConnect(context.Background()) //assume tcp connect to test locally
	if err != nil {
		t.Fatal(err)
	}
	rsp, err := client.Send(req)
	if err != nil {
		t.Fatal(err)
	}
	//lth: just compare 7 first byte
	if !bytes.Equal(req, rsp) {
		t.Fatalf("unexpected response: %x", rsp)
	}
	time.Sleep(150 * time.Millisecond)
	if client.conn != nil {
		t.Fatalf("connection is not closed: %+v", client.conn)
	}
}

// TestTCPTransporterConnectContext tests context-aware connection with cancellation support
// TestTCPTransporterConnectContext 测试支持取消的 context 感知连接
func TestTCPTransporterConnectContext(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			t.Error(err)
			return
		}
		defer conn.Close()
		_, err = io.Copy(conn, conn)
		if err != nil {
			t.Error(err)
			return
		}
	}()

	client := &tcpTransporter{
		Address:     ln.Addr().String(),
		Timeout:     200 * time.Second,
		IdleTimeout: 100 * time.Millisecond,
	}

	req := []byte{0, 1, 0, 17, 0, 2, 1, 2, 0, 1, 0, 17, 0, 2, 1, 2, 2}

	// Test ConnectContext with background context
	// 使用 background context 测试 ConnectContext
	ctx := context.Background()
	err = client.tcpConnect(ctx)
	if err != nil {
		t.Fatal(err)
	}

	rsp, err := client.Send(req)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(req, rsp) {
		t.Fatalf("unexpected response: %x", rsp)
	}

	client.Close()
}

// TestTCPTransporterConnectContextCancellation tests connection cancellation via context
// TestTCPTransporterConnectContextCancellation 测试通过 context 取消连接
func TestTCPTransporterConnectContextCancellation(t *testing.T) {
	// Use an unreachable address to ensure connection takes time
	// 使用不可达地址确保连接需要时间
	client := &tcpTransporter{
		Address:     "192.0.2.1:102", // TEST-NET-1, should be unreachable / 测试网络，应该不可达
		Timeout:     10 * time.Second,
		IdleTimeout: 100 * time.Millisecond,
	}

	// Create a context that will be cancelled quickly
	// 创建一个快速取消的 context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// This should fail due to context cancellation
	// 这应该由于 context 取消而失败
	start := time.Now()
	err := client.tcpConnect(ctx)
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected error due to context cancellation, got nil")
	}

	// Verify that we didn't wait for the full timeout
	// 验证我们没有等待完整的超时时间
	if elapsed > 1*time.Second {
		t.Fatalf("context cancellation took too long: %v", elapsed)
	}

	t.Logf("Connection cancelled successfully after %v with error: %v", elapsed, err)
}
