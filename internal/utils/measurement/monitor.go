package measurement

import "time"

type Monitor interface {
	Start()
	Stop() bool
	Pause() bool
	Resume() bool
	IsPaused() bool
	IsRunning() bool
	Accrued() time.Duration
	Reset()
	SetError()
}

var (
	_ Monitor = (*defaultMonitor)(nil)
	_ Monitor = (*nullMonitor)(nil)
)

type defaultMonitor struct {
	start   time.Time
	pause   time.Time
	accrued time.Duration
	running bool
	paused  bool
	point   *Point
}

type nullMonitor struct {
	started bool
	paused  bool
}

func (m *nullMonitor) Start() {
	m.started = true
}
func (m *nullMonitor) Stop() bool {
	m.started = false
	return true
}
func (m *nullMonitor) Pause() bool {
	m.paused = true
	return true
}
func (m *nullMonitor) Resume() bool {
	m.paused = false
	return true
}
func (m *nullMonitor) IsPaused() bool {
	return m.paused
}
func (m *nullMonitor) IsRunning() bool {
	return m.started
}
func (m *nullMonitor) Accrued() time.Duration {
	return 0
}
func (m *nullMonitor) Reset() {
	m.started = false
	m.paused = false
}

func (m *nullMonitor) SetError() {
	// no op
}

func newMonitor(p *Point) *defaultMonitor {
	return &defaultMonitor{
		point:   p,
		running: false,
		paused:  false,
		accrued: 0,
	}
}

// Start the time measurment for this monitor
func (m *defaultMonitor) Start() {
	m.start = time.Now()
	m.running = true
	if m.point != nil {
		m.point.activateMonitor(m)
	}
}

// Pause set the monitor into paused mode, return true if ok
func (m *defaultMonitor) Pause() bool {
	if m.running {
		m.pause = time.Now()
		m.paused = true
		return true
	}
	return false
}

// Resume the paused time measurement
func (m *defaultMonitor) Resume() bool {
	if m.running && m.paused {
		m.accrued += m.pause.Sub(m.start)
		m.start = time.Now()
		m.paused = false
		return true
	}
	return false
}

// IsPaused true if the monitor is in paused mode
func (m *defaultMonitor) IsPaused() bool {
	return m.paused
}

// IsRunning true if the monitor is in paused mode
func (m *defaultMonitor) IsRunning() bool {
	return m.running
}

// Stop the time measurement of this monitor
func (m *defaultMonitor) Stop() bool {
	if m.running {
		m.accrued += time.Now().Sub(m.start)
		m.running = false
		if m.point != nil {
			m.point.processMonitor(m)
		}
		return true
	}
	return false
}

// Accrued getting the accrued duration
func (m *defaultMonitor) Accrued() time.Duration {
	return m.accrued
}

// Reset this monitor
func (m *defaultMonitor) Reset() {
	m.accrued = 0
	m.running = false
	m.paused = false
}

func (m *defaultMonitor) SetError() {
	if m.point != nil {
		m.point.IncError(1)
	}
}
