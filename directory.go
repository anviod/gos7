package gos7

// Copyright 2018 Trung Hieu Le. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

import "encoding/binary"

const (
	// Block type byte
	blockOB  = 56
	blockDB  = 65
	blockSDB = 66
	blockFC  = 67
	blockSFC = 68
	blockFB  = 69
	blockSFB = 70
)

//S7BlocksList Block List
type S7BlocksList struct {
	OBList  []int
	FBList  []int
	FCList  []int
	SFBList []int
	SFCList []int
	DBList  []int
	SDBList []int
}

//implement list block
func (mb *client) PGListBlocks() (list S7BlocksList, err error) {
	list.OBList, err = mb.pgBlockList(blockOB)
	if err != nil {
		return
	}
	list.DBList, err = mb.pgBlockList(blockDB)
	if err != nil {
		return
	}
	list.FCList, err = mb.pgBlockList(blockFC)
	if err != nil {
		return
	}
	list.FBList, err = mb.pgBlockList(blockFB)
	if err != nil {
		return
	}
	list.SDBList, err = mb.pgBlockList(blockSDB)
	if err != nil {
		return
	}
	list.SFBList, err = mb.pgBlockList(blockSFB)
	if err != nil {
		return
	}
	list.SFCList, err = mb.pgBlockList(blockSFC)
	return
}

func (mb *client) pgBlockList(blockType byte) (arr []int, err error) {
	bl := make([]byte, len(s7PGBlockListTelegram))
	copy(bl, s7PGBlockListTelegram)
	bl = append(bl, make([]byte, 1)...)
	switch blockType {
	case blockDB:
		bl[len(bl)-1] = blockDB
	case blockOB:
		bl[len(bl)-1] = blockOB
	case blockSDB:
		bl[len(bl)-1] = blockSDB
	case blockFC:
		bl[len(bl)-1] = blockFC
	case blockSFC:
		bl[len(bl)-1] = blockSFC
	case blockFB:
		bl[len(bl)-1] = blockFB
	case blockSFB:
		bl[len(bl)-1] = blockSFB
	default:
		return
	}
	request := NewProtocolDataUnit(bl)
	//send
	response, err := mb.send(&request)
	if err == nil {
		res := make([]byte, len(response.Data)-33) //remove first 26 byte function and 7 byte header
		copy(res, response.Data[33:len(response.Data)])
		arr = dataToBlocks(res)
	}
	return
}
func dataToBlocks(data []byte) []int {
	count := len(data) / 4
	arr := make([]int, count)
	for i := 0; i < count; i++ {
		arr[i] = int(binary.BigEndian.Uint16(data[i*4 : i*4+2]))
	}
	return arr
}
