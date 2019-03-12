package main

import (
	"fmt"
	"time"
)

type Struct struct {
	Name string
}

var chans map[int]chan *Struct

func main() {
	chans = make(map[int]chan *Struct, 100)

	chans[1] = make(chan *Struct, 10)

	go func() {
		for {
			select {
			case msg := <-chans[1]:
				fmt.Println("leu")
				fmt.Println(msg)
			default:
				//			fmt.Println("no recv")
			}
		}
	}()

	go func() {
		for {
			select {
			case msg := <-chans[1]:
				fmt.Println("leu2")
				fmt.Println(msg)
			default:
				//			fmt.Println("no recv")
			}
		}
	}()

	select {
	case chans[1] <- &Struct{Name: "gustavo"}:
		fmt.Println("escreveu")
	default:
		fmt.Println("no message sent")
	}

	select {
	case chans[1] <- &Struct{Name: "gustavo"}:
		fmt.Println("escreveu")
	default:
		fmt.Println("no message sent")
	}

	select {
	case chans[1] <- &Struct{Name: "gustavo"}:
		fmt.Println("escreveu")
	default:
		fmt.Println("no message sent")
	}

	select {
	case chans[1] <- &Struct{Name: "gustavo"}:
		fmt.Println("escreveu")
	default:
		fmt.Println("no message sent")
	}

	time.Sleep(time.Minute)
}
