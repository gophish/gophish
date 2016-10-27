package models

func wait(ch chan string, l int, callback func()) {
	n := 0
	for {
		<-ch
		n++
		if n == l {
			callback()
			break
		}
	}
}
