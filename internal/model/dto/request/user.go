package request

type UserCreateRequest struct {
	Username string
	Password string
	Role     string
	Status   int8
}

type UserUpdateRequest struct {
	ID       int
	Username string
	Password string
	Role     string
	Status   int8
}

type UserDeleteRequest struct {
	ID int
}
