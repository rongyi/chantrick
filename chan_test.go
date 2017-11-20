package chantrick

import (
	"fmt"
	"log"
	"testing"
	"time"
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

	<-Or(sig(2*time.Hour),
		sig(2*time.Minute),
		sig(3*time.Second),
		sig(5*time.Minute),
	)

	fmt.Printf("done fater %v", time.Since(start))
}

func TestPipe(t *testing.T) {
	f := func(i int) int {
		return i * i
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
	for i := 0; i < 5; i++ {
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

func TestBridge(t *testing.T) {
	genVals := func() <-chan <-chan interface{} {
		chch := make(chan (<-chan interface{}))
		go func() {
			defer close(chch)
			for i := 0; i < 10; i++ {
				s := make(chan interface{}, 1)
				s <- i
				close(s)
				chch <- s
			}
		}()
		return chch
	}
	for v := range Bridge(nil, genVals()) {
		fmt.Printf("%v ", v)
	}
	fmt.Println("")
}

func TestHeartBeat(t *testing.T) {
	done := make(chan interface{})
	time.AfterFunc(10*time.Second, func() { close(done) })

	const timeout = 2 * time.Second
	hb, rets := HeartBeat(done, timeout/2)
	for {
		select {
		case _, ok := <-hb:
			if ok == false {
				return
			}
			fmt.Println("pulse")
		case r, ok := <-rets:
			if ok == false {
				return
			}
			fmt.Printf("results %v \n", r.Second())
		case <-time.After(timeout):
			return
		}
	}
}

func TestWatchdogNormal(t *testing.T) {
	log.SetFlags(log.Ltime | log.LUTC)
	doWork := func(done <-chan interface{}, n time.Duration) <-chan interface{} {
		log.Println("ward: Hello, I'm normal")
		// go func() {
		// 	<-done
		// 	log.Println("ward: Iam halting.")
		// }()

		ch := make(chan interface{})
		go func() {
			for {
				time.Sleep(time.Second * 1)
				log.Println("working...")
				ch <- struct{}{}
			}
		}()
		return ch
	}
	doWorkWithSteward := Watchdog(4*time.Second, doWork)

	done := make(chan interface{})
	time.AfterFunc(9*time.Second, func() {
		log.Println("main: halting steward and ward")
		close(done)
	})
	for range doWorkWithSteward(done, 2*time.Second) {
	}
	log.Println("Done")
}

func TestWatchdogAbnormal(t *testing.T) {
	log.SetFlags(log.Ltime | log.LUTC)
	doWork := func(done <-chan interface{}, n time.Duration) <-chan interface{} {
		log.Println("ward: Hello, I'm bastard, do nothing and halt intentionally")
		go func() {
			<-done
			log.Println("ward: shit, they kill me")
		}()
		return nil
	}
	doWorkWithSteward := Watchdog(4*time.Second, doWork)

	done := make(chan interface{})
	time.AfterFunc(9*time.Second, func() {
		log.Println("main: halting steward and ward")
		close(done)
	})
	for range doWorkWithSteward(done, 2*time.Second) {
	}
	log.Println("Done")
}
