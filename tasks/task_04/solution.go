package main

type Stats struct {
	Count         int
	Sum, Min, Max int64
}

func Calc(nums []int64) Stats {
	if len(nums) < 2 {
		return Stats{}
	}
	diff := nums[1] - nums[0]
	s := Stats{Count: len(nums) - 1, Sum: diff, Min: diff, Max: diff}

	for i := 2; i < len(nums); i++ {
		diff = nums[i] - nums[i-1]
		if diff < s.Min {
			s.Min = diff
		}
		if diff > s.Max {
			s.Max = diff
		}
		s.Sum += diff
	}
	return s
}
