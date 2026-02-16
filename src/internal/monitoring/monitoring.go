package monitoring

import (
	"sync"
	"sync/atomic"
	"time"

	"src/internal/data/logger"
	"src/internal/platform/proc_sensing"
)

const (
	DefaultPollingInterval = 2 * time.Second
)

type ProcessSnapshot struct {
	// Processes is the list of all running processes at the time of snapshot.
	Processes []proc_sensing.ProcessInfo
	// Timestamp is when this snapshot was captured.
	Timestamp time.Time
}

type ProcessSubscriber interface {
	OnProcessesChanged(snapshot ProcessSnapshot)
	Name() string
}

type ResettableSubscriber interface {
	ProcessSubscriber
	Reset()
}

type MonitoringManager struct {
	logger              logger.Logger
	subscribers         []ProcessSubscriber
	pollingInterval     time.Duration
	stopCh              chan struct{}
	wg                  sync.WaitGroup
	processSnapshotCh   chan ProcessSnapshot
	resetCh             chan struct{}
	isRunning           atomic.Bool
	lastTickTime        atomic.Value
	restartDelay        time.Duration
	restartMaxRetries   int
	consecutiveFailures atomic.Int32
}

var globalManager *MonitoringManager

func SetGlobalManager(manager *MonitoringManager) {
	globalManager = manager
}

func ResetGlobalManager() {
	if globalManager != nil {
		globalManager.Reset()
	}
}

func NewMonitoringManager(appLogger logger.Logger, pollingInterval time.Duration) *MonitoringManager {
	if pollingInterval <= 0 {
		pollingInterval = DefaultPollingInterval
	}
	m := &MonitoringManager{
		logger:            appLogger,
		subscribers:       make([]ProcessSubscriber, 0),
		pollingInterval:   pollingInterval,
		stopCh:            make(chan struct{}),
		processSnapshotCh: make(chan ProcessSnapshot, 1),
		resetCh:           make(chan struct{}, 1),
		restartDelay:      5 * time.Second,
		restartMaxRetries: 3,
	}
	m.lastTickTime.Store(time.Time{})
	return m
}

func (m *MonitoringManager) RegisterSubscriber(subscriber ProcessSubscriber) {
	m.subscribers = append(m.subscribers, subscriber)
}

func (m *MonitoringManager) Start() {
	if m.isRunning.Load() {
		m.logger.Printf("[MonitoringManager] Already running, skipping start")
		return
	}
	m.isRunning.Store(true)
	m.wg.Add(1)
	go m.runEventLoopWithRecovery()
}

func (m *MonitoringManager) Stop() {
	if !m.isRunning.Load() {
		return
	}
	m.isRunning.Store(false)
	close(m.stopCh)
	m.wg.Wait()
}

func (m *MonitoringManager) runEventLoopWithRecovery() {
	defer func() {
		m.wg.Done()
		if r := recover(); r != nil {
			m.logger.Printf("[MonitoringManager] PANIC RECOVERY: %v", r)
			m.handleFailure()
		}
	}()
	m.runEventLoop()
}

func (m *MonitoringManager) handleFailure() {
	m.isRunning.Store(false)
	m.consecutiveFailures.Add(1)
	failures := m.consecutiveFailures.Load()

	if failures >= int32(m.restartMaxRetries) {
		m.logger.Printf("[MonitoringManager restart] Max retries (%d) reached, giving up", m.restartMaxRetries)
		return
	}

	m.logger.Printf("[MonitoringManager] Restarting in %v (attempt %d/%d)",
		m.restartDelay, failures+1, m.restartMaxRetries)

	time.Sleep(m.restartDelay)

	m.logger.Printf("[MonitoringManager] Restarting monitoring loop...")
	m.Start()
}

func (m *MonitoringManager) Reset() {
	m.logger.Printf("[MonitoringManager] Reset signal received")
	for _, subscriber := range m.subscribers {
		if resettable, ok := subscriber.(ResettableSubscriber); ok {
			resettable.Reset()
		}
	}
}

func (m *MonitoringManager) runEventLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.pollingInterval)
	defer ticker.Stop()

	m.logger.Printf("[MonitoringManager] Started with polling interval: %v", m.pollingInterval)

	for {
		select {
		case <-m.stopCh:
			m.logger.Printf("[MonitoringManager] Stopping")
			return
		case snapshot := <-m.processSnapshotCh:
			m.notifySubscribers(snapshot)
		case <-m.resetCh:
			m.Reset()
		case <-ticker.C:
			m.captureAndNotify()
		}
	}
}

func (m *MonitoringManager) captureAndNotify() {
	procs, err := proc_sensing.GetAllProcessesCached()
	if err != nil {
		m.logger.Printf("[MonitoringManager] Failed to capture process snapshot: %v", err)
		return
	}

	m.lastTickTime.Store(time.Now())
	m.consecutiveFailures.Store(0)

	snapshot := ProcessSnapshot{
		Processes: procs,
		Timestamp: time.Now(),
	}

	m.notifySubscribers(snapshot)
}

func (m *MonitoringManager) notifySubscribers(snapshot ProcessSnapshot) {
	for _, subscriber := range m.subscribers {
		func() {
			defer func() {
				if r := recover(); r != nil {
					m.logger.Printf("[MonitoringManager] Subscriber %s panicked: %v", subscriber.Name(), r)
				}
			}()
			subscriber.OnProcessesChanged(snapshot)
		}()
	}
}

type HealthStatus struct {
	IsHealthy           bool
	LastTickTime        time.Time
	ConsecutiveFailures int32
	SubscriberCount     int
}

func (m *MonitoringManager) HealthCheck() HealthStatus {
	lastTick := m.lastTickTime.Load().(time.Time)
	return HealthStatus{
		IsHealthy:           m.isRunning.Load(),
		LastTickTime:        lastTick,
		ConsecutiveFailures: m.consecutiveFailures.Load(),
		SubscriberCount:     len(m.subscribers),
	}
}
