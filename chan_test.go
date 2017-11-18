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
