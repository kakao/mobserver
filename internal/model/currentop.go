package model

type CurrentOpBatchField struct {
	Op                    string                 `bson:"op"`
	MicrosecsRunning      int64                  `bson:"microsecs_running"`
	SecsRunning           int                    `bson:"secs_running"`
	Ns                    string                 `bson:"ns"`
	Command               map[string]interface{} `bson:"command"`
	Msg                   string                 `bson:"msg"`
	PlanSummary           string                 `bson:"planSummary"`
	WaitingForLatch       interface{}            `bson:"waitingForLatch"`
	WaitingForLock        bool                   `bson:"waitingForLock"`
	WaitingForFlowControl bool                   `bson:"waitingForFlowControl"`
	Transaction           interface{}            `bson:"transaction"`
}

type CurrentOp struct {
	Cursor struct {
		FirstBatch []CurrentOpBatchField `bson:"firstBatch"`
	} `bson:"cursor"`
}
