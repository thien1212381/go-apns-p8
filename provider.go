package apns

import (
	"crypto/ecdsa"
	"io/ioutil"
	"encoding/pem"
	"crypto/x509"
	"bytes"
	"encoding/json"
	"encoding/base64"
	"time"
	"crypto"
	"encoding/asn1"
	"crypto/rand"
	"math/big"
)


type ProviderToken struct {
	keyID 		string
	teamID 		string
	privateKey 	*ecdsa.PrivateKey
}

func NewProvierToken(path string, keyID string, teamID string) (*ProviderToken, error){
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return &ProviderToken{ keyID: keyID, teamID: teamID, privateKey: key.(*ecdsa.PrivateKey) },nil
}

func (this *ProviderToken) GetJWT() (string,error) {
	var jwtJson bytes.Buffer
	header,err := this.getHeader()
	if err!=nil {
		return "", err
	}
	jwtJson.Write(header)

	jwtJson.WriteString(".")

	claim, err := this.getClaim()
	if err!=nil {
		return "", err
	}
	jwtJson.Write(claim)

	sign, err := this.signSHA(jwtJson.Bytes())
	if err!=nil {
		return "", err
	}

	jwtJson.WriteString(".")
	jwtJson.Write(sign)
	return jwtJson.String(), nil
}

func (this *ProviderToken) getHeader() ([]byte, error){
	header := struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	} {
		"ES256",
		this.keyID,
	}
	var headerByte bytes.Buffer

	headerJson, err :=  json.Marshal(&header)
	if err!=nil {
		return nil, err
	}

	headerEnc := base64.NewEncoder(base64.RawURLEncoding, &headerByte)
	headerEnc.Write(headerJson)
	headerEnc.Close()
	return headerByte.Bytes(),nil
}

func (this *ProviderToken) getClaim() ([]byte, error) {
	claim := struct {
		Iss string `json:"iss"`
		Iat int64 `json:"iat"`
	} {
		this.teamID,
		time.Now().Unix(),
	}
	var claimByte bytes.Buffer

	claimJson, err := json.Marshal(&claim)
	if err!=nil {
		return nil, err
	}

	claimEnc := base64.NewEncoder(base64.RawURLEncoding, &claimByte)
	claimEnc.Write(claimJson)
	claimEnc.Close()
	return claimByte.Bytes(), nil
}

func (this *ProviderToken) signSHA(headerAndClaim []byte) ([]byte, error) {
	sha := crypto.SHA256.New()
	sha.Write(headerAndClaim)
	msg := sha.Sum(nil)

	r, s, err := ecdsa.Sign(rand.Reader, this.privateKey, msg)
	if err != nil {
		return nil, err
	}

	sig, err := asn1.Marshal(struct {
		R,S *big.Int
	}{
		r,s,
	})

	if err != nil {
		return nil, err
	}

	var signByte bytes.Buffer
	signEnc := base64.NewEncoder(base64.RawURLEncoding, &signByte)
	signEnc.Write(sig)
	signEnc.Close()

	return signByte.Bytes(), nil
}