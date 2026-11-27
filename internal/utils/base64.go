package utils

import "encoding/base64"

func IsBase64Encoded(data []byte) bool {
	_, err := base64.StdEncoding.DecodeString(string(data))

	return err == nil
}

func Base64Encoded(data []byte) string {
	if IsBase64Encoded(data) {
		return string(data)
	}

	return base64.StdEncoding.EncodeToString(data)
}
