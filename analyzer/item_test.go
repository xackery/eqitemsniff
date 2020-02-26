package analyzer

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestItem(t *testing.T) {
	assert := assert.New(t)
	a, err := New()
	if !assert.NoError(err) {
		t.Fatal(err)
	}
	packet := &EQPacket{}
	packet.Data, err = hex.DecodeString("1c817000000000ffffffff04000000303030303030303030303030303030300001000000ffffffffffffff00000000000000000100000000000000973e01000000000000000000ffffffff0000000000000000000000000000000000000000000000000000000000000000000000000000ffffffff000000000000000000ffffffff00000000004461726b656e65642053696e67696e6720537465656c20427265617374706c6174650053696e67696e6720537465656c20427265617374706c61746500495436330000be3b0200640000000100000300000200b0710b007002000000004848484848023c3c3c3c361e1e391700004e1600004e160000140300000700000007000000000000008000000049a00000000000000000000000000000ffffffff00000000000000000000000000000000000000000100000000690000000000000000000000000000000000000000000000000000000032a0b4ff180000000a03000000000000000000000000000000030000000000803f1e0000000d000000000000000000000000ffffffff000000000500000001000800000001000e0000000100150000000100010000000001020000000001000000000000000000000000000000000000000000000000000000ffffffff000000000000002c000000000000000000000000000000ffffffff000000")
	if !assert.NoError(err) {
		t.Fatal(err)
	}
	err = a.itemScan(packet)
	if !assert.NoError(err) {
		t.Fatal(err)
	}
	t.Fatal("done")
}
