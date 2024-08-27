package array

import (
	"testing"
)

func TestIsSubarray(t *testing.T) {
	arr1 := []int{1, 2, 3, 4, 5}
	subarr1 := []int{2, 3, 4}

	if IsSubArray(subarr1, arr1) {
		t.Errorf("Expected subarr1 not to be a subarray of arr1")
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

	if IsSubArray(subarr4, arr1) {
		t.Errorf("Expected subarr4 not to be a subarray of arr1")
	}

	// Test with the entire array as subarray
	subarr5 := []int{1, 2, 3, 4, 5}

	if !IsSubArray(subarr5, arr1) {
		t.Errorf("Expected subarr5 to be a subarray of arr1")
	}

	subarr8 := []string{"cool_order_created"}
	arr3 := []string{"cool_order_created", "sbu_verification_pending", "confirmed_by_mayor", "chinazes"}

	if !IsSubArray(subarr8, arr3) {
		t.Errorf("Expected subarr8 to be a subarray of arr3")
	}

	subarr9 := []string{"cool_order_created", "sbu_verification_pending", "confirmed_by_mayor", "chinazes"}

	if !IsSubArray(subarr9, arr3) {
		t.Errorf("Expected subarr9 to be a subarray of arr3")
	}

	subarr10 := []string{"sbu_verification_pending", "confirmed_by_mayor", "chinazes"}

	if IsSubArray(subarr10, arr3) {
		t.Errorf("Expected subarr10 not to be a subarray of arr3")
	}
}
