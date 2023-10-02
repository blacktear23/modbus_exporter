package modbus

import (
	"fmt"

	"github.com/goburrow/modbus"
)

type ReadaheadBuffer struct {
	StartAddr uint16
	Length    uint16
	Data      []byte
}

func (b *ReadaheadBuffer) Contains(modAddr uint16, quantity uint16) bool {
	if modAddr < b.StartAddr || modAddr > (b.StartAddr+b.Length-1) {
		return false
	}
	endAddr := modAddr + quantity - 1
	if endAddr > (b.StartAddr + b.Length - 1) {
		return false
	}
	return true
}

func (b *ReadaheadBuffer) Read(modAddr uint16, quantity uint16) ([]byte, error) {
	if !b.Contains(modAddr, quantity) {
		return nil, fmt.Errorf("Cannot read data from read-ahead buffer")
	}
	pos := (modAddr - b.StartAddr) * 2
	bytes := quantity * 2
	return b.Data[pos : pos+bytes], nil
}

type ModbusReadAhead struct {
	client        modbus.Client
	buffers       map[int][]*ReadaheadBuffer
	readaheadSize uint16
}

func NewModbusReadAhead(client modbus.Client, readaheadSize uint16) *ModbusReadAhead {
	return &ModbusReadAhead{
		client:        client,
		buffers:       make(map[int][]*ReadaheadBuffer),
		readaheadSize: readaheadSize,
	}
}

func (m *ModbusReadAhead) Clear() {
	m.buffers = make(map[int][]*ReadaheadBuffer)
}

func (m *ModbusReadAhead) modbusRead(modFunc int, modAddr uint16, quantity uint16) (*ReadaheadBuffer, error) {
	readSize := m.readaheadSize
	if quantity > readSize {
		readSize = quantity
	}
	maxReadSize := uint16(65536 - int(modAddr))
	if maxReadSize < readSize {
		readSize = maxReadSize
	}
	var (
		data []byte
		err  error
	)
	switch modFunc {
	case 1:
		data, err = m.client.ReadCoils(modAddr, readSize)
	case 2:
		data, err = m.client.ReadDiscreteInputs(modAddr, readSize)
	case 3:
		data, err = m.client.ReadHoldingRegisters(modAddr, readSize)
	case 4:
		data, err = m.client.ReadInputRegisters(modAddr, readSize)
	default:
		err = fmt.Errorf("Unsupport ModBus function id: %d", modFunc)
	}
	if err != nil {
		return nil, err
	}
	if len(data) != int(readSize)*2 {
		return nil, fmt.Errorf("data read length not enough")
	}

	rb := &ReadaheadBuffer{
		StartAddr: modAddr,
		Length:    readSize,
		Data:      data,
	}
	m.buffers[modFunc] = append(m.buffers[modFunc], rb)
	return rb, nil
}

func (m *ModbusReadAhead) readFunc(modFunc int, modAddr uint16, quantity uint16) ([]byte, error) {
	// Find read ahead buffer
	buf := m.findBuffer(modFunc, modAddr, quantity)
	if buf != nil {
		return buf.Read(modAddr, quantity)
	}
	// Not found just read it
	buf, err := m.modbusRead(modFunc, modAddr, quantity)
	if err != nil {
		return nil, err
	}
	return buf.Read(modAddr, quantity)
}

func (m *ModbusReadAhead) findBuffer(modFunc int, modAddr uint16, quantity uint16) *ReadaheadBuffer {
	bufList, have := m.buffers[modFunc]
	if !have {
		return nil
	}
	for _, b := range bufList {
		if b.Contains(modAddr, quantity) {
			return b
		}
	}
	return nil
}

func (m *ModbusReadAhead) ReadCoils(modAddr, quantity uint16) ([]byte, error) {
	return m.readFunc(1, modAddr, quantity)
}

func (m *ModbusReadAhead) ReadDiscreteInputs(modAddr, quantity uint16) ([]byte, error) {
	return m.readFunc(2, modAddr, quantity)
}

func (m *ModbusReadAhead) ReadHoldingRegisters(modAddr, quantity uint16) ([]byte, error) {
	return m.readFunc(3, modAddr, quantity)
}

func (m *ModbusReadAhead) ReadInputRegisters(modAddr, quantity uint16) ([]byte, error) {
	return m.readFunc(4, modAddr, quantity)
}
