package status

type Status[T ConditionType] struct {
	Conditions []Condition[T]
}

type Condition[T ConditionType] struct {
	Type T
}

type ConditionType interface {
	NodeCondition | WorkloadCondition
}

type NodeCondition string
type WorkloadCondition string
