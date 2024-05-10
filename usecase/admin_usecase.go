package usecase

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/yafireyhan01/e-wallet/model"
	"github.com/yafireyhan01/e-wallet/model/dto"
	"github.com/yafireyhan01/e-wallet/repository"
	"github.com/yafireyhan01/e-wallet/utils/common"
	encryption "github.com/yafireyhan01/e-wallet/utils/encription"
)

type AdminUseCase interface {
	RegisterAdmin(payload model.Admin) (model.Admin, error)
	LoginAdmin(payload dto.LoginRequestDto) (dto.LoginResponseDto, error)
	GetUserInfo(userID string) (model.User, error)
}

type adminUseCase struct {
	repo           repository.AdminRepository
	userRepository repository.UserRepository
}

func (a *adminUseCase) RegisterAdmin(payload model.Admin) (model.Admin, error) {
	response, err := a.repo.Register(payload)
	if err != nil {
		return model.Admin{}, err
	}
	return response, nil
}

func (a *adminUseCase) LoginAdmin(payload dto.LoginRequestDto) (dto.LoginResponseDto, error) {
	var claims dto.LoginResponseDto
	response, err := a.repo.Get(payload)
	if err != nil {
		return dto.LoginResponseDto{}, err
	}
	isValid := encryption.CheckPasswordHash(payload.Pass, response.Password)
	if !isValid {
		return dto.LoginResponseDto{}, errors.New("password salah")
	}

	loginExpDuration, _ := strconv.Atoi(os.Getenv("TOKEN_LIFE_TIME"))
	expiredAt := time.Now().Add(time.Duration(loginExpDuration) * time.Minute).Unix()
	// TODO: tempel generate token jwt
	accessToken, err := common.GenerateTokenJwt(response.Id, response.Name, response.Role, expiredAt)
	if err != nil {
		return dto.LoginResponseDto{}, err
	}

	claims.AccessToken = accessToken
	claims.UserId = response.Id
	return claims, nil

}
func (a *adminUseCase) GetUserInfo(userID string) (model.User, error) {

	info := "u.id= '" + userID + "'"
	limit := 1
	offset := 0

	users, err := a.userRepository.GetInfoUser(info, limit, offset)
	if err != nil {
		return model.User{}, err
	}
	if len(users) == 0 {
		return model.User{}, errors.New("user tidak ada")
	}
	return users[0], nil

}

func NewAdminUseCase(repo repository.AdminRepository) AdminUseCase {
	return &adminUseCase{repo: repo}
}
