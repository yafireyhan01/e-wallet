package usecase

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/yafireyhan01/e-wallet/model"
	"github.com/yafireyhan01/e-wallet/model/dto"
	"github.com/yafireyhan01/e-wallet/repository"
	"github.com/yafireyhan01/e-wallet/utils/common"
	encryption "github.com/yafireyhan01/e-wallet/utils/encription"
)

type UserUseCase interface {
	CreateUser(payload dto.UserRequestDto) (model.User, error)
	LoginUser(in dto.LoginRequestDto) (dto.LoginResponseDto, error)
	FindById(id string) (model.User, error)
	GetBalanceCase(id string) (model.UserSaldo, error)
	UpdateUser(id string, payload dto.UserRequestDto) (model.User, error)
	VerifyUser(payload dto.VerifyUser) (dto.VerifyUser, error)
	UpdatePinUser(payload dto.UpdatePinRequest) (dto.UpdatePinResponse, error)
	FindRekening(id string) (model.Rekening, error)
	CreateRekening(payload model.Rekening) (model.Rekening, error)
}

type userUseCase struct {
	repo repository.UserRepository
}

func (u *userUseCase) FindById(id string) (model.User, error) {
	user, err := u.repo.Get(id)
	if err != nil {
		return model.User{}, fmt.Errorf("user with ID %s not found", id)
	}

	return user, nil
}

func (u *userUseCase) CreateUser(payload dto.UserRequestDto) (model.User, error) {
	hashPassword, err := encryption.HashPassword(payload.Password)
	if err != nil {
		return model.User{}, err
	}
	newUser := dto.UserRequestDto{
		Id:          payload.Id,
		Name:        payload.Name,
		Username:    payload.Username,
		Password:    hashPassword,
		Role:        payload.Role,
		Email:       payload.Email,
		PhoneNumber: payload.PhoneNumber,
	}
	user, err := u.repo.Create(newUser)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to create user: %v", err.Error())
	}
	return user, nil
}

func (u *userUseCase) LoginUser(in dto.LoginRequestDto) (dto.LoginResponseDto, error) {
	userData, err := u.repo.GetByUsername(in.Username)
	if err != nil {
		return dto.LoginResponseDto{}, err
	}
	isValid := encryption.CheckPasswordHash(in.Pass, userData.Password)
	if !isValid {
		return dto.LoginResponseDto{}, errors.New("1")
	}

	loginExpDuration, _ := strconv.Atoi(os.Getenv("TOKEN_LIFE_TIME"))
	expiredAt := time.Now().Add(time.Duration(loginExpDuration) * time.Minute).Unix()
	accessToken, err := common.GenerateTokenJwt(userData.Id, userData.Name, userData.Role, expiredAt)
	if err != nil {
		return dto.LoginResponseDto{}, err
	}
	return dto.LoginResponseDto{
		AccessToken: accessToken,
		UserId:      userData.Id,
	}, nil
}

func (u *userUseCase) GetBalanceCase(id string) (model.UserSaldo, error) {
	response, err := u.repo.GetBalance(id)
	if err != nil {
		return model.UserSaldo{}, err
	}

	return response, nil
}

func (u *userUseCase) UpdateUser(id string, payload dto.UserRequestDto) (model.User, error) {
	updatedUser := model.User{}
	updatedUser, err := u.repo.Get(id)
	if err != nil {
		return model.User{}, err
	}
	if payload.Name != "" {
		updatedUser.Name = payload.Name
	}
	if payload.Email != "" {
		updatedUser.Email = payload.Email
	}
	if payload.PhoneNumber != "" {
		updatedUser.PhoneNumber = payload.PhoneNumber
	}
	updatedUser.Id = id

	user, err := u.repo.Update(id, updatedUser)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to update user : %v", err.Error())
	}
	return user, nil
}

func (u *userUseCase) VerifyUser(payload dto.VerifyUser) (dto.VerifyUser, error) {
	response, err := u.repo.Verify(payload)
	if err != nil {
		return dto.VerifyUser{}, err
	}
	return response, nil
}

func (u *userUseCase) UpdatePinUser(payload dto.UpdatePinRequest) (dto.UpdatePinResponse, error) {
	response, err := u.repo.UpdatePin(payload)
	if err != nil {
		return dto.UpdatePinResponse{}, err
	}
	return response, nil
}

func (u *userUseCase) FindRekening(id string) (model.Rekening, error) {
	res, err := u.repo.GetRekening(id)
	if err != nil {
		return model.Rekening{}, err
	}
	return res, nil
}

func (u *userUseCase) CreateRekening(payload model.Rekening) (model.Rekening, error) {
	res, err := u.repo.CreateRekening(payload)
	if err != nil {
		return model.Rekening{}, err
	}

	return res, nil
}

func NewUserUseCase(repo repository.UserRepository) UserUseCase {
	return &userUseCase{repo: repo}
}
