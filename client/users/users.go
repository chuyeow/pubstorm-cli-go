package users

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/franela/goreq"
	"github.com/nitrous-io/rise-cli-go/apperror"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/util"
)

var (
	ErrCodeRequestFailed    = "request_failed"
	ErrCodeUnexpectedError  = "unexpected_error"
	ErrCodeValidationFailed = "validation_failed"
)

func Create(email, password string) *apperror.Error {
	res, err := goreq.Request{
		Method:      "POST",
		Uri:         config.Host + "/users",
		ContentType: "application/x-www-form-urlencoded",

		Body: url.Values{
			"email":    {email},
			"password": {password},
		}.Encode(),
	}.Do()

	if err != nil {
		return apperror.New(ErrCodeRequestFailed, err, "", true)
	}

	if res.StatusCode != 422 && res.StatusCode != 201 {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == 422 {
		if j["error"] == "invalid_params" {
			fmt.Println("There were errors in your input. Please try again")
			return apperror.New(ErrCodeValidationFailed, nil, util.ValidationErrorsToString(j), false)
		} else {
			return apperror.New(ErrCodeUnexpectedError, err, "", true)
		}
	}

	return nil
}

func Confirm(email, confirmationCode string) *apperror.Error {
	res, err := goreq.Request{
		Method:      "POST",
		Uri:         config.Host + "/user/confirm",
		ContentType: "application/x-www-form-urlencoded",

		Body: url.Values{
			"email":             {email},
			"confirmation_code": {confirmationCode},
		}.Encode(),
	}.Do()

	if err != nil {
		return apperror.New(ErrCodeRequestFailed, err, "", true)
	}

	if res.StatusCode != 422 && res.StatusCode != http.StatusOK {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == 422 {
		resText, err := res.Body.ToString()
		if err != nil {
			return apperror.New(ErrCodeUnexpectedError, err, "", true)
		}

		if strings.Contains(resText, "invalid email or confirmation_code") {
			return apperror.New(ErrCodeValidationFailed, nil, "You've entered an incorrect confirmation code. Please try again.", false)
		} else {
			return apperror.New(ErrCodeUnexpectedError, err, resText, true)
		}
	}

	return nil
}
