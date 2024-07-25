package controllers

import (
	"regexp"
)

func IsValidPhoneNumber(tel string) bool {
	tel = regexp.MustCompile(`[-]`).ReplaceAllString(tel, "")
	return regexp.MustCompile(`^\d{10}$`).MatchString(tel)
}

func IsValidEmail(email string) bool {
	return regexp.MustCompile(`^([a-zA-Z0-9._%-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})$`).MatchString(email)
}
