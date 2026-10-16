package service

// CodeRunParams describes one ephemeral run. It never creates a submission.
type CodeRunParams struct {
	ProblemID string
	Code      string
	Input     string
	Language  int
}

// CodeRunResult is the bounded output returned by go-judge for one custom input.
type CodeRunResult struct {
	Status        int
	TimeUsed      int64
	MemoryUsed    int64
	Stdout        string
	Stderr        string
	CompileOutput string
}
