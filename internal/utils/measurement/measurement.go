package measurement

import (
	"slices"
	"strings"
	"sync"
)

type Service struct {
	active bool
	plock  sync.Mutex
	points map[string]*Point
}

type Data struct {
	Name      string `json:"name"`
	Min       int64  `json:"min"`
	Max       int64  `json:"max"`
	Average   int64  `json:"average"`
	Total     int64  `json:"total"`
	Count     int    `json:"count"`
	MaxActive int    `json:"maxActive"`
}

func New(active bool) *Service {
	service := Service{
		active: active,
		points: make(map[string]*Point),
		plock:  sync.Mutex{},
	}
	return &service
}

func (s *Service) Start(name string) Monitor {
	p := s.Point(name)
	m := p.Monitor()
	m.Start()
	return m
}

func (s *Service) Point(name string) *Point {
	s.plock.Lock()
	defer s.plock.Unlock()
	p, ok := s.points[name]
	if !ok {
		p = NewPoint(name, s.active)
		s.points[name] = p
	}
	return p
}

func (s *Service) Datas() []Data {
	datas := make([]Data, 0)
	for _, v := range s.points {
		datas = append(datas, v.Data())
	}
	slices.SortFunc(datas, func(d1, d2 Data) int {
		return strings.Compare(d1.Name, d2.Name)
	})
	return datas
}

func (s *Service) Reset() {
	for _, v := range s.points {
		v.Reset()
	}
}
