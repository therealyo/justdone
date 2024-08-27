package array

func IsSubArray[T comparable](subarray []T, array []T) bool {
	// if len(subarray) == 0 {
	// 	return true
	// }

	// subIndex := 0

	// for _, element := range array {
	// 	if subIndex < len(subarray) && element == subarray[subIndex] {
	// 		subIndex++
	// 	} else if subIndex > 0 && element != subarray[subIndex-1] {
	// 		return false
	// 	}

	// 	if subIndex == len(subarray) {
	// 		return true
	// 	}
	// }

	// return subIndex == len(subarray)
	if len(subarray) == 0 || len(subarray) > len(array) {
		return false
	}

	for i := 0; i < len(subarray); i++ {
		if array[i] != subarray[i] {
			return false
		}
	}

	return true
}
