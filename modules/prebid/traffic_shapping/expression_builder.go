package trafficshapping

import "fmt"

type Expression interface {
	Evaluate(map[string]string) bool
	GetName() string
}

type And struct {
	Left, Right Expression
}

func (a And) Evaluate(p map[string]string) bool {
	return a.Left.Evaluate(p) && a.Right.Evaluate(p)
}

func (a And) GetName() string {
	if a.Left == nil || a.Right == nil {
		return "And"
	}
	return fmt.Sprintf("(%v And %v)", a.Left.GetName(), a.Right.GetName())
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

func (e Eq) GetName() string {
	return fmt.Sprintf("(%v = %v)", e.Key, e.Value)
}
