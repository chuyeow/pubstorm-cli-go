package users

import (
	"net/http"
	"net/url"

	"github.com/franela/goreq"
	"github.com/nitrous-io/rise-cli-go/apperror"
	"github.com/nitrous-io/rise-cli-go/config"
	"github.com/nitrous-io/rise-cli-go/util"
)

const (
	ErrCodeRequestFailed    = "request_failed"
	ErrCodeUnexpectedError  = "unexpected_error"
	ErrCodeValidationFailed = "validation_failed"
)

func Create(email, password string) *apperror.Error {
	res, err := goreq.Request{
		Method:      "POST",
		Uri:         config.Host + "/users",
		ContentType: "application/x-www-form-urlencoded",
		Accept:      config.ReqAccept,
		UserAgent:   config.UserAgent,

		Body: url.Values{
			"email":    {email},
			"password": {password},
		}.Encode(),
	}.Do()

	if err != nil {
		return apperror.New(ErrCodeRequestFailed, err, "", true)
	}

	if !util.ContainsInt([]int{http.StatusCreated, 422}, res.StatusCode) {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == 422 {
		if j["error"] == "invalid_params" {
			return apperror.New(ErrCodeValidationFailed, nil, util.ValidationErrorsToString(j), false)
		}
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return nil
}

func Confirm(email, confirmationCode string) *apperror.Error {
	res, err := goreq.Request{
		Method:      "POST",
		Uri:         config.Host + "/user/confirm",
		ContentType: "application/x-www-form-urlencoded",
		Accept:      config.ReqAccept,
		UserAgent:   config.UserAgent,

		Body: url.Values{
			"email":             {email},
			"confirmation_code": {confirmationCode},
		}.Encode(),
	}.Do()

	if err != nil {
		return apperror.New(ErrCodeRequestFailed, err, "", true)
	}

	if !util.ContainsInt([]int{http.StatusOK, 422}, res.StatusCode) {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == 422 {
		if j["error"] == "invalid_params" && j["error_description"] == "invalid email or confirmation_code" {
			return apperror.New(ErrCodeValidationFailed, nil, "You've entered an incorrect confirmation code. Please try again.", false)
		}
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if v, ok := j["confirmed"].(bool); !v || !ok {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return nil
}

func ResendConfirmationCode(email string) *apperror.Error {
	res, err := goreq.Request{
		Method:      "POST",
		Uri:         config.Host + "/user/confirm/resend",
		ContentType: "application/x-www-form-urlencoded",
		Accept:      config.ReqAccept,
		UserAgent:   config.UserAgent,

		Body: url.Values{
			"email": {email},
		}.Encode(),
	}.Do()

	if err != nil {
		return apperror.New(ErrCodeRequestFailed, err, "", true)
	}

	if !util.ContainsInt([]int{http.StatusOK, 422}, res.StatusCode) {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	var j map[string]interface{}
	if err := res.Body.FromJsonTo(&j); err != nil {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if res.StatusCode == 422 {
		if j["error"] == "invalid_params" && j["error_description"] == "email is not found or already confirmed" {
			return apperror.New(ErrCodeValidationFailed, nil, "Could not request confirmation code to be resent. (Is it already confirmed?)", true)
		}
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	if v, ok := j["sent"].(bool); !v || !ok {
		return apperror.New(ErrCodeUnexpectedError, err, "", true)
	}

	return nil
}
