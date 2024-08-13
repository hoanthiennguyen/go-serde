package main

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestDeserailize(t *testing.T) {
	type args struct {
		raw  string
		dest any
	}
	x := 0
	xPtr := &x
	tests := []struct {
		name    string
		args    args
		wantErr bool
		want    any
	}{
		{
			name: "int",
			args: args{
				raw:  "1",
				dest: new(int),
			},
			wantErr: false,
			want:    1,
		},
		{
			name: "int ptr",
			args: args{
				raw:  "1",
				dest: &xPtr,
			},
			wantErr: false,
			want:    1,
		},
		{
			name: "float",
			args: args{
				raw:  "1.5",
				dest: new(float32),
			},
			wantErr: false,
			want:    1.5,
		},
		{
			name: "string",
			args: args{
				raw:  "\"aaa\"",
				dest: new(string),
			},
			wantErr: false,
			want:    "aaa",
		},
		{
			name: "boolean",
			args: args{
				raw:  "true",
				dest: new(bool),
			},
			wantErr: false,
			want:    true,
		},
		{
			name: "array",
			args: args{
				raw:  "[1,2,3]",
				dest: new([]int),
			},
			wantErr: false,
			want:    []int{1, 2, 3},
		},
		{
			name: "array pointer",
			args: args{
				raw:  "[1,2,3]",
				dest: new([]*int),
			},
			wantErr: false,
			want:    []int{1, 2, 3},
		},
		{
			name: "array 2 pointer",
			args: args{
				raw:  "[1,2,3]",
				dest: new([]**int),
			},
			wantErr: false,
			want:    []int{1, 2, 3},
		},
		{
			name: "array of struct",
			args: args{
				raw:  `[{"name":"John","age":30},{"name":"Doe","age":25}]`,
				dest: new([]User),
			},
			wantErr: false,
			want: []User{
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
		{
			name: "array of pointer struct",
			args: args{
				raw:  `[{"name":"John","age":30},{"name":"Doe","age":25}]`,
				dest: new([]*User),
			},
			wantErr: false,
			want: []*User{
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
		{
			name: "nested struct with array",
			args: args{
				raw:  `{"name":"Math","students":[{"name":"John","age":30},{"name":"Doe","age":25}]}`,
				dest: new(Class),
			},
			wantErr: false,
			want: Class{
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
		{
			name: "nested struct",
			args: args{
				raw:  `{"id":"abcdef","user":{"name":"Jonh","age":25}}`,
				dest: new(Account),
			},
			wantErr: false,
			want: &Account{
				ID: "abcdef",
				User: &User{
					Name: "Jonh",
					Age:  25,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Deserailize(tt.args.raw, tt.args.dest); (err != nil) != tt.wantErr {
				t.Errorf("Deserailize() error: %v, wantErr %v", err, tt.wantErr)
			}

			rawDest, _ := json.Marshal(tt.args.dest)
			rawWant, _ := json.Marshal(tt.want)
			if string(rawDest) != string(rawWant) {
				t.Errorf("Deserailize dest: %s, want: %s", rawDest, rawWant)
			}
		})
	}
}

func Test_separateKeyVal(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "struct",
			args: args{
				src: `{"name":"John","age":30}`,
			},
			want: map[string]string{
				"name": `"John"`,
				"age":  `30`,
			},
			wantErr: false,
		},
		{
			name: "struct with last val len = 1",
			args: args{
				src: `{"name":"John","age":3}`,
			},
			want: map[string]string{
				"name": `"John"`,
				"age":  `3`,
			},
			wantErr: false,
		},
		{
			name: "struct 2",
			args: args{
				src: `{"age":30,"name":"John"}`,
			},
			want: map[string]string{
				"age":  `30`,
				"name": `"John"`,
			},
			wantErr: false,
		},
		{
			name: "struct with space",
			args: args{
				src: `{"age":30, "name": "John"}`,
			},
			want: map[string]string{
				"age":  `30`,
				"name": `"John"`,
			},
			wantErr: false,
		},
		{
			name: "struct with space 2",
			args: args{
				src: `{"age": 30, "name": "John"}`,
			},
			want: map[string]string{
				"age":  `30`,
				"name": `"John"`,
			},
			wantErr: false,
		},
		{
			name: "struct with space 3",
			args: args{
				src: `{"age": 30,  "name": "John"}`,
			},
			want: map[string]string{
				"age":  `30`,
				"name": `"John"`,
			},
			wantErr: false,
		},
		{
			name: "struct with array",
			args: args{
				src: `{"name":"John","age":[1,2,3]}`,
			},
			want: map[string]string{
				"name": `"John"`,
				"age":  `[1,2,3]`,
			},
			wantErr: false,
		},
		{
			name: "struct with array string",
			args: args{
				src: `{"name":"John","age":["1","2"]}`,
			},
			want: map[string]string{
				"name": `"John"`,
				"age":  `["1","2"]`,
			},
			wantErr: false,
		},
		{
			name: "nested struct",
			args: args{
				src: `{"name":"Math","students":[{"name":"John","age":30},{"name":"Doe","age":25}]}`,
			},
			want: map[string]string{
				"name":     `"Math"`,
				"students": `[{"name":"John","age":30},{"name":"Doe","age":25}]`,
			},
			wantErr: false,
		},
		{
			name: "nested struct with space",
			args: args{
				src: `{"name":"Math","students":[{"name":"John","age":30},{"name":"Doe","age":25}] , "time": 11}`,
			},
			want: map[string]string{
				"name":     `"Math"`,
				"students": `[{"name":"John","age":30},{"name":"Doe","age":25}]`,
				"time":     "11",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := separateKeyVal(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("separateKeyVal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("separateKeyVal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_separateElements(t *testing.T) {
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "array of int",
			args: args{
				src: "[1,2,3]",
			},
			want: []string{"1", "2", "3"},
		},
		{
			name: "array of int with space",
			args: args{
				src: "[1, 2, 3]",
			},
			want: []string{"1", "2", "3"},
		},
		{
			name: "array of string",
			args: args{
				src: `["a","b","c"]`,
			},
			want: []string{`"a"`, `"b"`, `"c"`},
		},
		{
			name: "array of struct",
			args: args{
				src: `[{"name":"John","age":30},{"name":"Doe","age":25}]`,
			},
			want: []string{`{"name":"John","age":30}`, `{"name":"Doe","age":25}`},
		},
		{
			name: "array of struct with space",
			args: args{
				src: `[{"name": "John", "age": 30}, {"name":"Doe","age":25}]`,
			},
			want: []string{`{"name": "John", "age": 30}`, `{"name":"Doe","age":25}`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := separateElements(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("separateElements() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("separateElements() = %v, want %v", got, tt.want)
			}
		})
	}
}
