package workerstore

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type WorkerTestSuite struct {
	suite.Suite
	ws *WorkerStore
}

func (suite *WorkerTestSuite) SetupTest() {
	suite.ws = NewWorkerStore(4)
}

func (suite *WorkerTestSuite) TestWorker() {
	suite.Equal(suite.ws.next_id, 0)
}

func TestWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(WorkerTestSuite))
}
