package createcall

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
)

type dndFilterResponse struct {
	Numbers numbers `json:"Numbers"`
}

type numbers struct {
	DND string `json:"DND"`
}

// CheckDNDStatus checks if a phone number is registered for Do Not Disturb(DND)
func CheckDNDStatus(phoneNumber string) (bool, error) {
	url := "https://" + configmanager.ConfStore.ExotelAPIKey + ":" + configmanager.ConfStore.ExotelAPIToken + "@api.exotel.com/v1/Accounts/" + configmanager.ConfStore.ExotelAccountSID + "/Numbers/" + phoneNumber + ".json"

	response, err := http.Get(url)
	if err != nil {
		//false negative in case of DND status couldn't be retrieved
		return false, err
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return false, err
	}
	var body dndFilterResponse
	json.Unmarshal(data, &body)

	if strings.ToLower(body.Numbers.DND) == "yes" {
		return true, errors.New("The number is on DND")
	}
	return false, nil
}
