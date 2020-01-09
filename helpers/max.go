package osscluster2rl

func SliceMax(s []int) int {
	m := 0
	for i, e := range s {
		if i == 0 || e > m {
			m = e
		}
	}
	return (m)
}
