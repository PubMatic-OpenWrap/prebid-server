package floors

import (
	"reflect"
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestFetchQueueLen(t *testing.T) {
	tests := []struct {
		name string
		fq   FetchQueue
		want int
	}{
		{
			name: "Queue is empty",
			fq:   make(FetchQueue, 0),
			want: 0,
		},
		{
			name: "Queue is of lenght 1",
			fq:   make(FetchQueue, 1),
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fq.Len(); got != tt.want {
				t.Errorf("FetchQueue.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchQueueLess(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		fq   FetchQueue
		args args
		want bool
	}{
		{
			name: "first fetchperiod is less than second",
			fq:   FetchQueue{&FetchInfo{FetchPeriod: 10}, &FetchInfo{FetchPeriod: 20}},
			args: args{i: 0, j: 1},
			want: true,
		},
		{
			name: "first fetchperiod is greater than second",
			fq:   FetchQueue{&FetchInfo{FetchPeriod: 30}, &FetchInfo{FetchPeriod: 10}},
			args: args{i: 0, j: 1},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fq.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("FetchQueue.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchQueueSwap(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		fq   FetchQueue
		args args
	}{
		{
			name: "Swap two elements at index i and j",
			fq:   FetchQueue{&FetchInfo{FetchPeriod: 30}, &FetchInfo{FetchPeriod: 10}},
			args: args{i: 0, j: 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fInfo1, fInfo2 := tt.fq[0], tt.fq[1]
			tt.fq.Swap(tt.args.i, tt.args.j)
			assert.Equal(t, fInfo1, tt.fq[1], "elements are not swapped")
			assert.Equal(t, fInfo2, tt.fq[0], "elements are not swapped")
		})
	}
}

func TestFetchQueuePush(t *testing.T) {
	type args struct {
		element interface{}
	}
	tests := []struct {
		name string
		fq   *FetchQueue
		args args
	}{
		{
			name: "Push element to queue",
			fq:   &FetchQueue{},
			args: args{element: &FetchInfo{FetchPeriod: 10}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fq.Push(tt.args.element)
			q := *tt.fq
			assert.Equal(t, q[0], &FetchInfo{FetchPeriod: 10})
		})
	}
}

func TestFetchQueuePop(t *testing.T) {
	tests := []struct {
		name string
		fq   *FetchQueue
		want interface{}
	}{
		{
			name: "Pop element from queue",
			fq:   &FetchQueue{&FetchInfo{FetchPeriod: 10}},
			want: &FetchInfo{FetchPeriod: 10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fq.Pop(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchQueue.Pop() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFetchQueueTop(t *testing.T) {
	tests := []struct {
		name string
		fq   *FetchQueue
		want *FetchInfo
	}{
		{
			name: "Get top element from queue",
			fq:   &FetchQueue{&FetchInfo{FetchPeriod: 20}},
			want: &FetchInfo{FetchPeriod: 20},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fq.Top(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchQueue.Top() = %v, want %v", got, tt.want)
			}
		})
	}
}
