package main

import "testing"

func TestFilterByNumber(t *testing.T) {
	table := []struct {
		name string
		num  int
		mp   map[int]mapping
		exp  []int
	}{
		{
			name: "one level",
			num:  1,
			mp: map[int]mapping{
				1: {0, []int{2, 3}},
			},
			exp: []int{0, 1, 2, 3},
		},
		{
			name: "two level",
			num:  1,
			mp: map[int]mapping{
				1: {0, []int{2, 3}},
				2: {1, []int{4}},
				4: {2, []int{5}},
				5: {4, []int{}},
			},
			exp: []int{0, 1, 2, 4, 5, 3},
		},
		{
			name: "two level, multiple item",
			num:  1,
			mp: map[int]mapping{
				1: {0, []int{2, 3}},
				2: {1, []int{4, 6}},
				4: {2, []int{5, 7}},
				5: {4, []int{}},
				6: {2, []int{}},
				7: {4, []int{}},
			},
			exp: []int{0, 1, 2, 4, 5, 7, 6, 3},
		},
		{
			name: "multiple level",
			num:  4,
			mp: map[int]mapping{
				1: {0, []int{2, 3}},
				3: {1, []int{4}},
				4: {3, []int{5}},
				5: {4, []int{6}},
				6: {5, []int{7}},
			},
			exp: []int{0, 1, 3, 4, 5, 6, 7},
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			d := data{mappings: tt.mp}
			out := filterByNumber(d, tt.num)

			if len(tt.exp) != len(out) {
				t.Errorf("incorrect output. expected %v, got %v", tt.exp, out)
				return
			}

			for i := range tt.exp {
				if tt.exp[i] != out[i] {
					t.Errorf("incorrect output. expected %v, got %v", tt.exp, out)
					return
				}
			}
		})
	}
}
