package authentication

import (
	"reflect"
	"telarr/configuration"
	"testing"
)

func TestNew(t *testing.T) {
	authPath = t.TempDir()
	conf := configuration.Configuration{
		Telegram: configuration.Telegram{
			Passwd: "password",
		},
	}

	tests := []struct {
		name    string
		want    *Auth
		wantErr bool
	}{
		{
			name: "success",
			want: &Auth{
				Attempts: make(map[int]int),
				conf:     conf,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(conf)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuth_CheckAutorized(t *testing.T) {
	conf := configuration.Configuration{
		Telegram: configuration.Telegram{
			Passwd: "password",
		},
	}

	type fields struct {
		Blacklist []int
		Autorized []int
		Attempts  map[int]int
	}
	type args struct {
		userId int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   AuthStatus
	}{
		{
			name: "not autorized",
			fields: fields{
				Blacklist: []int{1, 2},
				Autorized: []int{3, 4},
				Attempts:  make(map[int]int),
			},
			args: args{
				userId: 1,
			},
			want: AuthStatusBlackListed,
		},
		{
			name: "autorized",
			fields: fields{
				Blacklist: []int{1, 2},
				Autorized: []int{3, 4},
				Attempts:  make(map[int]int),
			},
			args: args{
				userId: 3,
			},
			want: AuthStatusAutorized,
		},
		{
			name: "new user",
			fields: fields{
				Blacklist: []int{1, 2},
				Autorized: []int{3, 4},
				Attempts:  make(map[int]int),
			},
			args: args{
				userId: 5,
			},
			want: AuthStatusNewUser,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Auth{
				Blacklist: tt.fields.Blacklist,
				Autorized: tt.fields.Autorized,
				Attempts:  tt.fields.Attempts,
				conf:      conf,
			}
			if got := a.CheckAutorized(tt.args.userId); got != tt.want {
				t.Errorf("Auth.CheckAutorized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuth_Autorize(t *testing.T) {
	authPath = t.TempDir()
	conf := configuration.Configuration{
		Telegram: configuration.Telegram{
			Passwd: "password",
		},
	}

	type fields struct {
		Blacklist []int
		Autorized []int
		Attempts  map[int]int
	}
	type args struct {
		userId   int
		password string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   AuthStatus
	}{
		{
			name: "wrong password",
			fields: fields{
				Blacklist: []int{1, 2},
				Autorized: []int{3, 4},
				Attempts:  make(map[int]int),
			},
			args: args{
				userId:   1,
				password: "wrongPassword",
			},
			want: AuthStatusWrongPassword,
		},
		{
			name: "autorized",
			fields: fields{
				Blacklist: []int{1, 2},
				Autorized: []int{3, 4},
				Attempts:  make(map[int]int),
			},
			args: args{
				userId:   3,
				password: "password",
			},
			want: AuthStatusAutorized,
		},
		{
			name: "max attempts",
			fields: fields{
				Blacklist: []int{1, 2},
				Autorized: []int{3, 4},
				Attempts:  map[int]int{5: maxAttempts},
			},
			args: args{
				userId:   5,
				password: "password",
			},
			want: AuthStatusMaxAttempts,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Auth{
				Blacklist: tt.fields.Blacklist,
				Autorized: tt.fields.Autorized,
				Attempts:  tt.fields.Attempts,
				conf:      conf,
			}
			if got, _ := a.AutorizeNewUser(tt.args.userId, tt.args.password); got != tt.want {
				t.Errorf("Auth.Autorize() = %v, want %v", got, tt.want)
			}
		})
	}
}
