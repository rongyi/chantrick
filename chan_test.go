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

func TestFanIn(t *testing.T) {
	done := make(chan interface{})
	chans := make([]<-chan interface{}, 4)
	for i := 0; i < len(chans); i++ {
		chans[i] = Take(done, Repeat(done, 1, 2, 3, 4, 5), 1000)
	}
	for i := range Fanin(done, chans...) {
		fmt.Printf("%d ", i)
	}

	fmt.Println("")
}

func TestOrDone(t *testing.T) {
	ichan := make(chan interface{}, 5)
	for i  := 0; i < 5; i++ {
		ichan <- i
	}
	close(ichan)
	for v := range OrDone(nil, ichan) {
		fmt.Println(v)
	}
}

func TestTee(t *testing.T) {
	done := make(chan interface{})
	defer close(done)

	out1, out2 := Tee(done, Take(done, Repeat(done, 1, 2), 4))
	for val1 := range out1 {
		fmt.Printf("out1: %v, out2 %v\n", val1, <-out2)
	}
}
