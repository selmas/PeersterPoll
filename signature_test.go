package main

import "testing"

func mapToPointDeterministic(t  *testing.T){
	input := []byte("Test input")
	X1, Y1 := mapToPoint(input)
	X2, Y2 := mapToPoint(input)

	if X1 != X2 || Y1 != Y2 {
		t.Errorf("Mapped same input to different points, got:\n X = %d \n Y = %d \n expected: X = %d \n Y = %d", X1,
			Y1, X2,Y2)
	}
}
