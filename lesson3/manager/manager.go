package snapshotManager

import (
	"fmt"
	"sync"
	"time"
)

var CanceledMessage = "manager is canceled"

var snapshot []string

func GetSnapshot() []string {
	return snapshot
}

type Manager struct {
	transactions []string
	waitChan     chan string
	cancelChan   chan struct{}
	snapMutex    sync.Mutex
	canceled     bool
}

func NewManager() Manager {
	return Manager{
		waitChan:   make(chan string, 1024),
		cancelChan: make(chan struct{}, 1),
		snapMutex:  sync.Mutex{},
	}
}

func (m *Manager) snapshot() {
	m.snapMutex.Lock()
	defer m.snapMutex.Unlock()

	snapshot = m.transactions
}

func (m *Manager) Put(data string) error {
	if m.canceled {
		return fmt.Errorf(CanceledMessage)
	}

	m.waitChan <- data
	return nil
}

func (m *Manager) Manage() error {
	if m.canceled {
		return fmt.Errorf(CanceledMessage)
	}

loop:
	for {
		select {
		case <-m.cancelChan:
			break loop
		case data := <-m.waitChan:
			m.snapMutex.Lock()
			m.transactions = append(m.transactions, data)
			m.snapMutex.Unlock()
		case <-time.After(time.Second):
			m.snapshot()
		}
	}
	return nil
}

func (m *Manager) Cancel() {
	m.canceled = true
	m.cancelChan <- struct{}{}
}
