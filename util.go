package pollparty

import "log"

func assert(cond bool) {
	if !cond {
		log.Fatal("assert failed")
	}
}
