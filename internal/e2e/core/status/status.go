package status

type Status[T ConditionType] struct {
	Conditions map[T]bool
}

type ConditionType interface {
	NodeCondition | WorkloadCondition
}

type NodeCondition string
type WorkloadCondition string
