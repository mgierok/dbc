package fixture

type Worker interface {
	Work() string
}

type OldName struct{}

func (OldName) Work() string {
	return helper()
}

func helper() string {
	return "done"
}

func useWorker(worker Worker) string {
	return worker.Work()
}

func UseOldName(value OldName) string {
	return useWorker(value)
}
