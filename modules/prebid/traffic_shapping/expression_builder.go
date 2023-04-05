package trafficshapping

type Expression interface {
	Evaluate(map[string]string) bool
}

type And struct {
	Left, Right Expression
}

func (a And) Evaluate(p map[string]string) bool {
	return a.Left.Evaluate(p) && a.Right.Evaluate(p)
}

type Eq struct {
	Key, Value string
}

func (e Eq) Evaluate(p map[string]string) bool {
	if v, found := p[e.Key]; found {
		return e.Value == v
	}
	return false
}
