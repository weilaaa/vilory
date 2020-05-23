package vilory

import (
	"fmt"
	"testing"
	"time"
)

func runAction(id, str string) {
	fmt.Printf("[%s] str=%s\n", id, str)
}

func endAction(id string, err error) {
	if err != nil {
		panic(err)
	}
	fmt.Printf("[%s] finished\n", id)
}

func TestVilory(t *testing.T) {
	a := NewMaster()
	var err error

	fmt.Println("worker1 ----------------- start")
	worker1 := a.SetOrGetWorker("worker1")
	err = worker1.Start("ls", runAction, endAction)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(3 * time.Second)
	err = a.DelWorker("worker1")
	if err != nil {
		t.Error(err)
	}

	worker2 := a.SetOrGetWorker("worker2")
	go func() {
		fmt.Println("worker2 ----------------- start")
		err = worker2.Start("bash", runAction, endAction)
		if err != nil {
			t.Error(err)
		}
		err = worker2.Input("go run ./example/main.go")
		if err != nil {
			t.Error(err)
		}
	}()

	worker3 := a.SetOrGetWorker("worker3")
	go func() {
		fmt.Println("worker3 ----------------- start")
		err = worker3.Start("bash", runAction, endAction)
		if err != nil {
			t.Error(err)
		}
		err = worker3.Input("go run ./example/main.go")
		if err != nil {
			t.Error(err)
		}
	}()

	time.Sleep(15 * time.Second)
	err = a.DelWorker("worker2")
	if err != nil {
		t.Error(err)
	}

	time.Sleep(30 * time.Second)
	err = a.DelWorker("worker3")
	if err != nil {
		t.Error(err)
	}

	if worker1.IsRunning || worker2.IsRunning || worker3.IsRunning {
		t.Error("worker still running")
	}

	fmt.Println("end ------------------------ end")
}
