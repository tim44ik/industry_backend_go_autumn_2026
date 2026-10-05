package main

import (
	"errors"
	"strconv"
)

var ErrZero = errors.New("zero is not allowed")

func fizzBuzz(n int) (string, error) {
	if n == 0 {
		return "", ErrZero
	}

	isFizz := n%3 == 0
	isBuzz := n%5 == 0

	if isFizz && isBuzz {
		return "FizzBuzz", nil
	}
	if isFizz {
		return "Fizz", nil
	}
	if isBuzz {
		return "Buzz", nil
	}

	return strconv.Itoa(n), nil
}
