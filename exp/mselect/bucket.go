package mselect

const bucketSize = 512

var (
	blockch   chan interface{}
	blockTask = NewTask(blockch, nil, nil)
)

type taskBucket struct {
	many *manySelect

	cases []runtimeSelect
	tasks []*Task

	block bool
}

func newTaskBucket(many *manySelect, userTask *Task) *taskBucket {
	b := &taskBucket{
		many:  many,
		cases: make([]runtimeSelect, 0, 16),
		tasks: make([]*Task, 0, 16),
	}
	sigTask := NewTask(b.many.tasks, nil, nil)
	b.addTask(sigTask)
	b.addTask(userTask)
	go b.loop()
	return b
}

func (b *taskBucket) loop() {
	for {
		// Wait on select cases.
		i, ok := reflect_rselect(b.cases)
		task := b.tasks[i]
		recv := task.getAndResetRecvValue(&b.cases[i])

		// Got a signal or a new task submitted.
		if i == 0 {
			if !ok { // closed
				b.purgeTasks()
				return
			}

			// Add a new task.
			newTask := task.convFunc(recv).(*Task)
			b.addTask(newTask)

			// If the bucket is full, don't accept new tasks.
			if len(b.cases) == bucketSize {
				b.tasks[0] = blockTask
				b.cases[0] = blockTask.newRuntimeSelect()
				b.block = true
			}
			continue
		}

		// Execute the registered task.
		if task.execFunc != nil {
			task.execFunc(recv, ok)
		}

		// The channel has been closed, delete the task.
		if !ok {
			b.deleteTask(i)
			if b.block && len(b.cases) < bucketSize {
				sigTask := NewTask(b.many.tasks, nil, nil)
				b.tasks[0] = sigTask
				b.cases[0] = sigTask.newRuntimeSelect()
				b.block = false
			}
			b.many.decrCount()
		}
	}
}

func (b *taskBucket) addTask(task *Task) {
	b.tasks = append(b.tasks, task)
	b.cases = append(b.cases, task.newRuntimeSelect())
}

func (b *taskBucket) deleteTask(i int) {
	n := len(b.cases)
	b.cases[i] = b.cases[n-1]
	b.cases[n-1] = runtimeSelect{}
	b.cases = b.cases[:n-1]
	b.tasks[i] = b.tasks[n-1]
	b.tasks[n-1] = nil
	b.tasks = b.tasks[:n-1]
}

func (b *taskBucket) purgeTasks() {
	b.cases = nil
	b.tasks = nil
}
