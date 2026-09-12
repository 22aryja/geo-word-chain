package storage

import (
	"sync"
	"testing"

	"github.com/22aryja/geo-word-chain/internal/cities"
)

// Memory must satisfy Store.
var _ Store = (*Memory)(nil)

func store(t *testing.T) *Memory {
	t.Helper()
	idx, err := cities.New()
	if err != nil {
		t.Fatal(err)
	}
	return NewMemory(idx)
}

func TestLifecycle(t *testing.T) {
	s := store(t)

	if _, ok := s.Get(1); ok {
		t.Error("a fresh store must have no games")
	}
	if s.Stop(1) {
		t.Error("Stop must report false when nothing was running")
	}

	started := s.Start(1)
	got, ok := s.Get(1)
	if !ok || got != started {
		t.Fatal("Get must return the game Start created")
	}
	if !s.Stop(1) {
		t.Error("Stop must report true when a game was running")
	}
	if _, ok := s.Get(1); ok {
		t.Error("game survived Stop")
	}
}

func TestChatsAreIsolated(t *testing.T) {
	s := store(t)

	a, b := s.Start(1), s.Start(2)
	if a == b {
		t.Fatal("two chats got the same game")
	}

	a.Play("Москва")
	if b.Score() != 0 {
		t.Error("a move in chat 1 changed chat 2")
	}

	// Restarting one chat must not disturb the other.
	s.Start(1)
	if _, ok := s.Get(2); !ok {
		t.Error("chat 2 lost its game")
	}
}

func TestConcurrentAccess(t *testing.T) {
	s := store(t)

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(1)
		go func(chatID int64) {
			defer wg.Done()
			s.Start(chatID)
			if g, ok := s.Get(chatID); ok {
				g.Play("Москва")
			}
			s.Stop(chatID)
		}(int64(i))
	}
	wg.Wait()

	if got := s.Len(); got != 0 {
		t.Errorf("Len = %d, want 0 after every game was stopped", got)
	}
}
