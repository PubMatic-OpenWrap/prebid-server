package xmlparser

import (
	"testing"
)

func TestStack(t *testing.T) {
	// Test pushing and popping elements
	intStack := stack[int]{}
	intStack.push(1)
	intStack.push(2)
	intStack.push(3)

	// Test len method
	if len := intStack.len(); len != 3 {
		t.Errorf("Expected length 3, got %d", len)
	}

	// Test popping elements
	if popped := intStack.pop(); popped == nil || *popped != 3 {
		t.Errorf("Expected popped value 3, got %v", popped)
	}

	// Test peek method
	if top := intStack.peek(); top == nil || *top != 2 {
		t.Errorf("Expected peeked value 2, got %v", top)
	}

	// Test len method after pops
	if len := intStack.len(); len != 2 {
		t.Errorf("Expected length 2, got %d", len)
	}

	// Test popping the rest of the elements
	if popped := intStack.pop(); popped == nil || *popped != 2 {
		t.Errorf("Expected popped value 2, got %v", popped)
	}
	if popped := intStack.pop(); popped == nil || *popped != 1 {
		t.Errorf("Expected popped value 1, got %v", popped)
	}

	// Test len method after all pops
	if len := intStack.len(); len != 0 {
		t.Errorf("Expected length 0, got %d", len)
	}

	// Test peek method on an empty stack
	if top := intStack.peek(); top != nil {
		t.Errorf("Expected nil peek on an empty stack, got %v", top)
	}
}

func TestEmptyStack(t *testing.T) {
	// Test operations on an empty stack
	emptyStack := stack[int]{}

	// Test popping from an empty stack
	if popped := emptyStack.pop(); popped != nil {
		t.Errorf("Expected popped value 0 from an empty stack, got %v", *popped)
	}

	// Test peeking into an empty stack
	if top := emptyStack.peek(); top != nil {
		t.Errorf("Expected nil peek from an empty stack, got %v", top)
	}

	// Test len method on an empty stack
	if len := emptyStack.len(); len != 0 {
		t.Errorf("Expected length 0 for an empty stack, got %d", len)
	}
}
