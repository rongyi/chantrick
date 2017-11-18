package chantrick
import (
	"testing"
	"time"
	"fmt"
)

func TestOr(t *testing.T) {
	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()

	<-Or(sig(2 * time.Hour),
		sig(2 * time.Minute),
		sig(3 * time.Second),
		sig(5 * time.Minute),
	)

	fmt.Printf("done fater %v", time.Since(start))
}

func TestPipe(t *testing.T) {
	f := func(i int) int {
		return i * i;
	}
	for i := range ConsumeChan(nil, GenChan(nil, 1, 2, 3, 4, 5), f) {
		fmt.Printf("%d ", i)
	}
	fmt.Println("Done")
}

func TestRepeatAndTake(t *testing.T) {
	for i := range Take(nil, Repeat(nil, 1, 2, 3, 4), 5) {
		fmt.Printf("%d ", i)
	}
	fmt.Println("")
}
