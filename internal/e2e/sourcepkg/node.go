package sourcepkg

import (
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/hashicorp/mog/internal/e2e/core"
	"github.com/hashicorp/mog/internal/e2e/core/inner"
)

// Node source structure for e2e testing mog.
//
// mog annotation:
//
// name=Core
// target=github.com/hashicorp/mog/internal/e2e/core.ClusterNode
// output=node_gen.go
type Node struct {
	ID     string
	Weight int64
	Label  string
	// Labels []string
	Meta map[string]interface{}
	Work []Workload
	// WorkPointer []*Workload

	O *core.Other
	I inner.Inner

	F1 Workload  // for testing struct-to-struct
	F2 *Workload // for testing ptr-to-ptr
	F3 *Workload // for testing ptr-to-struct
	F4 Workload  // for testing struct-to-ptr

	S1 []string  // for testing struct-to-struct for basic slices
	S2 []*string // for testing ptr-to-ptr for basic slices
	S3 []*string // for testing ptr-to-struct for basic slices
	S4 []string  // for testing struct-to-ptr for basic slices

	S5 []Workload  // for testing struct-to-struct for struct slices
	S6 []*Workload // for testing ptr-to-ptr for struct slices
	S7 []*Workload // for testing ptr-to-struct for struct slices
	S8 []Workload  // for testing struct-to-ptr for struct slices

	S9  []string
	S10 StringSlice

	S11 []Workload
	S12 WorkloadSlice
	S13 WorkloadSlice

	M1 map[string]string  // for testing struct-to-string for basic map values
	M2 map[string]*string // for testing ptr-to-ptr for basic map values
	M3 map[string]*string // for testing ptr-to-string for basic map values
	M4 map[string]string  // for testing struct-to-ptr for basic map values

	M5 map[string]Workload
	M6 map[string]*Workload
	M7 map[string]*Workload
	M8 map[string]Workload

	// T1 timestamp.Timestamp  // for testing struct-to-struct for time
	// T2 *timestamp.Timestamp // for testing ptr-to-ptr for time
	T3 *timestamp.Timestamp // for testing ptr-to-struct for time
	// T4 timestamp.Timestamp  // for testing struct-to-ptr for time

	// D1 duration.Duration  // for testing struct-to-struct for duration
	// D2 *duration.Duration // for testing ptr-to-ptr for duration
	D3 *duration.Duration // for testing ptr-to-struct for duration
	// D4 duration.Duration  // for testing struct-to-ptr for duration
}

type StringSlice []string
type WorkloadSlice []Workload

// mog annotation:
//
// name=Core
// target=github.com/hashicorp/mog/internal/e2e/core.Workload
// output=node_gen.go
type Workload struct {
	ID string
	// mog: func-to=int func-from=int32
	Value int32
}
