package snapshotManager

import (
	"fmt"
	"sync"
	"sync/atomic"

	jsonpatch "github.com/evanphx/json-patch/v5"
)

var CanceledMessage = "manager is canceled"

type Transaction struct {
	Data   string
	Source string
	Id     int
}

type Manager struct {
	cancelChan   chan struct{}
	mutex        sync.Mutex
	canceled     bool
	waitChan     chan Transaction
	transId      atomic.Int32
	transactions []Transaction
	clock        map[string]int
	snap         string

	clientId    int
	clientsTalk map[int]chan Transaction
}

func NewManager() Manager {
	return Manager{
		waitChan:    make(chan Transaction, 1024),
		cancelChan:  make(chan struct{}, 1),
		mutex:       sync.Mutex{},
		transId:     atomic.Int32{},
		snap:        "{}",
		clientsTalk: map[int]chan Transaction{},
		clock:       map[string]int{},
	}
}

func (m *Manager) NewTransaction(data, source string) Transaction {
	if m.canceled {
		panic(fmt.Errorf(CanceledMessage))
	}

	currId := m.transId.Add(1)
	return Transaction{data, source, int(currId)}
}

func (m *Manager) Snapshot() string {
	if m.canceled {
		panic(fmt.Errorf(CanceledMessage))
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.snap
}

func (m *Manager) Clock() map[string]int {
	if m.canceled {
		panic(fmt.Errorf(CanceledMessage))
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.clock
}

func (m *Manager) Put(trx Transaction) error {
	if m.canceled {
		return fmt.Errorf(CanceledMessage)
	}

	m.waitChan <- trx
	return nil
}

func (m *Manager) Manage() error {
	if m.canceled {
		return fmt.Errorf(CanceledMessage)
	}

	for {
		select {
		case trx := <-m.waitChan:
			if m.clock[trx.Source] >= trx.Id {
				break
			}
			data, _ := jsonpatch.DecodePatch([]byte(trx.Data))
			newData, _ := data.Apply([]byte(m.snap))

			m.mutex.Lock()

			m.snap = string(newData)
			m.transactions = append(m.transactions, trx)
			m.clock[trx.Source] = trx.Id

			go func() {
				for _, clientChan := range m.clientsTalk {
					clientChan <- trx
				}
			}()

			m.mutex.Unlock()
		case <-m.cancelChan:
			return nil
		default:
		}
	}
}

func (m *Manager) Cancel() {
	m.canceled = true
	m.cancelChan <- struct{}{}
}

func (m *Manager) NewClient() (int, []Transaction, chan Transaction) {
	if m.canceled {
		panic(fmt.Errorf(CanceledMessage))
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	currId := m.clientId
	m.clientId++
	m.clientsTalk[currId] = make(chan Transaction, 1024)
	return currId, m.transactions, m.clientsTalk[currId]
}

func (m *Manager) DeleteClient(id int) {
	if m.canceled {
		panic(fmt.Errorf(CanceledMessage))
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.clientsTalk, id)
}
