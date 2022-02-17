package gosession

import (
	"fmt"
	"regexp"
	"time"
)

// AGE VALIDATION ------------------------------------------------------------------------------

// AgeAt gets the age of an entity at a certain time.
func ageAt(birthDate time.Time, now time.Time) int {
	// Get the year number change since the player's birth.
	years := now.Year() - birthDate.Year()

	// If the date is before the date of birth, then not that many years have elapsed.
	birthDay := getAdjustedBirthDay(birthDate, now)
	if now.YearDay() < birthDay {
		years--
	}

	return years
}

// Age is shorthand for AgeAt(birthDate, time.Now()), and carries the same usage and limitations.
func age(birthDate time.Time) int {
	return ageAt(birthDate, time.Now())
}

// Gets the adjusted date of birth to work around leap year differences.
func getAdjustedBirthDay(birthDate time.Time, now time.Time) int {
	birthDay := birthDate.YearDay()
	currentDay := now.YearDay()
	if isLeap(birthDate) && !isLeap(now) && birthDay >= 60 {
		return birthDay - 1
	}
	if isLeap(now) && !isLeap(birthDate) && currentDay >= 60 {
		return birthDay + 1
	}
	return birthDay
}

// Works out if a time.Time is in a leap year.
func isLeap(date time.Time) bool {
	year := date.Year()
	if year%400 == 0 {
		return true
	} else if year%100 == 0 {
		return false
	} else if year%4 == 0 {
		return true
	}
	return false
}

// EMAIL VALIDATION ------------------------------------------------------------------------------

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// isEmailValid checks if the email provided passes the required structure
// and length test. It also checks the domain has a valid MX record.
func isEmailValid(e string) error {
	if len(e) < 3 && len(e) > 254 {
		return fmt.Errorf("The email size is invalid")
	}
	if !emailRegex.MatchString(e) {
		return fmt.Errorf("The email is invalid")
	}
	//TODO readd this
	// parts := strings.Split(e, "@")
	// mx, err := net.LookupMX(parts[1])
	// if err != nil || len(mx) == 0 {
	// 	return fmt.Errorf("The email is invalid")
	// }
	return nil
}

func isTimeValid(e time.Time) error {

	layout := "2006-01-02T15:04:05.000Z"
	_, err := time.Parse(layout, e.Format(layout))
	if err != nil {
		return err
	}

	return nil
}
