package measurement

import (
	"sync"
	"time"
)

type Point struct {
	name                     string
	sactive                  bool
	min, max, average, total time.Duration
	errorCount, count        int
	squareSum                float64
	active, maxActive        int
	calcLock                 sync.Mutex
}

func NewPoint(name string, active bool) *Point {
	return &Point{
		name:     name,
		sactive:  active,
		calcLock: sync.Mutex{},
	}
}

// Name the name of this measure point
func (p *Point) Name() string {
	return p.name
}

func (p *Point) Reset() {
	p.calcLock.Lock()
	defer p.calcLock.Unlock()
	p.min = 0
	p.max = 0
	p.average = 0
	p.total = 0
	p.errorCount = 0
	p.count = 0
	p.squareSum = 0
	p.active = 0
	p.maxActive = 0
}

// Monitor get a new monitor
func (p *Point) Monitor() Monitor {
	if p.sactive {
		return newMonitor(p)
	}
	return &nullMonitor{}
}

func (p *Point) processMonitor(m *defaultMonitor) {
	p.calculateMonitor(m)
}

func (p *Point) calculateMonitor(m *defaultMonitor) {
	p.calcLock.Lock()
	defer p.calcLock.Unlock()

	accrued := m.Accrued()

	p.active--
	p.count++
	p.total += accrued
	p.average = p.total / time.Duration(p.count)
	if accrued > p.max {
		p.max = accrued
	}
	if (accrued < p.min) || (p.min == 0) {
		p.min = accrued
	}
}

func (p *Point) activateMonitor(m *defaultMonitor) {
	p.calcLock.Lock()
	defer p.calcLock.Unlock()
	p.active++
	if p.active > p.maxActive {
		p.maxActive = p.active
	}
}

func (p *Point) Active() int {
	return p.active
}

func (p *Point) Data() Data {
	return Data{
		Name:      p.name,
		Min:       p.calcTime(p.min),
		Max:       p.calcTime(p.max),
		Average:   p.calcTime(p.average),
		Total:     p.calcTime(p.total),
		Count:     p.count,
		MaxActive: p.maxActive,
	}
}

func (p *Point) calcTime(d time.Duration) int64 {
	// converting nano seconds to milli
	return int64(d) / 1000000
}
