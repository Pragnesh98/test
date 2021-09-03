package phonenumber

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"github.com/ttacon/libphonenumber"
)

var (
	indianTollfreeRegexp        = regexp.MustCompile(`(?m)^(\+|0|00|\+91|\+9100)?(18[06]0([0-9]{4,9}))$`)
	internationalTollfreeRegexp = regexp.MustCompile(`(?m)^(18[06]0([0-9]{4,8}))$`)
)

// PhoneNumber contains the metadata of a number
type PhoneNumber struct {
	RawNumber              string
	E164Format             string
	LocalFormat            string
	NationalFormat         string
	WithZeroNationalFormat string
	ISOCountryCode         string
	IsLandline             bool
	IsTollFree             bool
	IsInternational        bool
	IsSipUser              bool
}

// NewPhoneNumber returns the PhoneNumber struct with the given Raw number
func NewPhoneNumber(number string) PhoneNumber {
	return PhoneNumber{
		RawNumber: number,
	}
}

// Parse fills the number's metadata for a give number
func (pn *PhoneNumber) Parse(ctx context.Context) error {
	if pn.RawNumber == "" {
		errors.New("Raw number is empty")
	}
	// Check for SIP User
	if err := pn.populateIsSIPUser(); err == nil {
		return nil
	}

	// Check for Indian TollFree number
	if err := pn.populateIndianTollFree(); err == nil {
		return nil
	}

	// Check for International TollFree number
	if err := pn.populateInternationalTollFree(); err == nil {
		return nil
	}

	// Parse the number with Lib phone number library
	if err := pn.parseWithLibPhonenumber(); err != nil {
		return err
	}
	return nil
}

func (pn *PhoneNumber) populateIndianTollFree() error {
	matches := indianTollfreeRegexp.FindStringSubmatch(pn.RawNumber)
	if len(matches) < 3 {
		return errors.New("Number is not an Indian Toll Free Number")
	}
	pn.E164Format = matches[2]
	pn.ISOCountryCode = "91"
	pn.IsTollFree = true
	return nil
}

func (pn *PhoneNumber) populateInternationalTollFree() error {
	matches := internationalTollfreeRegexp.FindStringSubmatch(pn.RawNumber)
	if len(matches) < 2 {
		return errors.New("Number is not an International Toll Free Number")
	}
	pn.E164Format = matches[1]
	pn.IsTollFree = true
	return nil
}

func (pn *PhoneNumber) populateIsSIPUser() error {
	if !strings.HasPrefix(strings.ToLower(pn.RawNumber), "sip:") {
		return errors.New("Number is not a sip user")
	}
	pn.IsSipUser = true
	pn.E164Format = pn.RawNumber[4:]
	return nil
}

func (pn *PhoneNumber) parseWithLibPhonenumber() error {
	var defaultRegion = ""
	if configmanager.ConfStore.DefaultRegion != "" {
		defaultRegion = configmanager.ConfStore.DefaultRegion
	}

	number, err := libphonenumber.Parse(pn.RawNumber, defaultRegion)
	if err != nil {
		return err
	}
	pn.NationalFormat = strconv.Itoa(int(number.GetNationalNumber()))
	pn.WithZeroNationalFormat = "0" + pn.NationalFormat
	pn.ISOCountryCode = strconv.Itoa(int(number.GetCountryCode()))
	pn.E164Format = libphonenumber.Format(number, libphonenumber.E164)
	if pn.ISOCountryCode == "91" {
		intFormat := libphonenumber.Format(number, libphonenumber.INTERNATIONAL)
		intFormatSplit := strings.Split(intFormat, " ")
		if len(intFormatSplit) >= 4 {
			pn.IsLandline = true
			pn.LocalFormat = intFormatSplit[2] + intFormatSplit[3]
		} else {
			pn.LocalFormat = pn.NationalFormat
		}
	} else {
		pn.IsInternational = true
		pn.NationalFormat = libphonenumber.Format(number, libphonenumber.E164)
		pn.WithZeroNationalFormat = libphonenumber.Format(number, libphonenumber.E164)
		pn.LocalFormat = libphonenumber.Format(number, libphonenumber.E164)
	}
	return nil
}
