package auth

// Service aggregates the Auth use cases behind a single struct so the HTTP
// handler layer depends on one thing instead of three. Each field is still
// independently testable and independently injectable — this is purely a
// wiring convenience, not a god-object (each use case keeps its own single
// responsibility).
type Service struct {
	Register *RegisterUseCase
	Login    *LoginUseCase
	Logout   *LogoutUseCase
	Refresh  *RefreshUseCase
}

// NewService constructs the Auth service facade.
func NewService(register *RegisterUseCase, login *LoginUseCase, logout *LogoutUseCase, refresh *RefreshUseCase) *Service {
	return &Service{Register: register, Login: login, Logout: logout, Refresh: refresh}
}
