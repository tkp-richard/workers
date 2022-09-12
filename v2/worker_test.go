package workers

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	type args struct {
		maxWorker int
	}
	tests := []struct {
		name string
		args args
		want *WorkerPool
	}{
		{
			name: "Test New",
			args: args{
				maxWorker: 5,
			},
			want: New(5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.maxWorker)
			assert.NotNil(t, got, "got is nil")
			assert.Condition(t, func() (success bool) {
				return len(got.workerQueue) == tt.args.maxWorker
			})
			time.Sleep(2 * time.Second)
		})
	}
}

func TestFullFlow_Workers(t *testing.T) {
	got := New(5, Opts{Verbose: true})
	assert.NotNil(t, got, "got is nil")
	got.StoreTask(Task{
		Name: "Test",
		Do: func(w *WorkerPool) {
			fmt.Println("test ya bor")
		},
	})
	got.StoreTask(Task{
		Name: "Test",
		Do: func(w *WorkerPool) {
			fmt.Println("test ya bor")
		},
	})
	time.Sleep(2 * time.Second)
}
