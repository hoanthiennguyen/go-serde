package main

import "fmt"

type Class struct {
	Name     string
	Students []*User
}

type User struct {
	Name string
	Age  int
}

func main() {
	students := []*User{{Name: "John", Age: 30}, {Name: "Jane", Age: 25}}
	class := &Class{Name: "Math", Students: students}
	fmt.Println(Serialize(class))
}
