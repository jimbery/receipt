package similarity

// JaroWinkler returns a similarity score in [0, 1].
func JaroWinkler(s1, s2 string) float64 {
	if s1 == s2 {
		return 1
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0
	}

	jaro := jaro(s1, s2)
	prefix := commonPrefixLen(s1, s2, 4)
	return jaro + float64(prefix)*0.1*(1-jaro)
}

func jaro(s1, s2 string) float64 {
	l1, l2 := len(s1), len(s2)
	matchDist := max(max(l1, l2)/2-1, 0)

	s1Matches := make([]bool, l1)
	s2Matches := make([]bool, l2)
	matches := 0

	for i := range l1 {
		start := max(0, i-matchDist)
		end := min(i+matchDist+1, l2)
		for j := start; j < end; j++ {
			if s2Matches[j] || s1[i] != s2[j] {
				continue
			}
			s1Matches[i] = true
			s2Matches[j] = true
			matches++
			break
		}
	}
	if matches == 0 {
		return 0
	}

	transpositions := 0
	k := 0
	for i := range l1 {
		if !s1Matches[i] {
			continue
		}
		for !s2Matches[k] {
			k++
		}
		if s1[i] != s2[k] {
			transpositions++
		}
		k++
	}

	m := float64(matches)
	t := float64(transpositions) / 2
	return (m/float64(l1) + m/float64(l2) + (m-t)/m) / 3
}

func commonPrefixLen(s1, s2 string, limit int) int {
	n := min(len(s1), len(s2), limit)
	for i := range n {
		if s1[i] != s2[i] {
			return i
		}
	}
	return n
}
