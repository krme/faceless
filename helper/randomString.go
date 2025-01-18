package helper

import (
	"crypto/rand"
	math "math/rand"
)

type RandomStringType int

const (
	LettersAndNumbers     RandomStringType = 0
	OnlyNumbers           RandomStringType = 1
	OnlyLowerCharacters   RandomStringType = 2
	OnlyUpperCharacters   RandomStringType = 3
	OnlySpecialCharacters RandomStringType = 4
	Password              RandomStringType = 5
)

func CreateRandomPassword() (string, error) {
	lowerCharacters, err := CreateRandomString(3, OnlyLowerCharacters)
	if err != nil {
		return "", err
	}

	upperCharacters, err := CreateRandomString(3, OnlyUpperCharacters)
	if err != nil {
		return "", err
	}

	specialCharacter, err := CreateRandomString(1, OnlySpecialCharacters)
	if err != nil {
		return "", err
	}

	number, err := CreateRandomString(1, OnlyNumbers)
	if err != nil {
		return "", err
	}

	in := lowerCharacters + upperCharacters + specialCharacter + number

	inRune := []rune(in)
	math.Shuffle(len(inRune), func(i, j int) {
		inRune[i], inRune[j] = inRune[j], inRune[i]
	})

	return string(inRune), nil
}

func CreateRandomString(length int, stringType RandomStringType) (string, error) {
	var letterBytes = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	if stringType == OnlyNumbers {
		letterBytes = "0123456789"
	} else if stringType == OnlyUpperCharacters {
		letterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	} else if stringType == OnlyLowerCharacters {
		letterBytes = "abcdefghijklmnopqrstuvwxyz"
	} else if stringType == OnlySpecialCharacters {
		letterBytes = "!@#$%^&*()_+={};':\"|\\,.<>/?~-"
	}

	var b = make([]byte, length)

	_, err := rand.Read([]byte(b))
	if err != nil {
		return "", err
	}

	for i := 0; i < length; i++ {
		b[i] = letterBytes[int(b[i])%len(letterBytes)]
	}
	return string(b[:]), nil
}
