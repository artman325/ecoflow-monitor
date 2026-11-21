package ecoflow

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"maps"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"time"
)

func computeHMAC256(message, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))

	rawHash := h.Sum(nil)

	return hex.EncodeToString(rawHash)

}

// sign returns the hex-encoded HMAC-SHA256 signature of the request parameters, as required by the EcoFlow API.
func sign(params map[string]string, accessKey string, secretKey string, nonce string, timestamp string) string {
	// the logic is
	// 1.  request parameters must be sorted by ASCII value and concatenated with characters =, &
	//		str=param1=value1&param2=value2
	// !notice! in params map[string]string we get already flattened    like deviceList[0].id , deviceList[1].id ,ids[0]  so we sorted it
	// 2. with others(accessKey,nonce,timestamp) should do the same  like accessKey=***&nonce=...&timestamp=
	// 3. join string from 1. and 2.
	// it can be like str like this
	// 	params.cmdSet=11&params.eps=0&params.id=24&sn=123456789&accessKey=*****&nonce=345164&timestamp=1671171709428
	// 4. get sign bytes
	// E.g. byte[] signBytes = HMAC-SHA256(str, secretKey)
	// 5. convert byte[] to hexadecimal string. String sign = bytesToHexString(signBytes)
	// E.g. sign=85776ede686fe4783eac48135b0b1748ba2d7e9bb7791b826dc942fc29d4ada8

	sortedKeyList := slices.Sorted(maps.Keys(params))

	var sortedParamsList []string
	for _, key := range sortedKeyList {
		sortedParamsList = append(sortedParamsList, key+"="+params[key])
	}
	sortedParamsList = append(sortedParamsList, "accessKey="+accessKey)
	sortedParamsList = append(sortedParamsList, "nonce="+nonce)
	sortedParamsList = append(sortedParamsList, "timestamp="+timestamp)

	payload := strings.Join(sortedParamsList, "&")
	return computeHMAC256(payload, secretKey)

}

func nonce() string {
	minNonce := 100000
	maxNonce := 999999
	return strconv.Itoa(rand.IntN(maxNonce-minNonce+1) + minNonce)
}

func timestamp() string {
	return strconv.FormatInt(time.Now().UnixMilli(), 10)
}
