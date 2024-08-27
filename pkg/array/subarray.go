package array

func IsSubArray[T comparable](subarray []T, array []T) bool {
	if len(subarray) == 0 {
		return true
	}

	subIndex := 0

	for _, element := range array {
		if subIndex < len(subarray) && element == subarray[subIndex] {
			subIndex++
		} else if subIndex > 0 && element != subarray[subIndex-1] {
			// If we find a mismatch after matching some items, return false
			return false
		}

		if subIndex == len(subarray) {
			return true
		}
	}

	// Ensure all items from the subarray were matched in sequence
	return subIndex == len(subarray)
}
