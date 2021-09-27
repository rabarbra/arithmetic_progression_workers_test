package workerstore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type WorkerTestSuite struct {
	suite.Suite
	ws *WorkerStore
}

func (suite *WorkerTestSuite) SetupTest() {
	suite.ws = NewWorkerStore(2)
}

func (suite *WorkerTestSuite) TestWorker() {
	tasks := []Task{
		{N: 1, D: 0, N1: 0, I: 1, TTL: 1},
		{N: 1, D: 0, N1: 0, I: 1, TTL: 1},
	}
	suite.Equal(suite.ws.next_id, 0)
	worker := suite.ws.AddTask(tasks[0])
	scheduled, err := time.Parse(time.RFC3339Nano, worker.ScheduleTime)
	if err != nil {
		suite.Error(err)
	}
	suite.LessOrEqual(scheduled.UnixNano(), time.Now().UTC().UnixNano())
	suite.Equal(worker.Status, Scheduled)
	suite.Equal(worker.NumInQueue, uint(1))
	suite.Equal(worker.CurrIter, uint(0))
	suite.Equal(len(suite.ws.scheduled), 1)
	go suite.ws.StartWorkers()
	time.Sleep(time.Millisecond)
	suite.Equal(suite.ws.next_id, 1)
	suite.Equal(len(suite.ws.scheduled), 0)
	suite.IsType(suite.ws.workingdone[0], &worker)
	var stat Status
	stat = Working
	suite.Equal(suite.ws.workingdone[0].Status, stat)
	suite.Equal(suite.ws.workingdone[0].NumInQueue, uint(0))
	suite.Equal(suite.ws.workingdone[0].CurrIter, uint(1))
	//for _, task := range tasks {
	//	suite.Equal(task, 0)
	//}
}

func TestWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerTestSuite))
}
