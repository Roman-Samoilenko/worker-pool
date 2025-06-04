package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type WorkerPool struct {
	workers  map[int]*Worker
	taskCh   chan string
	addCh    chan struct{}
	removeCh chan int
	wg       sync.WaitGroup
	nextID   int64
}

type Worker struct {
	id     int
	taskCh chan string
	quit   chan struct{}
}

func NewWorkerPool() *WorkerPool {
	return &WorkerPool{
		workers:  make(map[int]*Worker),
		taskCh:   make(chan string),
		addCh:    make(chan struct{}),
		removeCh: make(chan int),
	}
}

// Start запускает горутину для отслеживания сигналов добавления и удаления воркера
func (wp *WorkerPool) Start() {
	go func() {
		for {
			select {
			case <-wp.addCh:
				wp.addWorker()
			case id := <-wp.removeCh:
				wp.removeWorker(id)
			}
		}
	}()
}

func (wp *WorkerPool) AddWorker() {
	wp.addCh <- struct{}{}
}

func (wp *WorkerPool) RemoveWorker(id int) {
	wp.removeCh <- id
}

// addWorker создает и запускает воркер
func (wp *WorkerPool) addWorker() {
	id := int(atomic.AddInt64(&wp.nextID, 1))
	worker := &Worker{
		id:     id,
		taskCh: wp.taskCh,
		quit:   make(chan struct{}),
	}
	wp.workers[worker.id] = worker
	go worker.run(wp)
}

// removeWorker завершает работу воркера
func (wp *WorkerPool) removeWorker(id int) {
	if worker, exists := wp.workers[id]; exists {
		close(worker.quit)
		delete(wp.workers, id)
	}
}

// Submit отправляет задачу
func (wp *WorkerPool) Submit(task string) {
	wp.wg.Add(1)
	wp.taskCh <- task
}

// run выполняет основной цикл воркера
func (w *Worker) run(wp *WorkerPool) {
	for {
		select {
		case task, ok := <-w.taskCh:
			if !ok {
				return
			}
			fmt.Printf("Воркер %d обрабатывает строку: %s\n", w.id, task)
			// имитация тяжелой работы
			time.Sleep(1 * time.Second)
			wp.wg.Done()

		case <-w.quit:
			return
		}
	}
}

// generateRandomTask генерирует случайную строку
func generateRandomTask() string {
	const letters = "qwertyuiop[]';lkjhgfdsazxcvbnm,./"
	b := make([]byte, 7)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	rand.Seed(time.Now().UnixNano())
	wp := NewWorkerPool()
	wp.Start()

	var numWorkers, numTasks int
	fmt.Print("Введите количество воркеров: ")
	fmt.Scan(&numWorkers)
	fmt.Print("Введите количество задач: ")
	fmt.Scan(&numTasks)

	for range numWorkers {
		wp.AddWorker()
	}

	for range numTasks {
		wp.Submit(generateRandomTask())
	}

	wp.wg.Wait()
}
