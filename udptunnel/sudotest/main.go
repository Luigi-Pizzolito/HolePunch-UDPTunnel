package main

import (
    "fmt"
    "strings"
    "syscall"

    "golang.org/x/term"
)

func main() {
    password, _ := getPasswd()
    fmt.Printf("Password: %s\n", password)
}

func getPasswd() (string, error) {
    fmt.Print("Enter Password: ")
    bytePassword, err := term.ReadPassword(int(syscall.Stdin))
    if err != nil {
        return "", err
    }

    password := string(bytePassword)
    return strings.TrimSpace(password), nil
}