package ecoflow

import (
	"testing"
)

// accessKey=Fp4SvIprYSDPXtYJidEtUAd1o
// secretKey=WIbFEKre0s6sLnh4ei7SPUeYnptHG6V
// nonce=345164
// timestamp=1671171709428

// // HTTP json Body
// {
//     "sn" : "123456789" ,
//     "params" :{
//         "cmdSet" :11,
//         "id" :24,
//         "eps" :0
//     }
// }

// // key-value string
// params.cmdSet=11&params.eps=0&params.id=24&sn=123456789&accessKey=Fp4SvIprYSDPXtYJidEtUAd1o&nonce=345164&timestamp=1671171709428

// // sign. If the signature you generated is this value, then it is correct
// 07c13b65e037faf3b153d51613638fa80003c4c38d2407379a7f52851af1473e

func TestSign(t *testing.T) {
	params := make(map[string]string)
	params["params.cmdSet"] = "11"
	params["params.eps"] = "0"
	params["params.id"] = "24"
	params["sn"] = "123456789"

	got := sign(params, "Fp4SvIprYSDPXtYJidEtUAd1o", "WIbFEKre0s6sLnh4ei7SPUeYnptHG6V", "345164", "1671171709428")
	expected := "07c13b65e037faf3b153d51613638fa80003c4c38d2407379a7f52851af1473e"
	if expected != got {
		t.Errorf("Sign: получено %q, ожидалось %q", got, expected)
	}

}

func TestSignWithoutParams(t *testing.T) {
	params := make(map[string]string)

	got := sign(params, "Fp4SvIprYSDPXtYJidEtUAd1o", "WIbFEKre0s6sLnh4ei7SPUeYnptHG6V", "345164", "1671171709428")
	expected := "4409ae9efeaf0be9b8b48f02ebef8e804387e745e513588c136955606ca496f5"
	if expected != got {
		t.Errorf("Sign: получено %q, ожидалось %q", got, expected)
	}

}
