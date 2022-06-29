package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type Arguments map[string]string

type User struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   string `json:"age"`
}

func Perform(args Arguments, writer io.Writer) error {
	fmt.Println(args["operation"])
	if _, ok := args["operation"]; !ok || args["operation"] == "" {
		return errors.New("-operation flag has to be specified")
	}
	if _, ok := args["fileName"]; !ok || args["fileName"] == "" {
		return errors.New("-fileName flag has to be specified")
	}
	file, err := os.OpenFile(args["fileName"], os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return fmt.Errorf("Failed to open/create file, %w", err)
	}
	defer file.Close()

	switch args["operation"] {
	case "list":
		list(file, writer)
	case "add":
		err := add(file, args)
		if err != nil {
			return err
		}
	case "findById":
		user, err := find(file, args)
		if err != nil {
			return err
		}
		writer.Write(user)
		fmt.Println("findById")
	case "remove":
		err := remove(file, args)
		if err != nil {
			return err
		}

	}
	return errors.New("Operation abcd not allowed!")
}

func main() {
	err := Perform(parseArgs(), os.Stdout)

	if err != nil {
		fmt.Println(err)
	}
}

func parseArgs() Arguments {
	id := flag.String("id", "", "User id")
	operation := flag.String("operation", "", "list - users\nadd - add user\nfindById - get user by id\nremove - remove user")
	fileName := flag.String("fileName", "", ".json file name")
	item := flag.String("item", "", "json item of user")
	flag.Parse()

	return Arguments{"id": *id, "operation": *operation, "fileName": *fileName, "item": *item}
}

func add(file *os.File, arg Arguments) error {
	if !json.Valid([]byte(arg["item"])) {
		return errors.New("-item flag has to be specified")
	}

	var user Arguments
	err := json.Unmarshal([]byte(arg["item"]), &user)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal json, %w", err)
	}

	data, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("Failed to marshal json, %w", err)
	}

	file.Write(data)

	return nil
}

func find(file *os.File, arg Arguments) ([]byte, error) {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file, %w", err)
	}
	var users []User

	err = json.Unmarshal(data, &users)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal json, %w", err)
	}
	for _, v := range users {
		if v.Id == arg["id"] {
			bytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("Failed to marshal json, %w", err)
			}
			return bytes, nil
		}
	}
	return nil, errors.New("user not found")
}

func list(file *os.File, writer io.Writer) {
	data, _ := ioutil.ReadAll(file)
	writer.Write(data)
}
func remove(file *os.File, arg Arguments) error {
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("Failed to read file, %w", err)
	}
	var users []User

	err = json.Unmarshal(data, &users)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal json, %w", err)
	}
	for i, v := range users {
		if v.Id == arg["id"] {
			users = append(users[:i], users[i+1:]...)
		}
	}

	buffer, err := json.Marshal(users)
	if err != nil {
		return fmt.Errorf("Failed to marshal json, %w", err)
	}
	file.Write(buffer)
	file.Truncate(0)
	//file.Seek(0, 0)
	return nil
}
