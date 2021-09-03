package contracts

import (
	"encoding/json"
	"errors"

	"github.com/labstack/echo"
)

type CreateSIPRegisterRequest struct {
	UserID    *string `json:"user_id"`
	MD5Secret *string `json:"md5_secret"`
}

func (csr *CreateSIPRegisterRequest) ExtractFromHTTP(c echo.Context) error {
	request := c.Request()
	err := json.NewDecoder(request.Body).Decode(csr)
	if err != nil {
		return err
	}
	return nil
}

func (csr *CreateSIPRegisterRequest) Validate() error {
	if csr.UserID == nil || len(*csr.UserID) <= 0 {
		return errors.New("user_id parameter is missing or empty")
	}
	if csr.MD5Secret == nil || len(*csr.MD5Secret) <= 0 {
		return errors.New("md5_secret parameter is missing or empty")
	}
	return nil
}
