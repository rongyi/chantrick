package chantrick

import (
	"log"
	"sync"
	"time"
)

// Or just like circute or for channels
func Or(chans ...<-chan interface{}) <-chan interface{} {
	switch len(chans) {
	case 0:
		return nil
	case 1:
		return chans[0]
	}

	orDone := make(chan interface{})
	go func() {
		defer close(orDone)
		switch len(chans) {
		case 2:
			select {
			case <-chans[0]:
			case <-chans[1]:
			}
		default:
			select {
			case <-chans[0]:
			case <-chans[1]:
			case <-chans[2]:
			case <-Or(append(chans[3:], orDone)...):
			}
		}
	}()

	return orDone
}

// GenChan generate a channel based on raw input
func GenChan(done <-chan interface{}, ints ...int) <-chan int {
	ichan := make(chan int)
	go func() {
		defer close(ichan)
		for _, i := range ints {
			select {
			case <-done:
				return
			case ichan <- i:
			}
		}
	}()
	return ichan
}

// MapInt is a mapper of int function
type MapInt func(i int) int

// ConsumeChan is pipeline for
func ConsumeChan(done <-chan interface{}, ichan <-chan int, f MapInt) <-chan int {
	mchan := make(chan int)
	go func() {
		defer close(mchan)
		for i := range ichan {
			select {
			case <-done:
				return
			case mchan <- f(i):
			}
		}
	}()
	return mchan
}

// Repeat just like generate but for ever
func Repeat(done <-chan interface{}, valus ...interface{}) <-chan interface{} {
	vs := make(chan interface{})
	go func() {
		defer close(vs)
		for {
			for _, v := range valus {
				select {
				case <-done:
					return
				case vs <- v:
				}
			}
		}
	}()

	return vs
}

// Take take a limited number of element from vs
func Take(done <-chan interface{}, vs <-chan interface{}, count int) <-chan interface{} {
	tchan := make(chan interface{})
	go func() {
		defer close(tchan)
		for i := 0; i < count; i++ {
			select {
			case <-done:
				return
			case tchan <- <-vs:
			}
		}
	}()
	return tchan
}

// Fanin 收口飞起的goroutine
func Fanin(done <-chan interface{}, chans ...<-chan interface{}) <-chan interface{} {
	var wg sync.WaitGroup
	muChan := make(chan interface{})
	doPlex := func(c <-chan interface{}) {
		defer wg.Done()
		for i := range c {
			select {
			case <-done:
				return
			case muChan <- i:
			}
		}
	}

	wg.Add(len(chans))
	for _, c := range chans {
		go doPlex(c)
	}
	go func() {
		wg.Wait()
		close(muChan)
	}()
	return muChan
}

// OrDone wrap a done channel
func OrDone(done, c <-chan interface{}) <-chan interface{} {
	vs := make(chan interface{})
	go func() {
		defer close(vs)
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if ok == false {
					return
				}
				select {
				case vs <- v:
				case <-done:
				}
			}
		}
	}()

	return vs
}

// Tee just like tee command for channel
func Tee(done, in <-chan interface{}) (_, _ <-chan interface{}) {
	out1 := make(chan interface{})
	out2 := make(chan interface{})

	go func() {
		defer close(out1)
		defer close(out2)
		for val := range OrDone(done, in) {
			var out1, out2 = out1, out2
			for i := 0; i < 2; i++ {
				select {
				case <-done:
				case out1 <- val:
					// to make out1 block
					out1 = nil
				case out2 <- val:
					// to make out2 block
					out2 = nil
				}
			}
		}
	}()
	return out1, out2
}

// Bridge link multiple channel to one
func Bridge(done <-chan interface{}, chch <-chan <-chan interface{}) <-chan interface{} {
	vs := make(chan interface{})
	go func() {
		defer close(vs)
		for {
			var stream <-chan interface{}
			select {
			case <-done:
				return
			case maybe, ok := <-chch:
				if ok == false {
					return
				}
				stream = maybe
			}
			for val := range OrDone(done, stream) {
				select {
				case vs <- val:
				case <-done:
				}
			}
		}
	}()
	return vs
}

// HeartBeat
func HeartBeat(done <-chan interface{}, pulseInterval time.Duration) (<-chan interface{}, <-chan time.Time) {
	hb := make(chan interface{})
	rets := make(chan time.Time)
	go func() {
		defer close(hb)
		defer close(rets)

		pulse := time.Tick(pulseInterval)
		workGen := time.Tick(2 * pulseInterval)

		sendPulse := func() {
			select {
			case hb <- struct{}{}:
			default:
			}
		}

		sendResult := func(r time.Time) {
			for {
				select {
				case <-done:
					return
				// 保证正常逻辑的同时心跳也正常走
				case <-pulse:
					sendPulse()
				case rets <- r:
					return
				}
			}
		}

		for {
			select {
			case <-done:
				return
			case <-pulse:
				sendPulse()
			case r := <-workGen:
				sendResult(r)
			}
		}
	}()

	return hb, rets
}

type Worker func(done <-chan interface{}, normalInterval time.Duration) <-chan interface{}

func Watchdog(timeout time.Duration, work Worker) Worker {
	return func(done <-chan interface{}, normalInterval time.Duration) <-chan interface{} {
		hb := make(chan interface{})
		go func() {
			defer close(hb)
			var wardDone chan interface{}
			var wardHeartbeat <-chan interface{}

			startWard := func() {
				wardDone = make(chan interface{})
				wardHeartbeat = work(Or(wardDone, done), normalInterval)
			}
			startWard()
			pulse := time.Tick(normalInterval)
		monitorLoop:
			for {
				timeoutSignal := time.After(timeout)
				for {
					select {
					case <-pulse:
						select {
						case hb <- struct{}{}:
						default:
						}
					case <-wardHeartbeat:
						continue monitorLoop
					case <-timeoutSignal:
						close(wardDone)
						log.Println("start a new one")
						startWard()
						continue monitorLoop
					case <-done:
						return
					}
				}
			}
		}()
		return hb
	}
}
