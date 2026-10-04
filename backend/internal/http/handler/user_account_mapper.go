package handler

import (
	"zoj/internal/dto"
	"zoj/internal/service"
)

func toCreateUserParams(req dto.CreateUserReq) service.CreateUserParams {
	return service.CreateUserParams{
		Username:  req.Username,
		StudentNo: req.StudentNo,
		RealName:  req.RealName,
		Password:  req.Password,
		Email:     req.Email,
		Code:      req.Code,
		Avatar:    req.Avatar,
	}
}

func toUpdateProfileParams(req dto.UpdateProfileReq) service.UpdateProfileParams {
	return service.UpdateProfileParams{
		Signature: req.Signature,
		School:    req.School,
		Gender:    req.Gender,
		Avatar:    req.Avatar,
	}
}

func toAuthenticatedUser(u service.AuthenticatedUser) dto.UserInfo {
	return dto.UserInfo{ID: u.ID, Username: u.Username, Role: u.Role}
}

func toLoginResp(r *service.LoginResult) *dto.LoginResp {
	return &dto.LoginResp{
		User:   toAuthenticatedUser(r.User),
		Avatar: r.Avatar,
	}
}

func toCaptchaChallengeResp(r *service.CaptchaChallenge) *dto.CaptchaResp {
	return &dto.CaptchaResp{ID: r.ID, Image: r.Image}
}
