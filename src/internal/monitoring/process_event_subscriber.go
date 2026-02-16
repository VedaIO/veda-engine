package monitoring

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"src/internal/data/logger"
	"src/internal/data/repository"
	"src/internal/platform/app_filter"
	"src/internal/platform/proc_sensing"
)

type ProcessEventSubscriber struct {
	logger logger.Logger
	repo   *repository.AppRepository

	runningProcs     map[string]string
	runningAppCounts map[string]int
	sync.Mutex
}

func NewProcessEventSubscriber(appLogger logger.Logger, appRepo *repository.AppRepository) *ProcessEventSubscriber {
	return &ProcessEventSubscriber{
		logger:           appLogger,
		repo:             appRepo,
		runningProcs:     make(map[string]string),
		runningAppCounts: make(map[string]int),
	}
}

func (s *ProcessEventSubscriber) Name() string {
	return "ProcessEventSubscriber"
}

func (s *ProcessEventSubscriber) OnProcessesChanged(snapshot ProcessSnapshot) {
	currentKeys := make(map[string]bool)
	for _, p := range snapshot.Processes {
		currentKeys[p.UniqueKey()] = true
	}

	s.logEndedProcesses(currentKeys)
	s.logNewProcesses(snapshot.Processes)
}

func (s *ProcessEventSubscriber) logEndedProcesses(currentKeys map[string]bool) {
	s.Lock()
	defer s.Unlock()

	for key, nameLower := range s.runningProcs {
		if !currentKeys[key] {
			s.repo.CloseAppEvent(key, time.Now().Unix())

			delete(s.runningProcs, key)
			s.runningAppCounts[nameLower]--
			if s.runningAppCounts[nameLower] <= 0 {
				delete(s.runningAppCounts, nameLower)
			}
		}
	}
}

func (s *ProcessEventSubscriber) logNewProcesses(procs []proc_sensing.ProcessInfo) {
	s.Lock()
	defer s.Unlock()

	for _, p := range procs {
		key := p.UniqueKey()
		if _, exists := s.runningProcs[key]; exists {
			continue
		}

		name := p.Name
		if name == "" {
			continue
		}
		nameLower := strings.ToLower(name)

		exePath := p.ExePath
		if exePath == "" {
			continue
		}

		if app_filter.ShouldExclude(exePath, &p) {
			s.runningProcs[key] = nameLower
			continue
		}

		isAlreadyLogged := s.runningAppCounts[nameLower] > 0
		if isAlreadyLogged {
			s.runningProcs[key] = nameLower
			s.runningAppCounts[nameLower]++
			continue
		}

		if !app_filter.ShouldTrack(exePath, &p) {
			continue
		}

		parentName := fmt.Sprintf("PID: %d", p.ParentPID)
		s.repo.LogAppEvent(name, p.PID, parentName, exePath, time.Now().Unix(), key)

		s.runningProcs[key] = nameLower
		s.runningAppCounts[nameLower]++
	}
}

func (s *ProcessEventSubscriber) InitializeFromDatabase() {
	activeSessions, err := s.repo.GetActiveSessions()
	if err != nil {
		s.logger.Printf("[ProcessEventSubscriber] Failed to load active sessions: %v", err)
		return
	}

	s.Lock()
	defer s.Unlock()

	procs, _ := proc_sensing.GetAllProcessesCached()
	currentKeys := make(map[string]bool)
	for _, p := range procs {
		currentKeys[p.UniqueKey()] = true
	}

	for _, session := range activeSessions {
		if currentKeys[session.Key] {
			nameLower := strings.ToLower(session.Name)
			s.runningProcs[session.Key] = nameLower
			s.runningAppCounts[nameLower]++
		} else {
			s.repo.CloseAppEvent(session.Key, time.Now().Unix())
		}
	}
}

func (s *ProcessEventSubscriber) Reset() {
	s.Lock()
	defer s.Unlock()

	s.logger.Printf("[ProcessEventSubscriber] Reset signal received. Clearing in-memory state.")
	s.runningProcs = make(map[string]string)
	s.runningAppCounts = make(map[string]int)
}
