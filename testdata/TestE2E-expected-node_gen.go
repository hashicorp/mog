// Code generated by mog. DO NOT EDIT.

package sourcepkg

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/hashicorp/mog/internal/e2e/core"
)

func NodeToCore(s *Node, t *core.ClusterNode) {
	if s == nil {
		return
	}
	t.ID = s.ID
	t.Label = core.Label(s.Label)
	t.O = s.O
	t.I = s.I
	WorkloadToCore(&s.F1, &t.F1)
	if s.F2 != nil {
		var x core.Workload
		WorkloadToCore(s.F2, &x)
		t.F2 = &x
	}
	if s.F3 != nil {
		WorkloadToCore(s.F3, &t.F3)
	}
	{
		var x core.Workload
		WorkloadToCore(&s.F4, &x)
		t.F4 = &x
	}
	t.S1 = s.S1
	t.S2 = s.S2
	{
		t.S3 = make([]string, len(s.S3))
		for i := range s.S3 {
			t.S3[i] = *s.S3[i]
		}
	}
	{
		t.S4 = make([]*string, len(s.S4))
		for i := range s.S4 {
			t.S4[i] = &s.S4[i]
		}
	}
	{
		t.S5 = make([]core.Workload, len(s.S5))
		for i := range s.S5 {
			WorkloadToCore(&s.S5[i], &t.S5[i])
		}
	}
	{
		t.S6 = make([]*core.Workload, len(s.S6))
		for i := range s.S6 {
			if s.S6[i] != nil {
				var x core.Workload
				WorkloadToCore(s.S6[i], &x)
				t.S6[i] = &x
			}
		}
	}
	{
		t.S7 = make([]core.Workload, len(s.S7))
		for i := range s.S7 {
			if s.S7[i] != nil {
				WorkloadToCore(s.S7[i], &t.S7[i])
			}
		}
	}
	{
		t.S8 = make([]*core.Workload, len(s.S8))
		for i := range s.S8 {
			{
				var x core.Workload
				WorkloadToCore(&s.S8[i], &x)
				t.S8[i] = &x
			}
		}
	}
	t.S9 = s.S9
	t.S10 = s.S10
	{
		t.S11 = make(core.WorkloadSlice, len(s.S11))
		for i := range s.S11 {
			{
				var x core.Workload
				WorkloadToCore(&s.S11[i], &x)
				t.S11[i] = &x
			}
		}
	}
	{
		t.S12 = make([]*core.Workload, len(s.S12))
		for i := range s.S12 {
			{
				var x core.Workload
				WorkloadToCore(&s.S12[i], &x)
				t.S12[i] = &x
			}
		}
	}
	{
		t.S13 = make(core.WorkloadSlice, len(s.S13))
		for i := range s.S13 {
			{
				var x core.Workload
				WorkloadToCore(&s.S13[i], &x)
				t.S13[i] = &x
			}
		}
	}
	t.M1 = s.M1
	t.M2 = s.M2
	{
		t.M3 = make(map[string]string, len(s.M3))
		for k, v := range s.M3 {
			var y string
			y = *v
			t.M3[k] = y
		}
	}
	{
		t.M4 = make(map[string]*string, len(s.M4))
		for k, v := range s.M4 {
			var y *string
			y = &v
			t.M4[k] = y
		}
	}
	{
		t.M5 = make(map[string]core.Workload, len(s.M5))
		for k, v := range s.M5 {
			var y core.Workload
			WorkloadToCore(&v, &y)
			t.M5[k] = y
		}
	}
	{
		t.M6 = make(map[string]*core.Workload, len(s.M6))
		for k, v := range s.M6 {
			var y *core.Workload
			if v != nil {
				var x core.Workload
				WorkloadToCore(v, &x)
				y = &x
			}
			t.M6[k] = y
		}
	}
	{
		t.M7 = make(map[string]core.Workload, len(s.M7))
		for k, v := range s.M7 {
			var y core.Workload
			if v != nil {
				WorkloadToCore(v, &y)
			}
			t.M7[k] = y
		}
	}
	{
		t.M8 = make(map[string]*core.Workload, len(s.M8))
		for k, v := range s.M8 {
			var y *core.Workload
			{
				var x core.Workload
				WorkloadToCore(&v, &x)
				y = &x
			}
			t.M8[k] = y
		}
	}
	t.T3, _ = ptypes.Timestamp(s.T3)
	t.D3, _ = ptypes.Duration(s.D3)
}
func NodeFromCore(t *core.ClusterNode, s *Node) {
	if s == nil {
		return
	}
	s.ID = t.ID
	s.Label = string(t.Label)
	s.O = t.O
	s.I = t.I
	WorkloadFromCore(&t.F1, &s.F1)
	if t.F2 != nil {
		var x Workload
		WorkloadFromCore(t.F2, &x)
		s.F2 = &x
	}
	{
		var x Workload
		WorkloadFromCore(&t.F3, &x)
		s.F3 = &x
	}
	if t.F4 != nil {
		WorkloadFromCore(t.F4, &s.F4)
	}
	s.S1 = t.S1
	s.S2 = t.S2
	{
		s.S3 = make([]*string, len(t.S3))
		for i := range t.S3 {
			s.S3[i] = &t.S3[i]
		}
	}
	{
		s.S4 = make([]string, len(t.S4))
		for i := range t.S4 {
			s.S4[i] = *t.S4[i]
		}
	}
	{
		s.S5 = make([]Workload, len(t.S5))
		for i := range t.S5 {
			WorkloadFromCore(&t.S5[i], &s.S5[i])
		}
	}
	{
		s.S6 = make([]*Workload, len(t.S6))
		for i := range t.S6 {
			if t.S6[i] != nil {
				var x Workload
				WorkloadFromCore(t.S6[i], &x)
				s.S6[i] = &x
			}
		}
	}
	{
		s.S7 = make([]*Workload, len(t.S7))
		for i := range t.S7 {
			{
				var x Workload
				WorkloadFromCore(&t.S7[i], &x)
				s.S7[i] = &x
			}
		}
	}
	{
		s.S8 = make([]Workload, len(t.S8))
		for i := range t.S8 {
			if t.S8[i] != nil {
				WorkloadFromCore(t.S8[i], &s.S8[i])
			}
		}
	}
	s.S9 = t.S9
	s.S10 = t.S10
	{
		s.S11 = make([]Workload, len(t.S11))
		for i := range t.S11 {
			if t.S11[i] != nil {
				WorkloadFromCore(t.S11[i], &s.S11[i])
			}
		}
	}
	{
		s.S12 = make(WorkloadSlice, len(t.S12))
		for i := range t.S12 {
			if t.S12[i] != nil {
				WorkloadFromCore(t.S12[i], &s.S12[i])
			}
		}
	}
	{
		s.S13 = make(WorkloadSlice, len(t.S13))
		for i := range t.S13 {
			if t.S13[i] != nil {
				WorkloadFromCore(t.S13[i], &s.S13[i])
			}
		}
	}
	s.M1 = t.M1
	s.M2 = t.M2
	{
		s.M3 = make(map[string]*string, len(t.M3))
		for k, v := range t.M3 {
			var y *string
			y = &v
			s.M3[k] = y
		}
	}
	{
		s.M4 = make(map[string]string, len(t.M4))
		for k, v := range t.M4 {
			var y string
			y = *v
			s.M4[k] = y
		}
	}
	{
		s.M5 = make(map[string]Workload, len(t.M5))
		for k, v := range t.M5 {
			var y Workload
			WorkloadFromCore(&v, &y)
			s.M5[k] = y
		}
	}
	{
		s.M6 = make(map[string]*Workload, len(t.M6))
		for k, v := range t.M6 {
			var y *Workload
			if v != nil {
				var x Workload
				WorkloadFromCore(v, &x)
				y = &x
			}
			s.M6[k] = y
		}
	}
	{
		s.M7 = make(map[string]*Workload, len(t.M7))
		for k, v := range t.M7 {
			var y *Workload
			{
				var x Workload
				WorkloadFromCore(&v, &x)
				y = &x
			}
			s.M7[k] = y
		}
	}
	{
		s.M8 = make(map[string]Workload, len(t.M8))
		for k, v := range t.M8 {
			var y Workload
			if v != nil {
				WorkloadFromCore(v, &y)
			}
			s.M8[k] = y
		}
	}
	s.T3, _ = ptypes.TimestampProto(t.T3)
	s.D3 = ptypes.DurationProto(t.D3)
}
func WorkloadToCore(s *Workload, t *core.Workload) {
	if s == nil {
		return
	}
	t.ID = s.ID
	t.Value = int(s.Value)
}
func WorkloadFromCore(t *core.Workload, s *Workload) {
	if s == nil {
		return
	}
	s.ID = t.ID
	s.Value = int32(t.Value)
}
