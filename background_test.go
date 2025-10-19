package notstd

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestResultWaiterFactory_SingleSubscriber_Success(t *testing.T) {
	resultCh := make(chan Result[int], 1)
	waiter := ResultWaiterFactory(resultCh)

	go func() {
		time.Sleep(10 * time.Millisecond)
		resultCh <- Result[int]{Result: 42}
		close(resultCh)
	}()

	ctx := context.Background()
	res, ok := waiter(ctx)
	if !ok {
		t.Fatalf("expected ok=true, got false")
	}
	if res.Result != 42 {
		t.Fatalf("expected result=42, got %v", res.Result)
	}
}

func TestResultWaiterFactory_SingleSubscriber_Error(t *testing.T) {
	resultCh := make(chan Result[int], 1)
	waiter := ResultWaiterFactory(resultCh)

	go func() {
		time.Sleep(5 * time.Millisecond)
		resultCh <- Result[int]{Error: errors.New("boom")}
		close(resultCh)
	}()

	res, ok := waiter(context.Background())
	if !ok {
		t.Fatalf("expected ok=true, got false")
	}
	if res.Error == nil || res.Error.Error() != "boom" {
		t.Fatalf("expected error=boom, got %v", res.Error)
	}
}

func TestResultWaiterFactory_MultipleSubscribers_AllReceiveSameResult(t *testing.T) {
	resultCh := make(chan Result[string], 1)
	waiter := ResultWaiterFactory(resultCh)

	ctx := context.Background()

	var wg sync.WaitGroup
	results := make([]string, 3)
	for i := range results {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			res, ok := waiter(ctx)
			if !ok {
				t.Errorf("subscriber %d got ok=false", i)
				return
			}
			results[i] = res.Result
		}(i)
	}

	time.Sleep(10 * time.Millisecond)
	resultCh <- Result[string]{Result: "done"}
	close(resultCh)
	wg.Wait()

	for i, v := range results {
		if v != "done" {
			t.Fatalf("subscriber %d expected 'done', got %q", i, v)
		}
	}
}

func TestResultWaiterFactory_SubscriberAfterDone_GetsCachedResult(t *testing.T) {
	resultCh := make(chan Result[int], 1)
	waiter := ResultWaiterFactory(resultCh)

	// Завершаем задачу до подписки
	resultCh <- Result[int]{Result: 99}
	close(resultCh)

	time.Sleep(5 * time.Millisecond) // даём времени горутине обработать результат

	ctx := context.Background()
	res, ok := waiter(ctx)
	if !ok {
		t.Fatalf("expected ok=true, got false")
	}
	if res.Result != 99 {
		t.Fatalf("expected result=99, got %v", res.Result)
	}
}

func TestResultWaiterFactory_ContextCanceledBeforeResult(t *testing.T) {
	resultCh := make(chan Result[int])
	waiter := ResultWaiterFactory(resultCh)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	res, ok := waiter(ctx)
	if ok {
		t.Fatalf("expected ok=false, got true with result=%+v", res)
	}
}

func TestResultWaiterFactory_ChannelClosedWithoutResult(t *testing.T) {
	resultCh := make(chan Result[int])
	waiter := ResultWaiterFactory(resultCh)

	close(resultCh)

	res, ok := waiter(context.Background())
	if ok {
		t.Fatalf("expected ok=false when channel closed without result, got true with %+v", res)
	}
}
