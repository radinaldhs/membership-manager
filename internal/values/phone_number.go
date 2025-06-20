package values

import (
	"errors"
	"regexp"
	"strings"
)

var (
	phoneNumberRegex          = regexp.MustCompile(`^(0|(\+)?62(0)?)?(?P<number>8\d+)$`)
	phoneRealNumberGroupIndex = phoneNumberRegex.SubexpIndex("number")
)

// PhoneNumber represent an Indonesian phone number
type PhoneNumber string

func parsePhoneNumber(str string) (PhoneNumber, error) {
	submatch := phoneNumberRegex.FindStringSubmatch(str)
	if submatch == nil {
		return "", errors.New("phone number can not be parsed")
	}

	number := submatch[phoneRealNumberGroupIndex]

	return PhoneNumber(number), nil
}

// ParseDirtyPhoneNumber should be used when you want to parse phone
// number from a source that doesn't have good format standardization,
// in our case this source is from the database
func ParseDirtyPhoneNumber(str string) (PhoneNumber, error) {
	// I've hit the limitation of regex, so we need to
	// clean the data before we can extract the phone number
	str = strings.ReplaceAll(str, "-", "")
	str = strings.ReplaceAll(str, " ", "")

	return parsePhoneNumber(str)
}

func ParsePhoneNumber(str string) (PhoneNumber, error) {
	return parsePhoneNumber(str)
}

func (phoneNum PhoneNumber) Standard() string {
	return "0" + string(phoneNum)
}

func (phoneNum PhoneNumber) WithIDCountryCode() string {
	return "62" + string(phoneNum)
}

func (phoneNum PhoneNumber) IsEqual(p PhoneNumber) bool {
	return phoneNum == p
}

func IsPhoneNumberValid(phoneNum string) bool {
	return phoneNumberRegex.MatchString(phoneNum)
}
