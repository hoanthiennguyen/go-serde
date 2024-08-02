package main

import (
	"encoding/json"
	"testing"
)

type Class struct {
	Name     string
	Students []*User
}

type User struct {
	Name string
	Age  int
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
			name: "nested struct",
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
