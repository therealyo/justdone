package array

import (
	"testing"
)

func TestIsSubarray(t *testing.T) {
	// Test with integers
	arr1 := []int{1, 2, 3, 4, 5}
	subarr1 := []int{2, 3, 4}

	if !IsSubArray(subarr1, arr1) {
		t.Errorf("Expected subarr1 to be a subarray of arr1")
	}

	// Test where subarray is missing an item
	subarr2 := []int{2, 4}

	if IsSubArray(subarr2, arr1) {
		t.Errorf("Expected subarr2 not to be a subarray of arr1")
	}

	// Test with a non-contiguous subarray
	subarr3 := []int{1, 3, 4}

	if IsSubArray(subarr3, arr1) {
		t.Errorf("Expected subarr3 not to be a subarray of arr1")
	}

	// Test with a single item subarray
	subarr4 := []int{3}

	if !IsSubArray(subarr4, arr1) {
		t.Errorf("Expected subarr4 to be a subarray of arr1")
	}

	// Test with the entire array as subarray
	subarr5 := []int{1, 2, 3, 4, 5}

	if !IsSubArray(subarr5, arr1) {
		t.Errorf("Expected subarr5 to be a subarray of arr1")
	}

	// Test with an empty subarray
	subarr6 := []int{}

	if !IsSubArray(subarr6, arr1) {
		t.Errorf("Expected an empty subarr6 to be a valid subarray of arr1")
	}

	// Test with strings
	arr2 := []string{"apple", "banana", "cherry", "date"}
	subarr7 := []string{"banana", "cherry"}

	if !IsSubArray(subarr7, arr2) {
		t.Errorf("Expected subarr7 to be a subarray of arr2")
	}

	arr3 := []string{"cool_order_created"}
	subarr8 := []string{"cool_order_created", "sbu_verification_pending", "confirmed_by_mayor", "chinazes"}

	if IsSubArray(subarr8, arr3) {
		t.Errorf("Expected subarr8 not to be a subarray of arr2")
	}
}
