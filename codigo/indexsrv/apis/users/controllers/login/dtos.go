package login

type userRegistrationDTO struct {
	NameField     string `json:"name"`
	EmailField    string `json:"email"`
	PasswordField string `json:"password"`
}

type userLoginDTO struct {
	EmailField    string `json:"email"`
	PasswordField string `json:"password"`
	OTP           string `json:"OTP"`
}

type tokenDTO struct {
	Token string `json:"token"`
}
