package main

func NewStat() *Stat {
	return &Stat{ByMethod: map[string]uint64{}, ByConsumer: map[string]uint64{}}
}

func (s *Stat) SoftReset() {
	s.ByConsumer = map[string]uint64{}
	s.ByMethod = map[string]uint64{}
}
