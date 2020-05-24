package vilory

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

const waitTimeout = 30 * time.Second

type Master struct {
	sync.Mutex
	pool map[string]*Worker
}

func NewMaster() *Master {
	return &Master{
		pool: make(map[string]*Worker),
	}
}

func (m *Master) SetOrGetWorker(id string) *Worker {
	m.Lock()
	defer m.Unlock()
	if m.pool == nil {
		m.pool = make(map[string]*Worker)
	}
	if _, ok := m.pool[id]; !ok {
		m.pool[id] = NewWorker(id)
		m.pool[id].bindMaster(m)
	}

	return m.pool[id]
}

func (m *Master) runWorker(id string) error {
	m.Lock()
	defer m.Unlock()
	if m.pool == nil {
		return errors.New("worker pool is nil")
	}
	worker, ok := m.pool[id]
	if !ok {
		return errors.New("can't find worker id=" + id)
	}

	worker.IsRunning = true

	go func() {
		doneCh := make(chan error)
		reader := bufio.NewReader(worker.output())

		go func() {
			<-worker.stopCh
			worker.in.Close()
			worker.out.Close()
		}()

		var buf bytes.Buffer
		for {
			line, isPrefix, err := reader.ReadLine()
			if len(line) > 0 {
				buf.Write(line)
				if !isPrefix {
					if worker.RunningCall != nil {
						worker.RunningCall(id, buf.String())
					}
					buf.Reset()
				}
			}

			if err == io.EOF || err != nil {
				break
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
		defer cancel()

		// wait process release resource
		go func() { doneCh <- worker.wait() }()

		for {
			select {
			case <-time.Tick(20 * time.Millisecond):
			case <-ctx.Done():
				panic("wait process release timeout")
			case err := <-doneCh:
				if worker.FinishCall != nil {
					// wait error should be deal upstairs
					worker.FinishCall(id, err)
				}
				worker.IsRunning = false
				m.Lock()
				delete(m.pool, worker.Id)
				m.Unlock()
				return
			}
		}
	}()

	return nil
}

// DelWorker return until worker process release resource and be killed
func (m *Master) DelWorker(id string) {
	m.Lock()
	worker, ok := m.pool[id]
	// must unlock here to avoid deadlock
	m.Unlock()
	if !ok {
		fmt.Printf("[info] %s has already exited\n", id)
		return
	}

	worker.stop()
	worker.waitStop()
}
