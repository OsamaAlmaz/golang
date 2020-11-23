package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	NumRoutines = 3
	NumRequests = 1000
)

//global semaphores monitoring the number of routines
var semRout = make(chan int, NumRoutines)

// global semaphores monitoring the number of console

var semDisp = make(chan int, 1)

//wait group to ensure that main does not exit until all done

var wgRout sync.WaitGroup
var wgDisp sync.WaitGroup

//Structure

type Task struct {
	a, b float32
	disp chan float32
}

func solve(t *Task) {
	time.Sleep(15 * time.Second)
	sum := t.a + t.b
	t.disp <- sum
	wgRout.Done()
}

func handleReq(t *Task) {
	solve(t)
	//function that acts intermediare between computer server and solve
}

func ComputerServer() chan *Task {
	c := make(chan *Task)
	for i := 0; i < NumRoutines; i++ {
		semRout <- 1
		wgRout.Add(1)
		go func() {
			//c := make(chan *Task)
			//fmt.Println("This is the begining of the computer server")
			a := <-c
			handleReq(a)
			time.Sleep(time.Second)
			<-semRout
		}()

	}

	return c
}
func DisplayServer() chan float32 {
	c := make(chan float32)
	semDisp <- 1
	go func() {
		//display in the channel.
		wgDisp.Add(1)
		a := <-c
		fmt.Println("The sum is", a)
		wgDisp.Done()
		<-semDisp
	}()
	wgDisp.Wait()
	return c
}

func main() {
	dispChan := DisplayServer()
	reqChan := ComputerServer()
	for {
		var a, b float32
		//make sure to use semDisp
		//...

		fmt.Println("Enter two numbers")
		fmt.Scanf("%f %f \n", &a, &b)
		fmt.Println("%f %f \n", a, b)
		if a == 0 && b == 0 {
			break
		}
		//send to SemRout the stuff that you need to add.
		//wg.Wait() //I think.
		var c Task
		c.a = a
		c.b = b
		d := &c
		reqChan <- d
		//communicate with the display to display the result.

		for i := range reqChan {
			task := i
			e := <-task.disp
			//to print the value
			dispChan <- e
		}
		//create task and send to computer server
		//......
		time.Sleep(1e9)

	}
	//Don't exit until all it is done.
	wgRout.Wait()
	wgDisp.Wait()
}
