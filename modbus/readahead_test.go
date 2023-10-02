package modbus

import (
	"fmt"
	"log"
	"testing"
)

func TestReadaheadBuffer(t *testing.T) {
	rb := ReadaheadBuffer{
		StartAddr: 100,
		Length:    64,
		Data:      make([]byte, 128),
	}

	for i := 0; i < 128; i++ {
		rb.Data[i] = byte(i + 1)
	}

	if rb.Contains(99, 2) {
		t.Fatal("Should not contain 99, 2")
	}

	if !rb.Contains(100, 1) {
		t.Fatal("Should contain 100, 1")
	}

	if !rb.Contains(163, 1) {
		t.Fatal("Should contains 163, 1")
	}

	if rb.Contains(163, 2) {
		t.Fatal("Should not contains 163, 2")
	}

	if !rb.Contains(162, 2) {
		t.Fatal("Should contains 162, 2")
	}

	if rb.Contains(164, 1) {
		t.Fatal("Should not contains 164, 1")
	}
}

func assertByteArr(t *testing.T, data []byte, expect string) {
	dataStr := fmt.Sprintf("%v", data)
	if dataStr != expect {
		t.Fatalf("Read data should be '%s' but got '%s'", expect, dataStr)
	}
}

func TestReadaheadBufferRead(t *testing.T) {
	rb := ReadaheadBuffer{
		StartAddr: 100,
		Length:    64,
		Data:      make([]byte, 128),
	}

	for i := 0; i < 128; i++ {
		rb.Data[i] = byte(i + 1)
	}

	data, _ := rb.Read(100, 1)
	assertByteArr(t, data, "[1 2]")

	data, _ = rb.Read(163, 1)
	assertByteArr(t, data, "[127 128]")
	// fmt.Println(data)
	data, _ = rb.Read(102, 2)
	assertByteArr(t, data, "[5 6 7 8]")

	data, err := rb.Read(163, 2)
	assertByteArr(t, data, "[]")
	if err == nil {
		log.Fatal("Expect an error")
	}
}

func assertByStr(t *testing.T, value any, expect string) {
	valueStr := fmt.Sprintf("%v", value)
	if valueStr != expect {
		t.Fatalf("Expect %s but got %s", expect, valueStr)
	}
}

func assertNotByStr(t *testing.T, value any, expect string) {
	valueStr := fmt.Sprintf("%v", value)
	if valueStr == expect {
		t.Fatalf("Expect not %s but got %s", expect, valueStr)
	}
}

func TestModbusReadAhead(t *testing.T) {
	rb := &ReadaheadBuffer{
		StartAddr: 100,
		Length:    64,
		Data:      make([]byte, 128),
	}

	for i := 0; i < 128; i++ {
		rb.Data[i] = byte(i + 1)
	}

	rb2 := &ReadaheadBuffer{
		StartAddr: 165,
		Length:    64,
		Data:      make([]byte, 128),
	}

	for i := 0; i < 128; i++ {
		rb2.Data[i] = byte(i + 1)
	}

	mra := NewModbusReadAhead(nil, 64)
	mra.buffers[3] = []*ReadaheadBuffer{rb, rb2}
	buf := mra.findBuffer(3, 100, 2)
	assertNotByStr(t, buf, "<nil>")
	data, _ := buf.Read(100, 2)
	assertByteArr(t, data, "[1 2 3 4]")
	buf = mra.findBuffer(3, 380, 2)
	assertByStr(t, buf, "<nil>")
	buf = mra.findBuffer(3, 163, 2)
	assertByStr(t, buf, "<nil>")
	buf = mra.findBuffer(3, 164, 4)
	assertByStr(t, buf, "<nil>")
}
