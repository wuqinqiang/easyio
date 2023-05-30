package easyio

type Manger struct {
	num int
	//balance
	polls []*Poller
}

func NewManger(e *Engine, number int) (*Manger, error) {
	m := new(Manger)
	m.num = number

	for i := 0; i < number; i++ {
		p, err := NewPoller(e)
		p.index = i
		if err != nil {
			_ = m.Stop()
			return nil, err
		}
		m.polls = append(m.polls, p)
	}

	m.init()
	return m, nil
}

func (m *Manger) init() {
	for _, poller := range m.polls {
		p := poller
		go p.Wait()
	}
}

func (m *Manger) Stop() error {
	for _, poller := range m.polls {
		_ = poller.Close()
	}
	return nil
}

func (m *Manger) Pick(fd int) *Poller {
	return m.polls[fd%m.num]
}
