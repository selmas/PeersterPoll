package main

import (
	"testing"
	"crypto/elliptic"
)

func TestMapToPointDeterministic(t  *testing.T){
	curve = elliptic.P256()

	input := []byte("Test input")

	X1, Y1 := mapToPoint(input)
	X2, Y2 := mapToPoint(input)

	if X1.Cmp(X2) != 0 || Y1.Cmp(Y2) != 0 {
		t.Errorf("Mapped same input to different points, " +
			"\n X1 = %d \n X2 = %d \n Y1 = %d \n Y2 = %d", X1,
			X2, Y1, Y2)
	}
}

func TestDifferentInputMapsToDifferentPoints(t  *testing.T){
	curve = elliptic.P256()

	input1 := []byte("Test input 1")
	input2 := []byte("Test input 2")

	X1, Y1 := mapToPoint(input1)
	X2, Y2 := mapToPoint(input2)

	if X1.Cmp(X2) == 0 && Y1.Cmp(Y2) == 0 {
		t.Errorf("Mapped different input to same point, got:\n X = %d \n tag = %d ", X1,Y1)
	}
}

func TestMapToPointReturnsPointOnCurve(t  *testing.T)  {
	curve = elliptic.P256()

	input := []byte("Test input")
	X1, Y1 := mapToPoint(input)

	if !curve.IsOnCurve(X1,Y1) {
		t.Errorf("Point not on curve, got:\n X = %d \n tag = %d", X1, Y1)
	}
}