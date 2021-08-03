/**
 * @author Jose Nidhin
 */
package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

//
func main() {
	wg := &sync.WaitGroup{}
	queue := make(chan int)
	shutdownNotifier := make(chan int)

	startPool(queue, shutdownNotifier, wg)

	wg.Add(1)
	go signalHandler(shutdownNotifier, wg)

	wg.Add(1)
	go startScheduler(queue, shutdownNotifier, wg)

	wg.Wait()
	close(queue)
}

//
func startPool(queue, shutdownNotifier <-chan int, wg *sync.WaitGroup) {
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go worker(i, queue, shutdownNotifier, wg)
	}
}

//
func worker(id int, queue, shutdownNotifier <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("Worker %d - starting \n", id)

	for {
		select {
		case work := <-queue:
			fmt.Printf("Worker %d - Processing work %d \n", id, work)
		case <-shutdownNotifier:
			fmt.Printf("Worker %d - Shutting down \n", id)
			return
		}
	}
}

//
func startScheduler(queue chan<- int, shutdownNotifier <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("Scheduler - starting")

	tick := time.NewTicker(time.Second)

	for {
		select {
		case <-shutdownNotifier:
			fmt.Println("Scheduler - Shutting down")
			return
		case <-tick.C:
			queue <- rand.Intn(1000)
		}
	}
}

//
func signalHandler(shutdownNotifier chan int, wg *sync.WaitGroup) {
	defer func() {
		fmt.Println("Graceful shutdown initialised")
	}()
	defer wg.Done()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, syscall.SIGTERM)

	select {
	case <-shutdownNotifier:
		return
	case <-sigChan:
		close(shutdownNotifier)
	}
}
