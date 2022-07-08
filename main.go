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
	Age   uint   `json:"age"`
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
		if !json.Valid([]byte(args["item"])) || args["item"] == "" {
			return errors.New("-item flag has to be specified")
		}

		err := add(file, args)
		if err != nil {
			writer.Write([]byte(err.Error()))
			break
		}

	case "findById":
		if !json.Valid([]byte(args["id"])) || args["id"] == "" {
			return errors.New("-id flag has to be specified")
		}
		user, err := find(file, args["id"])
		if err != nil {
			writer.Write([]byte(err.Error()))
		}
		writer.Write(user)
		fmt.Println("findById")
	case "remove":
		if !json.Valid([]byte(args["id"])) || args["id"] == "" {
			return errors.New("-id flag has to be specified")
		}
		val, _ := find(file, args["id"])
		if val == nil {
			writer.Write([]byte(fmt.Errorf("Item with id %s not found", args["id"]).Error()))
		}
		err := remove(file, args)
		if err != nil {
			writer.Write([]byte(err.Error()))
		}

	default:

		return fmt.Errorf("Operation %s not allowed!", args["operation"])
	}

	return nil
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

	var user User
	err := json.Unmarshal([]byte(arg["item"]), &user)
	if err != nil {
		return err
	}
	us, _ := find(file, user.Id)
	if us != nil {
		return fmt.Errorf("Item with id %s already exists", user.Id)
	}

	var users []User
	allUsersBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(allUsersBytes, &users)
	if err != nil {
		return err
	}

	users = append(users, user)
	newUsersBytes, err := json.Marshal(users)
	if err != nil {
		return err
	}

	_, err = file.Write(newUsersBytes)
	if err != nil {
		return err
	}

	return nil
}

func find(file *os.File, id string) ([]byte, error) {
	defer file.Seek(0, 0)
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Failed to read file, %w", err)
	}
	var users []User

	err = json.Unmarshal(data, &users)
	if err != nil {
		return nil, errors.New("-id flag has to be specified")
	}
	for _, v := range users {
		if v.Id == id {
			bytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("Failed to marshal json, %w", err)
			}
			return bytes, nil
		}
	}

	return nil, errors.New("")
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
		return errors.New("-id flag has to be specified")
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
	file.Truncate(0)
	file.Write(buffer)
	file.Seek(0, 0)
	return nil
}
