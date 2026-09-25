package main

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestWithCancel_ParentCancelPropagates(t *testing.T) {
	parent, cancelParent := WithCancel(nil)
	child, _ := WithCancel(parent)
	grandchild, _ := WithCancel(child)

	cancelParent()

	select {
	case <-parent.Done():
	case <-time.After(time.Second):
		t.Fatal("parent was not canceled")
	}

	select {
	case <-child.Done():
	case <-time.After(time.Second):
		t.Fatal("child was not canceled")
	}

	select {
	case <-grandchild.Done():
	case <-time.After(time.Second):
		t.Fatal("grandchild was not canceled")
	}

	if parent.Err() != ErrCanceled {
		t.Fatalf("parent.Err() = %v, want %v", parent.Err(), ErrCanceled)
	}

	if child.Err() != ErrCanceled {
		t.Fatalf("child.Err() = %v, want %v", child.Err(), ErrCanceled)
	}

	if grandchild.Err() != ErrCanceled {
		t.Fatalf("grandchild.Err() = %v, want %v", grandchild.Err(), ErrCanceled)
	}
}

func TestWithCancel_ChildCancelDoesNotCancelParent(t *testing.T) {
	parent, _ := WithCancel(nil)
	child, cancelChild := WithCancel(parent)

	cancelChild()

	select {
	case <-child.Done():
		// expected
	case <-time.After(time.Second):
		t.Fatal("child was not canceled")
	}

	// Parent must still be alive.
	select {
	case <-parent.Done():
		t.Fatal("parent was canceled when child was canceled")
	default:
	}

	if parent.Err() != nil {
		t.Fatalf("parent.Err() = %v, want nil", parent.Err())
	}

	if child.Err() != ErrCanceled {
		t.Fatalf("child.Err() = %v, want %v", child.Err(), ErrCanceled)
	}
}

func TestWithCancel_CancelTwiceDoesNotPanic(t *testing.T) {
	ctx, cancel := WithCancel(nil)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("cancel twice panicked: %v", r)
		}
	}()

	cancel()
	cancel()

	select {
	case <-ctx.Done():
		// expected
	case <-time.After(time.Second):
		t.Fatal("context was not canceled")
	}
}

func TestWithCancel_AlreadyCanceledParentCancelsChildImmediately(t *testing.T) {
	parent, cancelParent := WithCancel(nil)

	cancelParent()

	child, _ := WithCancel(parent)

	select {
	case <-child.Done():
		// expected
	case <-time.After(time.Second):
		t.Fatal("child was not immediately canceled")
	}

	if child.Err() != ErrCanceled {
		t.Fatalf("child.Err() = %v, want %v", child.Err(), ErrCanceled)
	}
}

func TestWithCancel_ChildRemovedFromParentAfterCancel(t *testing.T) {
	parent, _ := WithCancel(nil)
	child, cancelChild := WithCancel(parent)

	cancelChild()

	parent.mu.Lock()
	_, exists := parent.children[child]
	parent.mu.Unlock()

	if exists {
		t.Fatal("child still exists in parent.children after cancellation")
	}
}

func TestWithCancel_ParentRemovesCanceledChild(t *testing.T) {
	parent, _ := WithCancel(nil)
	child1, cancelChild1 := WithCancel(parent)
	child2, _ := WithCancel(parent)

	cancelChild1()

	parent.mu.Lock()
	defer parent.mu.Unlock()

	if _, exists := parent.children[child1]; exists {
		t.Fatal("canceled child still exists in parent.children")
	}

	if _, exists := parent.children[child2]; !exists {
		t.Fatal("active child was removed from parent.children")
	}
}

func TestWithCancel_MultipleChildren(t *testing.T) {
	parent, cancelParent := WithCancel(nil)

	const childCount = 100

	children := make([]*ctxt, childCount)

	for i := 0; i < childCount; i++ {
		children[i], _ = WithCancel(parent)
	}

	cancelParent()

	for i, child := range children {
		select {
		case <-child.Done():
		case <-time.After(time.Second):
			t.Fatalf("child %d was not canceled", i)
		}

		if child.Err() != ErrCanceled {
			t.Fatalf(
				"child %d Err() = %v, want %v",
				i,
				child.Err(),
				ErrCanceled,
			)
		}
	}
}

func TestWithCancel_DeepPropagation(t *testing.T) {
	root, cancelRoot := WithCancel(nil)

	const depth = 100

	contexts := make([]*ctxt, depth)
	contexts[0] = root

	for i := 1; i < depth; i++ {
		contexts[i], _ = WithCancel(contexts[i-1])
	}

	cancelRoot()

	for i, ctx := range contexts {
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
			t.Fatalf("context at depth %d was not canceled", i)
		}
	}
}

func TestWithCancel_ErrBeforeCancelIsNil(t *testing.T) {
	ctx, _ := WithCancel(nil)

	if ctx.Err() != nil {
		t.Fatalf("Err() before cancellation = %v, want nil", ctx.Err())
	}

	select {
	case <-ctx.Done():
		t.Fatal("Done() is already closed")
	default:
	}
}

func TestWithCancel_DoneIsClosedAfterCancel(t *testing.T) {
	ctx, cancel := WithCancel(nil)

	cancel()

	select {
	case <-ctx.Done():
		// expected
	default:
		t.Fatal("Done() is not closed after cancellation")
	}
}

func TestWithCancel_ConcurrentCancel(t *testing.T) {
	ctx, cancel := WithCancel(nil)

	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			cancel()
		}()
	}

	wg.Wait()

	select {
	case <-ctx.Done():
		// expected
	case <-time.After(time.Second):
		t.Fatal("context was not canceled")
	}

	if ctx.Err() != ErrCanceled {
		t.Fatalf("ctx.Err() = %v, want %v", ctx.Err(), ErrCanceled)
	}
}

func TestWithCancel_ConcurrentErr(t *testing.T) {
	ctx, cancel := WithCancel(nil)

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()

		for i := 0; i < 1000; i++ {
			_ = ctx.Err()
		}
	}()

	go func() {
		defer wg.Done()

		cancel()
	}()

	wg.Wait()

	if ctx.Err() != ErrCanceled {
		t.Fatalf("ctx.Err() = %v, want %v", ctx.Err(), ErrCanceled)
	}
}

func TestWithCancel_ConcurrentChildCreationAndParentCancel(t *testing.T) {
	parent, cancelParent := WithCancel(nil)

	const childCount = 100

	children := make([]*ctxt, childCount)

	var wg sync.WaitGroup

	wg.Add(childCount)

	for i := 0; i < childCount; i++ {
		i := i

		go func() {
			defer wg.Done()

			children[i], _ = WithCancel(parent)
		}()
	}

	cancelParent()
	wg.Wait()

	for i, child := range children {
		select {
		case <-child.Done():
			// expected
		case <-time.After(time.Second):
			t.Fatalf("child %d was not canceled", i)
		}
	}
}

func TestWithCancel_NoGoroutinePerChild(t *testing.T) {
	parent, _ := WithCancel(nil)

	before := runtime.NumGoroutine()

	const childCount = 1000

	for i := 0; i < childCount; i++ {
		WithCancel(parent)
	}

	// Give the scheduler a moment if something spawned goroutines.
	time.Sleep(100 * time.Millisecond)

	after := runtime.NumGoroutine()

	// Creating 1000 children should NOT create 1000 waiting goroutines.
	//
	// We allow a small amount of noise from the test/runtime itself.
	if after-before > 20 {
		t.Fatalf(
			"too many goroutines created: before=%d after=%d",
			before,
			after,
		)
	}
}

func TestWithCancel_NoGoroutineLeakAfterChildCancel(t *testing.T) {
	parent, _ := WithCancel(nil)

	before := runtime.NumGoroutine()

	const childCount = 1000

	cancels := make([]func(), childCount)

	for i := 0; i < childCount; i++ {
		_, cancels[i] = WithCancel(parent)
	}

	for _, cancel := range cancels {
		cancel()
	}

	time.Sleep(100 * time.Millisecond)

	after := runtime.NumGoroutine()

	if after-before > 20 {
		t.Fatalf(
			"possible goroutine leak: before=%d after=%d",
			before,
			after,
		)
	}
}

func TestWithCancel_ParentCancelCancelsAllDescendants(t *testing.T) {
	root, cancelRoot := WithCancel(nil)

	child1, _ := WithCancel(root)
	child2, _ := WithCancel(root)

	grandchild1, _ := WithCancel(child1)
	grandchild2, _ := WithCancel(child2)

	greatGrandchild, _ := WithCancel(grandchild1)

	cancelRoot()

	contexts := []*ctxt{
		root,
		child1,
		child2,
		grandchild1,
		grandchild2,
		greatGrandchild,
	}

	for i, ctx := range contexts {
		select {
		case <-ctx.Done():
		case <-time.After(time.Second):
			t.Fatalf("context %d was not canceled", i)
		}
	}
}
