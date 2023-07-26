package utils

type ResetStruct interface {
	Reset()
}

func FindIndexes(largerSlice, smallerSlice []byte) []int {
	indexes := make([]int, 0)

	for i := 0; i <= len(largerSlice)-len(smallerSlice); i++ {
		match := true
		for j := 0; j < len(smallerSlice); j++ {
			if largerSlice[i+j] != smallerSlice[j] {
				match = false
				break
			}
		}
		if match {
			indexes = append(indexes, i)
		}
	}

	return indexes
}
