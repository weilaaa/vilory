package vilory

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"syscall"
)

type runFunc func(id, info string)
type endFunC func(id string, err error)

type Worker struct {
	in     io.WriteCloser
	out    io.ReadCloser
	pid    int
	cmd    *exec.Cmd
	master *Master
	stopCh chan struct{}

	IsRunning bool
	//Cond to control worker if running in goroutine
	Cond *sync.Cond

	Id          string
	RunningCall runFunc
	FinishCall  endFunC
}

func NewWorker(id string) *Worker {
	return &Worker{
		Id:     id,
		Cond:   sync.NewCond(&sync.Mutex{}),
		stopCh: make(chan struct{}),
	}
}

// Run should be used when run cmd without pipe
func (r *Worker) Run(cmd string, RunningCall runFunc, FinishCall endFunC, arg ...string) error {
	err := r.Start(cmd, RunningCall, FinishCall, arg...)
	if err != nil {
		return err
	}

	r.waitStop()
	return nil
}

// Start must call DelWork when process end
func (r *Worker) Start(cmd string, RunningCall runFunc, FinishCall endFunC, arg ...string) error {
	r.cmd = exec.Command(cmd, arg...)
	var err error
	if r.in, err = r.cmd.StdinPipe(); err != nil {
		return err
	}

	if r.out, err = r.cmd.StdoutPipe(); err != nil {
		return err
	}

	if err := r.cmd.Start(); err != nil {
		return err
	}

	r.RunningCall = RunningCall
	r.FinishCall = FinishCall
	r.pid = r.cmd.Process.Pid

	if r.master != nil {
		err := r.master.runWorker(r.Id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Worker) bindMaster(m *Master) error {
	if m == nil {
		return errors.New("master is nil")
	}
	r.master = m
	return nil
}

func (r *Worker) wait() error {
	if r.pid == 0 {
		return errors.New("r.pid is 0")
	}

	defer r.in.Close()
	err := r.cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (r *Worker) stop() {
	if r.pid == 0 {
		fmt.Printf("[ERROR] %s has no pid\n", r.Id)
	}

	pro, err := os.FindProcess(r.pid)
	if err != nil {
		fmt.Printf("[WARN] %s has not existed:%s\n", r.Id, err)
	}

	err = pro.Signal(syscall.SIGKILL)
	if err != nil {
		fmt.Printf("[WARN] %s deal with kill signal failed:%s\n", r.Id, err)
	}

	r.stopCh <- struct{}{}
}

func (r *Worker) Input(in string) error {
	if r.in == nil {
		return errors.New("stdinpipe has not set")
	}

	if !strings.Contains(in, "\n") {
		in += "\n"
	}

	retryTimes := 0
	total := 0
	for {
		select {
		case <-time.Tick(100 * time.Millisecond):
			if n, err := r.in.Write([]byte(in)); err != nil {
				return err
			} else {
				total += n
			}
			if total >= len(in) {
				return nil
			}
			retryTimes++
			if retryTimes > 3 {
				return errors.New("write data not enough")
			}
		}
	}
}

func (r *Worker) output() io.ReadCloser {
	return r.out
}

func (r *Worker) waitStop() {
	for {
		time.Sleep(20 * time.Millisecond)
		if !r.IsRunning {
			r.master.Lock()
			delete(r.master.pool, r.Id)
			r.master.Unlock()
			return
		}
	}
}
