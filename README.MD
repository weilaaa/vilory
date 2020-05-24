# vilory
vilory is better modify release of pegent.

## modify
1. fix bug: can't kill process if have running child process
2. optimize: add timeout mechanism
3. optimize: adjust error and log output

## quick start
```go
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
		// see this err a normal info
		fmt.Println(err)
	}
	fmt.Printf("[%s] finished\n", id)
}

func TestVilory(t *testing.T) {
	a := NewMaster()
	var err error

	fmt.Println("worker1 ----------------- start")
	worker1 := a.SetOrGetWorker("worker1")
	err = worker1.Run("ls", runAction, endAction)
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
	a.DelWorker("worker2")


	time.Sleep(15 * time.Second)
	a.DelWorker("worker3")

	if worker1.IsRunning || worker2.IsRunning || worker3.IsRunning {
		t.Error("worker still running")
	}

	fmt.Println("end ------------------------ end")
}

```