package chantrick

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
