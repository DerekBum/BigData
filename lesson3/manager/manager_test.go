package snapshotManager

import (
	"testing"
	"time"
)

func TestManager(t *testing.T) {
	m := NewManager()
	res := GetSnapshot()

	if len(res) != 0 {
		t.Fatalf("snapshot is not emty at the start")
	}

	m.Put("123")

	go m.Manage()

	time.Sleep(1500 * time.Millisecond)

	res = GetSnapshot()
	if len(res) != 1 || res[0] != "123" {
		t.Fatalf("snapshot is not correct")
	}

	m.Put("321")

	time.Sleep(1500 * time.Millisecond)

	res = GetSnapshot()
	if len(res) != 2 || res[0] != "123" || res[1] != "321" {
		t.Fatalf("snapshot is not correct")
	}
}

func TestCancel(t *testing.T) {
	m := NewManager()
	m.Cancel()

	if !m.canceled {
		t.Fatalf("manager not canceled")
	}

	err := m.Put("123")

	if err == nil || err.Error() != CanceledMessage {
		t.Fatalf("manager not canceled")
	}

	err = m.Manage()

	if err == nil || err.Error() != CanceledMessage {
		t.Fatalf("manager not canceled")
	}
}
