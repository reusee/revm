package revm

type Program []Inst

type Inst struct {
	Op      Op
	Predict func(interface{}) bool
	A, B    int
}

type Op int

const (
	Predict Op = iota
	Ok
	Jump
	Split
)

type Threads struct {
	sparse, dense []int
	n             int
}

func (t *Threads) add(i int) {
	if t.sparse[i] < t.n && t.dense[t.sparse[i]] == i {
		return
	}
	t.dense[t.n] = i
	t.sparse[i] = t.n
	t.n++
}

type Iterator interface {
	Next() interface{}
}

func (p Program) Match(iter Iterator) bool {
	activeThreads := &Threads{make([]int, len(p)), make([]int, len(p)), 0}
	nextThreads := &Threads{make([]int, len(p)), make([]int, len(p)), 0}
	activeThreads.add(0)
	v := iter.Next()
	n := 0
	for v != nil {
		for i := 0; i < activeThreads.n; i++ {
			pc := activeThreads.dense[i]
			inst := p[pc]
			switch inst.Op {
			case Predict:
				if inst.Predict(v) {
					nextThreads.add(pc + 1)
				}
			case Ok:
				return true
			case Jump:
				activeThreads.add(inst.A)
			case Split:
				activeThreads.add(inst.A)
				activeThreads.add(inst.B)
			}
		}
		activeThreads, nextThreads = nextThreads, activeThreads
		nextThreads.n = 0 // clear
		v = iter.Next()
		n++
	}
	for i := 0; i < activeThreads.n; i++ {
		pc := activeThreads.dense[i]
		inst := p[pc]
		switch inst.Op {
		case Ok:
			return true
		case Jump:
			activeThreads.add(inst.A)
		case Split:
			activeThreads.add(inst.A)
			activeThreads.add(inst.B)
		}
	}
	return false
}
