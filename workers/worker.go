package workers

type Worker interface {
	Responses() <-chan Message
	PostMessage(m Message)
	Run()
	Close()
}

type BaseWorker struct {
	requests  chan Message
	responses chan Message
}

func (bw *BaseWorker) Responses() <-chan Message {
	return bw.responses
}

func (bw *BaseWorker) PostMessage(m Message) {
	bw.requests <- m
}

func (bw *BaseWorker) Close() {
	close(bw.requests)
}
