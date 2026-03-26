package llm

import "fmt"

// TaskType identifies the kind of LLM task being routed.
type TaskType string

const (
	TaskDecompose        TaskType = "decompose"
	TaskExtractJob       TaskType = "extract_job"
	TaskGapAnalysis      TaskType = "gap_analysis"
	TaskSynthesizeResume TaskType = "synthesize_resume"
	TaskCoverLetter      TaskType = "cover_letter"
	TaskCoach            TaskType = "coach"
)

// Router maps task types to LLM providers.
type Router struct {
	Providers map[string]Provider
	TaskMap   map[TaskType]string // task type -> provider name
}

// NewRouter creates a Router from provider instances and a task-to-provider mapping.
func NewRouter(providers map[string]Provider, taskMap map[string]string) *Router {
	tm := make(map[TaskType]string, len(taskMap))
	for k, v := range taskMap {
		tm[TaskType(k)] = v
	}
	return &Router{
		Providers: providers,
		TaskMap:   tm,
	}
}

// ForTask returns the provider configured for the given task type.
func (r *Router) ForTask(task TaskType) (Provider, error) {
	name, ok := r.TaskMap[task]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNoProvider, task)
	}

	p, ok := r.Providers[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}

	return p, nil
}
