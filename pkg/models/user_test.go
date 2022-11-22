package models

import (
	"log"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

var hash = "$2a$10$/pfkNRp/3k/M1NgNDWUQQulZ37Kt/f/VRdYHOVi2N4YOtI6wP4JfW"

type TestUser struct {
	ID                 string `json:"id" gorm:"primaryKey"`
	Email              string `json:"email" gorm:"unique"`
	Username           string `json:"username" gorm:"unique"`
	FirstName          string `json:"first_name"`
	LastName           string `json:"last_name"`
	ProfileDescription string `json:"profile_description"`
	UserPhotoURL       string `json:"user_photo_url"`
	ProfileType        string `json:"profile_type"`
}

// LastInsertId implements driver.Result
func (TestUser) LastInsertId() (int64, error) {
	panic("unimplemented")
}

// RowsAffected implements driver.Result
func (TestUser) RowsAffected() (int64, error) {
	panic("unimplemented")
}

func TestCreateUser(t *testing.T) {

	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		log.Println(err)
	}

	if err != nil {
		log.Println(err) // Error here
	}

	defer sqlDB.Close()
	type args struct {
		userID string
		body   User
		hash   []byte
	}
	tests := []struct {
		name    string
		args    args
		want    TestUser
		wantErr bool
	}{
		{
			name: "Test successfully create user",
			args: args{
				userID: "50569e94-124c-4f17-96f7-f5283df7d505",
				body: User{
					Email:     "admin@admin.com",
					Username:  "admin",
					FirstName: "Admin",
					LastName:  "A",
				},
				hash: []byte(hash),
			},
			want: TestUser{
				ID:          "50569e94-124c-4f17-96f7-f5283df7d505",
				Email:       "admin@admin.com",
				Username:    "admin",
				FirstName:   "Admin",
				LastName:    "A",
				ProfileType: "public",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectExec("INSERT INTO users").WithArgs(
				tt.args.userID,
				tt.args.body.Email,
				string(tt.args.hash),
				tt.args.body.Username,
				tt.args.body.FirstName,
				tt.args.body.LastName,
				"public",
				"",
				"",
			).WillReturnResult(tt.want)

			got, err := CreateUser(sqlDB, tt.args.userID, tt.args.body, tt.args.hash)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateUser() error = %v, wantErr %v", err.Error(), tt.wantErr)
				return
			}
			var gotUser = TestUser{
				ID:                 got.ID,
				Email:              got.Email,
				Username:           got.Username,
				FirstName:          got.FirstName,
				LastName:           got.LastName,
				ProfileDescription: got.ProfileDescription,
				UserPhotoURL:       got.UserPhotoURL,
				ProfileType:        got.ProfileType,
			}
			if !reflect.DeepEqual(gotUser, tt.want) {
				t.Errorf("CreateUser() = %v, want %v", gotUser, tt.want)
			}
		})
	}
}

func TestGetUserByEmail(t *testing.T) {

	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Println(err)
	}

	defer sqlDB.Close()

	type args struct {
		email string
	}
	tests := []struct {
		name    string
		args    args
		want    *sqlmock.Rows
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "Successfull search",
			args: args{
				email: "dika@bosnjak.com",
			},
			want: sqlmock.NewRows([]string{"id", "email", "password", "username", "first_name", "last_name", "profile_description", "user_photo_url", "profile_type"}).
				AddRow("40469e94-124c-4f17-96f7-f5283df7d707", "dika@bosnjak.com", "$2a$10$Oc9aQBx3ZYpJc6imqUnTE./ZkxXghU.t4atYUiZqbomAMZHuDCTFG", "dbosnjak", "Dika", "Bosnjak", "Test description", "http://res.cloudinary.com/bookingapp/image/upload/v1668097778/awugfhd8koi2nxodyj54.jpg", "public"),
			wantErr: false,
		},
		{
			name: "Test with error, no user in the database",
			args: args{
				email: "dika.bosnjak@size.ba",
			},
			want:    &sqlmock.Rows{},
			wantErr: true,
		},
		{
			name: "Test with error, empty email",
			args: args{
				email: "",
			},
			want:    &sqlmock.Rows{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectQuery(`SELECT id, email, password, username, first_name, last_name, profile_description, user_photo_url, profile_type FROM users WHERE email = ?`).WithArgs(tt.args.email).WillReturnRows(tt.want)
			_, err := GetUserByEmail(sqlDB, tt.args.email)
			if tt.wantErr {
				t.Logf("GetUserByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	if err != nil {
		log.Println(err)
	}

	defer sqlDB.Close()

	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    *sqlmock.Rows
		wantErr bool
	}{
		{
			name: "Successfull search",
			args: args{
				id: "40469e94-124c-4f17-96f7-f5283df7d707",
			},
			want: sqlmock.NewRows([]string{"id", "email", "password", "username", "first_name", "last_name", "profile_description", "user_photo_url", "profile_type"}).
				AddRow("40469e94-124c-4f17-96f7-f5283df7d707", "dika@bosnjak.com", "$2a$10$Oc9aQBx3ZYpJc6imqUnTE./ZkxXghU.t4atYUiZqbomAMZHuDCTFG", "dbosnjak", "Dika", "Bosnjak", "Test description", "http://res.cloudinary.com/bookingapp/image/upload/v1668097778/awugfhd8koi2nxodyj54.jpg", "public"),
			wantErr: false,
		},
		{
			name: "Test with error, no user in the database",
			args: args{
				id: "40469e94-124c-4f17-96f7-f5283df7d111",
			},
			want:    &sqlmock.Rows{},
			wantErr: true,
		},
		{
			name: "Test with error, empty id",
			args: args{
				id: "",
			},
			want:    &sqlmock.Rows{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.ExpectQuery(`SELECT id, email, password, username, first_name, last_name, profile_description, user_photo_url, profile_type FROM users WHERE id = ?`).WithArgs(tt.args.id).WillReturnRows(tt.want)
			_, err := GetUserByID(sqlDB, tt.args.id)
			if tt.wantErr {
				t.Logf("GetUserByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
