package measurement

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	PointName = "super-duper-messung"
)

func TestInactive(t *testing.T) {
	ast := assert.New(t)
	s := New(false)

	ast.NotNil(s)
	ast.False(s.active)

	m := s.Start(PointName)
	ast.NotNil(m)
	time.Sleep(1 * time.Second)
	m.Stop()
	d := m.Accrued()
	ast.Equal(0*time.Second, d)

	m = s.Start(PointName)
	ast.NotNil(m)
	time.Sleep(1 * time.Second)
	m.Stop()

	p := s.Point(PointName)
	ast.NotNil(p)
	dat := p.Data()
	ast.NotNil(dat)
	ast.Equal(0, dat.Count)
	ast.Equal(PointName, dat.Name)
	ast.Equal(int64(0), dat.Average)
	ast.LessOrEqual(int64(0), dat.Min)
	ast.LessOrEqual(int64(0), dat.Max)
	ast.Equal(dat.Max, dat.Min)

	dats := s.Datas()
	ast.Equal(1, len(dats))
	ast.Equal(PointName, dats[0].Name)
}

func TestSimple(t *testing.T) {
	ast := assert.New(t)
	s := New(true)

	ast.NotNil(s)
	ast.True(s.active)

	m := s.Start(PointName)
	ast.NotNil(m)
	time.Sleep(1 * time.Second)
	m.Stop()
	d := m.Accrued()
	ast.InEpsilon(1*time.Second, d, 0.01)

	m = s.Start(PointName)
	ast.NotNil(m)
	time.Sleep(1 * time.Second)
	m.Stop()

	p := s.Point(PointName)
	ast.NotNil(p)
	dat := p.Data()
	ast.NotNil(dat)
	ast.Equal(2, dat.Count)
	ast.Equal(PointName, dat.Name)
	ast.InEpsilon(int64(1000), dat.Average, 0.01)
	ast.LessOrEqual(int64(1000), dat.Min)
	ast.LessOrEqual(int64(1000), dat.Max)
	ast.GreaterOrEqual(dat.Max, dat.Min)

	dats := s.Datas()
	ast.Equal(1, len(dats))
	ast.Equal(PointName, dats[0].Name)
}

func TestReset(t *testing.T) {
	ast := assert.New(t)
	s := New(true)

	ast.NotNil(s)
	ast.True(s.active)

	m := s.Start(PointName)
	ast.NotNil(m)
	time.Sleep(1 * time.Second)
	m.Stop()
	d := m.Accrued()
	ast.InEpsilon(1*time.Second, d, 0.01)

	s.Reset()

	p := s.Point(PointName)
	ast.NotNil(p)
	dat := p.Data()
	ast.NotNil(dat)
	ast.Equal(0, dat.Count)
	ast.Equal(PointName, dat.Name)
	ast.Equal(int64(0), dat.Average)
	ast.Equal(int64(0), dat.Min)
	ast.Equal(int64(0), dat.Max)
}
