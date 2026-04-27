package http

import "pdd-service/internal/features/auth/application"

func toRegisterInput(req RegisterRequest) application.RegisterInput {
	return application.RegisterInput{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	}
}

func toRegisterResponse(output application.RegisterOutput) RegisterResponse {
	return RegisterResponse{
		User:   toUserResponse(output.User),
		Tokens: toTokensResponse(output.Tokens),
	}
}

func toLoginInput(req LoginRequest) application.LoginInput {
	return application.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	}
}

func toLoginResponse(output application.LoginOutput) LoginResponse {
	return LoginResponse{
		User:   toUserResponse(output.User),
		Tokens: toTokensResponse(output.Tokens),
	}
}

func toRefreshInput(req RefreshRequest) application.RefreshInput {
	return application.RefreshInput{RefreshToken: req.RefreshToken}
}

func toRefreshResponse(output application.RefreshOutput) RefreshResponse {
	return RefreshResponse{
		User:   toUserResponse(output.User),
		Tokens: toTokensResponse(output.Tokens),
	}
}

func toLogoutInput(req LogoutRequest) application.LogoutInput {
	return application.LogoutInput{RefreshToken: req.RefreshToken}
}

func toMeResponse(output application.MeOutput) MeResponse {
	return MeResponse{User: toUserResponse(output.User)}
}

func toUserResponse(output application.UserOutput) UserResponse {
	return UserResponse{
		ID:       output.ID.String(),
		Email:    output.Email,
		FullName: output.FullName,
		Role:     output.Role.String(),
	}
}

func toTokensResponse(output application.TokensOutput) TokensResponse {
	return TokensResponse{
		AccessToken:      output.AccessToken,
		RefreshToken:     output.RefreshToken,
		AccessExpiresAt:  output.AccessExpiresAt,
		RefreshExpiresAt: output.RefreshExpiresAt,
	}
}
