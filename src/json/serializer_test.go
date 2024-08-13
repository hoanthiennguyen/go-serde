package json

import (
	"encoding/json"
	"testing"
)

type Class struct {
	Name     string  `json:"name"`
	Students []*User `json:"students"`
}

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Account struct {
	ID   string `json:"id"`
	User User   `json:"user"`
}

func TestSerialize(t *testing.T) {
	type args struct {
		data any
	}
	n := 1
	tests := []struct {
		name string
		args args
	}{
		{
			name: "integer",
			args: args{data: 1},
		},
		{
			name: "pointer",
			args: args{data: &n},
		},
		{
			name: "float",
			args: args{data: 1.5},
		},
		{
			name: "string",
			args: args{data: "hello"},
		},
		{
			name: "array",
			args: args{data: []int{1, 2, 3}},
		},
		{
			name: "struct",
			args: args{
				data: User{
					Name: "John",
					Age:  30,
				},
			},
		},
		{
			name: "struct with pointer",
			args: args{
				data: &User{
					Name: "John",
					Age:  30,
				},
			},
		},
		{
			name: "struct missing field",
			args: args{
				data: &User{
					Name: "John",
				},
			},
		},
		{
			name: "nested struct with array",
			args: args{
				data: Class{
					Name: "Math",
					Students: []*User{
						{
							Name: "John",
							Age:  30,
						},
						{
							Name: "Doe",
							Age:  25,
						},
					},
				},
			},
		},
		{
			name: "nested struct",
			args: args{
				data: &Account{
					ID: "abcdef",
					User: User{
						Name: "Jonh",
						Age:  25,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Serialize(tt.args.data)
			want, _ := json.Marshal(tt.args.data)
			if got != string(want) {
				t.Errorf("Serialize() = %v, want %v", got, want)
			}
		})
	}
}
