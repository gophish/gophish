package models

func Wait(ch chan interface{}, s int, process func(a interface{})) {
	if s > 0 {
		n := 0
		for {
			process(<-ch)
			n++
			Logger.Println("Size ", s)
			Logger.Println("N ", n)
			if n == s {
				Logger.Println("Finished Waiting")
				break
			}
		}
	}
}
