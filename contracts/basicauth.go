package contracts

import (
	"errors"

	"bitbucket.org/yellowmessenger/asterisk-ari/configmanager"
	"github.com/labstack/echo"
)

type BasicAuthCreds struct {
	UserName string `required:"true"`
	Password string `required:"true"`
}

func (ba *BasicAuthCreds) ExtractFromHTTP(c echo.Context) error {
	request := c.Request()
	var ok bool
	ba.UserName, ba.Password, ok = request.BasicAuth()
	if !ok {
		return errors.New("User name or password is not available")
	}
	return nil
}

func (ba *BasicAuthCreds) Authenticate() error {
	if ba.UserName == configmanager.ConfStore.APIUsername && ba.Password == configmanager.ConfStore.APIPassword {
		return nil
	}
	return errors.New("User name or password is invalid")
}
