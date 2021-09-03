package sipregister

import (
	"context"

	"bitbucket.org/yellowmessenger/asterisk-ari/contracts"
	"bitbucket.org/yellowmessenger/asterisk-ari/models/mysql"
)

func Create(
	ctx context.Context,
	req contracts.CreateSIPRegisterRequest,
) (
	*contracts.CreateSIPRegisterResponse,
	error,
) {
	// Insert the user into database
	err := mysql.InsertSIPBuddies(*req.UserID, *req.MD5Secret)
	if err != nil {
		return &contracts.CreateSIPRegisterResponse{}, err
	}
	response := new(contracts.CreateSIPRegisterResponse)
	responseData := new(contracts.SingleCreateSIPRegisterResponse)
	responseData.Msg = "User successfully created"
	responseData.Status = "success"
	response.ResponseData = *responseData
	return response, nil
}

func UserAlreadyExists(
	ctx context.Context,
	req contracts.CreateSIPRegisterRequest,
) (bool, error) {
	exists, err := mysql.SIPBuddyExists(*req.UserID)
	if err != nil {
		return false, err
	}
	return exists, nil
}
