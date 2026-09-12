package storage

import (
	"sync"

	"github.com/22aryja/geo-word-chain/internal/cities"
	"github.com/22aryja/geo-word-chain/internal/game"
)

type Store interface {
	Get(chatID int64) (*game.Game, bool)
	Start(chatID int64) *game.Game
	Stop(chatID int64) bool
}

type Memory struct {
	index *cities.Index

	mu    sync.RWMutex
	games map[int64]*game.Game
}

func NewMemory(index *cities.Index) *Memory {
	return &Memory{
		index: index,
		games: make(map[int64]*game.Game),
	}
}

func (m *Memory) Get(chatID int64) (*game.Game, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	g, ok := m.games[chatID]
	return g, ok
}

func (m *Memory) Start(chatID int64) *game.Game {
	g := game.New(m.index)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.games[chatID] = g
	return g
}

func (m *Memory) Stop(chatID int64) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.games[chatID]
	delete(m.games, chatID)
	return ok
}

func (m *Memory) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.games)
}
