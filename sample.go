package main_ow

import "fmt"

type sae struct {
	enabledBidders map[string]struct{}
}

func (s sae) sample() map[string]struct{} {
	return s.enabledBidders
}

func tmo() {
	s := sae{
		enabledBidders: map[string]struct{}{
			"a": {},
			"b": {},
		},
	}

	tmp := s.sample()

	tmp["c"] = struct{}{}

	fmt.Println("old", s.enabledBidders)
	fmt.Println("new", tmp)
}
