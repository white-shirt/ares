package util

import (
	"crypto/md5"
	"errors"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func GetCurrentExecutable() string {
	return filepath.Base(os.Args[0])
}

func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\"`)
	}
	return string(path[0 : i+1]), nil
}

func MD5(raw string) string {
	md5Ctx := md5.New()
	md5Ctx.Write([]byte(raw))
	return string(md5Ctx.Sum(nil))
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(length uint8) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
