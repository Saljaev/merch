package usecase

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"merch/internal/entity"
	"merch/internal/usecase/mocks"

	"reflect"
	"testing"
)

func TestNewStorage(t *testing.T) {
	mockRepo := new(mocks.UserRepo)
	mockShop := new(mocks.ShopRepo)
	mockCache := new(mocks.Cache[string])

	type args struct {
		repo UserRepo
		c    Cache[string]
		shop ShopRepo
	}
	tests := []struct {
		want *UseCase
		args args
		name string
	}{
		{
			name: "Valid test",
			args: args{
				repo: mockRepo,
				c:    mockCache,
				shop: mockShop,
			},
			want: &UseCase{repo: mockRepo, cache: mockCache, shop: mockShop},
		},
		{
			name: "Nil repo",
			args: args{
				repo: nil,
				c:    nil,
				shop: nil,
			},
			want: &UseCase{
				repo:  nil,
				cache: nil,
				shop:  nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewStorage(tt.args.repo, tt.args.c, tt.args.shop)
			if (got == nil) != (tt.want == nil) {
				t.Errorf("NewStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUseCase_AddUser(t *testing.T) {
	mockRepo := new(mocks.UserRepo)
	mockShop := new(mocks.ShopRepo)
	mockCache := new(mocks.Cache[string])

	entity.InitSonyflake()

	testUseCase := UseCase{
		repo:  mockRepo,
		shop:  mockShop,
		cache: mockCache,
	}

	uPass := "1234"
	user, err := entity.NewUser("testuser", uPass)
	assert.NoError(t, err)

	type args struct {
		ctx      context.Context
		username string
		password string
	}
	tests := []struct {
		name      string
		mockSetup func()
		useCase   UseCase
		args      args
		wantErr   bool
	}{
		{
			name: "Successful user add",
			mockSetup: func() {
				mockCache.On("Get", mock.Anything).Return(nil, false).Once()
				mockRepo.On("AddUser", mock.Anything, mock.Anything).Return(nil).Once()
				mockCache.On("Set", mock.Anything, mock.Anything).Return(nil).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:      context.Background(),
				username: user.UserName,
			},
			wantErr: false,
		},
		{
			name: "UserRepo returns error",
			mockSetup: func() {
				mockCache.On("Get", mock.Anything).Return(user, true).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:      context.Background(),
				username: user.UserName,
				password: user.Password,
			},
			wantErr: true,
		},
		{
			name: "User from cache correct password",
			mockSetup: func() {
				mockCache.On("Get", mock.Anything).Return(user, true).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:      context.Background(),
				username: user.UserName,
				password: user.Password,
			},
			wantErr: true,
		},
		{
			name: "User from cache incorrect password",
			mockSetup: func() {
				mockCache.On("Get", mock.Anything).Return(user, true).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:      context.Background(),
				username: user.UserName,
				password: uPass,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			u := &UseCase{
				repo:  tt.useCase.repo,
				cache: tt.useCase.cache,
				shop:  tt.useCase.shop,
			}

			_, err := u.AddUser(tt.args.ctx, tt.args.username, tt.args.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddUser() error = %v, wantErr %v", err, tt.wantErr)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUseCase_GetInfo(t *testing.T) {
	mockRepo := new(mocks.UserRepo)
	mockCache := new(mocks.Cache[string])

	testUseCase := UseCase{
		repo:  mockRepo,
		cache: mockCache,
	}

	testUsername := "testuser"

	testUser := entity.User{
		ID:       1,
		UserName: testUsername,
	}

	testToUsername := "testuser1"

	testCoinHistory := []entity.CoinHistory{
		{
			FromUser: testUsername,
			ToUser:   testToUsername,
			Amount:   100,
			Type:     "sent",
		},
	}

	type args struct {
		id       int
		ctx      context.Context
		username string
	}
	tests := []struct {
		name        string
		mockSetup   func()
		useCase     UseCase
		args        args
		wantErr     bool
		wantUser    entity.User
		wantHistory []entity.CoinHistory
	}{
		{
			name: "Successful get all info from cache and repo",
			mockSetup: func() {
				mockCache.On("Get", mock.Anything).Return(testUser, true).Once()
				mockRepo.On("GetTransaction", mock.Anything, mock.Anything).Return(testCoinHistory, nil).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:      context.Background(),
				id:       int(testUser.ID),
				username: testUser.UserName,
			},
			wantErr:     false,
			wantUser:    testUser,
			wantHistory: testCoinHistory,
		},
		{
			name: "Successful user + inventory with zero transactions get all info from cache and repo",
			mockSetup: func() {
				mockCache.On("Get", mock.Anything).Return(testUser, true).Once()
				mockRepo.On("GetTransaction", mock.Anything, mock.Anything).Return(nil, nil).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:      context.Background(),
				id:       int(testUser.ID),
				username: testUser.UserName,
			},
			wantErr:     false,
			wantUser:    testUser,
			wantHistory: nil,
		},
		{
			name: "Successful get all info from repo",
			mockSetup: func() {
				mockCache.On("Get", mock.Anything).Return(testUser, false).Once()
				mockRepo.On("GetInfo", mock.Anything, mock.Anything).Return(testUser, testCoinHistory, nil).Once()
				mockCache.On("Set", mock.Anything, mock.Anything).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:      context.Background(),
				id:       int(testUser.ID),
				username: testUser.UserName,
			},
			wantErr:     false,
			wantUser:    testUser,
			wantHistory: testCoinHistory,
		},
		{
			name: "UserRepo GetInfo return error",
			mockSetup: func() {
				mockCache.On("Get", mock.Anything).Return(testUser, false).Once()
				mockRepo.On("GetInfo", mock.Anything, mock.Anything).Return(entity.User{}, nil, errors.New("get info error")).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:      context.Background(),
				id:       int(testUser.ID),
				username: testUser.UserName,
			},
			wantErr:     true,
			wantUser:    entity.User{},
			wantHistory: nil,
		},
		{
			name: "UserRepo GetTransaction return error",
			mockSetup: func() {
				mockCache.On("Get", mock.Anything).Return(testUser, true).Once()
				mockRepo.On("GetTransaction", mock.Anything, mock.Anything).Return(nil, errors.New("get transaction error")).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:      context.Background(),
				id:       int(testUser.ID),
				username: testUser.UserName,
			},
			wantErr:     true,
			wantUser:    entity.User{},
			wantHistory: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			u := &UseCase{
				repo:  tt.useCase.repo,
				cache: tt.useCase.cache,
				shop:  tt.useCase.shop,
			}

			got, got1, err := u.GetInfo(tt.args.ctx, tt.args.id, tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantUser) {
				t.Errorf("GetInfo() got = %v, want %v", got, tt.wantUser)
			}
			if !reflect.DeepEqual(got1, tt.wantHistory) {
				t.Errorf("GetInfo() got1 = %v, want %v", got1, tt.wantHistory)
			}

		})
	}
}

func TestUseCase_Purchase(t *testing.T) {
	mockRepo := new(mocks.UserRepo)
	mockShop := new(mocks.ShopRepo)
	mockCache := new(mocks.Cache[string])

	testUseCase := UseCase{
		repo:  mockRepo,
		shop:  mockShop,
		cache: mockCache,
	}

	validItem := "validItem"
	invalidItem := "invalidItem"

	testUser := entity.User{
		ID:       1,
		Coins:    100,
		UserName: "testUser",
	}

	shopMap := map[string]int{
		validItem: testUser.Coins,
	}

	type args struct {
		ctx      context.Context
		id       int
		username string
		item     string
	}
	tests := []struct {
		name      string
		mockSetup func()
		useCase   UseCase
		args      args
		wantErr   bool
	}{
		{
			name: "Successful buy item, user from cache",
			mockSetup: func() {
				mockShop.On("GetCost", mock.Anything).Return(shopMap[validItem], nil).Once()
				mockCache.On("Get", mock.Anything).Return(testUser, true).Once()
				mockRepo.On("Purchase", mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				mockCache.On("Set", mock.Anything, mock.Anything).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:      context.Background(),
				id:       int(testUser.ID),
				username: testUser.UserName,
				item:     validItem,
			},
			wantErr: false,
		},
		{
			name: "Successful buy item, user from repo",
			mockSetup: func() {
				mockShop.On("GetCost", mock.Anything).Return(shopMap[validItem], nil).Once()
				mockCache.On("Get", mock.Anything).Return(nil, false).Once()
				mockRepo.On("GetUserByID", mock.Anything, mock.Anything).Return(testUser, nil).Once()
				mockCache.On("Set", mock.Anything, mock.Anything).Once()
				mockRepo.On("Purchase", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(nil).Once()
				mockCache.On("Set", mock.Anything, mock.Anything).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:      context.Background(),
				id:       int(testUser.ID),
				username: testUser.UserName,
				item:     validItem,
			},
			wantErr: false,
		},
		{
			name: "Try to buy not existing item",
			mockSetup: func() {
				mockShop.On("GetCost", mock.Anything).Return(shopMap[invalidItem], ErrNoItem).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:      context.Background(),
				id:       int(testUser.ID),
				username: testUser.UserName,
				item:     invalidItem,
			},
			wantErr: true,
		},
		{
			name: "Buy item when item.cost > user coins",
			mockSetup: func() {
				mockShop.On("GetCost", mock.Anything).Return(shopMap[validItem]+1, nil).Once()
				mockCache.On("Get", mock.Anything).Return(testUser, true).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:      context.Background(),
				id:       int(testUser.ID),
				username: testUser.UserName,
				item:     validItem,
			},
			wantErr: true,
		},
		{
			name: "Buy item when UserRepo return errors on GetUserByID",
			mockSetup: func() {
				mockShop.On("GetCost", mock.Anything).Return(shopMap[validItem], nil).Once()
				mockCache.On("Get", mock.Anything).Return(nil, false).Once()
				mockRepo.On("GetUserByID", mock.Anything, mock.Anything).Return(testUser, ErrUserNotFound).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:      context.Background(),
				id:       int(testUser.ID),
				username: testUser.UserName,
				item:     validItem,
			},
			wantErr: true,
		},
		{
			name: "Buy item when UserRepo return errors on Purchase",
			mockSetup: func() {
				mockShop.On("GetCost", mock.Anything).Return(shopMap[validItem], nil).Once()
				mockCache.On("Get", mock.Anything).Return(testUser, true).Once()
				mockRepo.On("Purchase", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(errors.New("failed to buy")).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:      context.Background(),
				id:       int(testUser.ID),
				username: testUser.UserName,
				item:     validItem,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			u := &UseCase{
				repo:  tt.useCase.repo,
				cache: tt.useCase.cache,
				shop:  tt.useCase.shop,
			}

			if err := u.Purchase(tt.args.ctx, tt.args.id, tt.args.username, tt.args.item); (err != nil) != tt.wantErr {
				t.Errorf("Purchase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUseCase_Transfer(t *testing.T) {
	mockRepo := new(mocks.UserRepo)
	mockCache := new(mocks.Cache[string])

	testUseCase := UseCase{
		repo:  mockRepo,
		cache: mockCache,
	}

	toUser := entity.User{
		ID:       1,
		Coins:    100,
		UserName: "toUser",
	}

	fromUser := entity.User{
		ID:       1,
		Coins:    100,
		UserName: "fromUser",
	}

	type args struct {
		amount int
		fromID int

		fromUsername string
		toUsername   string

		ctx context.Context
	}
	tests := []struct {
		name      string
		mockSetup func()
		useCase   UseCase
		args      args
		wantErr   bool
	}{
		{
			name: "Successful transfer from cache",
			mockSetup: func() {
				mockCache.On("Get", mock.Anything).Return(toUser, true).Once()
				mockCache.On("Get", mock.Anything).Return(fromUser, true).Once()
				mockRepo.On("Transfer", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				mockCache.On("Set", mock.Anything, mock.Anything).Twice()
			},
			useCase: testUseCase,
			args: args{
				ctx:          context.Background(),
				amount:       fromUser.Coins,
				fromID:       int(fromUser.ID),
				fromUsername: fromUser.UserName,
				toUsername:   toUser.UserName,
			},
			wantErr: false,
		},
		{
			name: "Successful transfer from repo",
			mockSetup: func() {
				mockCache.On("Get", mock.Anything).Return(toUser, false).Once()
				mockRepo.On("GetUserByUsername", mock.Anything, mock.Anything).Return(toUser, nil).Once()
				mockCache.On("Set", mock.Anything, mock.Anything).Once()
				mockCache.On("Get", mock.Anything).Return(fromUser, false).Once()
				mockRepo.On("GetUserByID", mock.Anything, mock.Anything).Return(fromUser, nil).Once()
				mockCache.On("Set", mock.Anything, mock.Anything).Once()
				mockRepo.On("Transfer", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
				mockCache.On("Set", mock.Anything, mock.Anything).Twice()
			},
			useCase: testUseCase,
			args: args{
				ctx:          context.Background(),
				amount:       fromUser.Coins,
				fromID:       int(fromUser.ID),
				fromUsername: fromUser.UserName,
				toUsername:   toUser.UserName,
			},
			wantErr: false,
		},
		{
			name: "Failed to get user by Username",
			mockSetup: func() {
				mockCache.On("Get", mock.Anything).Return(toUser, false).Once()
				mockRepo.On("GetUserByUsername", mock.Anything, mock.Anything).Return(toUser, ErrUserNotFound).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:          context.Background(),
				amount:       fromUser.Coins,
				fromID:       int(fromUser.ID),
				fromUsername: fromUser.UserName,
				toUsername:   toUser.UserName,
			},
			wantErr: true,
		},
		{
			name: "Failed to get user by ID",
			mockSetup: func() {
				mockCache.On("Get", mock.Anything).Return(toUser, true).Once()
				mockCache.On("Get", mock.Anything).Return(fromUser, false).Once()
				mockRepo.On("GetUserByID", mock.Anything, mock.Anything).Return(toUser, ErrUserNotFound).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:          context.Background(),
				amount:       fromUser.Coins,
				fromID:       int(fromUser.ID),
				fromUsername: fromUser.UserName,
				toUsername:   toUser.UserName,
			},
			wantErr: true,
		},
		{
			name: "Failed transfer not enough coin from cache",
			mockSetup: func() {
				mockCache.On("Get", mock.Anything).Return(toUser, true).Once()
				mockCache.On("Get", mock.Anything).Return(fromUser, true).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:          context.Background(),
				amount:       fromUser.Coins + 1,
				fromID:       int(fromUser.ID),
				fromUsername: fromUser.UserName,
				toUsername:   toUser.UserName,
			},
			wantErr: true,
		},
		{
			name: "Failed transfer not enough coin from repo",
			mockSetup: func() {
				mockCache.On("Get", mock.Anything).Return(toUser, true).Once()
				mockCache.On("Get", mock.Anything).Return(fromUser, true).Once()
				mockRepo.On("Transfer", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(ErrNotEnoughCoin).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:          context.Background(),
				amount:       fromUser.Coins + 1,
				fromID:       int(fromUser.ID),
				fromUsername: fromUser.UserName,
				toUsername:   toUser.UserName,
			},
			wantErr: true,
		},
		{
			name: "Failed transfer repo return transactions error",
			mockSetup: func() {
				mockCache.On("Get", mock.Anything).Return(toUser, true).Once()
				mockCache.On("Get", mock.Anything).Return(fromUser, true).Once()
				mockRepo.On("Transfer", mock.Anything, mock.Anything, mock.Anything,
					mock.Anything).Return(errors.New("some error from Tx")).Once()
			},
			useCase: testUseCase,
			args: args{
				ctx:          context.Background(),
				amount:       fromUser.Coins,
				fromID:       int(fromUser.ID),
				fromUsername: fromUser.UserName,
				toUsername:   toUser.UserName,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			u := &UseCase{
				repo:  tt.useCase.repo,
				cache: tt.useCase.cache,
			}

			if err := u.Transfer(tt.args.ctx, tt.args.amount, tt.args.fromID, tt.args.fromUsername,
				tt.args.toUsername); (err != nil) != tt.wantErr {
				t.Errorf("Transfer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
