package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/isd-sgcu/rpkm66-file/constant/file"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GetObjectName(filename string, secret string, fileType file.Type) (string, error) {
	text := fmt.Sprintf("%s%s%v", filename, secret, time.Now().Unix())
	hashed := Hash([]byte(text))

	hashed = strings.ReplaceAll(hashed, "/", "")

	switch fileType {
	case file.FILE:
		return fmt.Sprintf("file-%s-%d-%s", filename, time.Now().Unix(), hashed), nil
	case file.IMAGE:
		return fmt.Sprintf("image-%s-%d-%s", filename, time.Now().Unix(), hashed), nil
	default:
		return "", status.Error(codes.InvalidArgument, "Invalid file type")
	}
}
