package apns

import (
	"net/http"
	"golang.org/x/net/http2"
	"log"
	"encoding/json"
)

const (
	Development_URL = "https://api.development.push.apple.com"
	Production_URL  = "https://api.push.apple.com"
)

type Client struct {
	host 		string
	client 		*http.Client
	CProviderToken 	*ProviderToken
}

type Response struct {
	ApnsID 		string
	StatusCode 	int
}

type Error struct {
	Reason string `json:"reason"`
	Timestamp int `json:"timestamp"`
}

func NewClient(providerToken *ProviderToken, is_production bool) (*Client, error){
	var host string

	if is_production {
		host = Production_URL
	} else {
		host = Development_URL
	}
	transport := &http.Transport{}

	if err := http2.ConfigureTransport(transport); err != nil {
		return nil,err
	}

	client := &http.Client {
		Transport: transport,
	}
	return &Client{ host, client, providerToken }, nil
}

func (this *Client) Push(notification *Notification) (*Response, *Error){
	jwt,_ := this.CProviderToken.GetJWT()
	req, err := notification.newRequest(this.host, jwt)
	if err!=nil {
		log.Fatal(err)
		return nil,&Error{ Reason: err.Error() }
	}

	resp,err := this.client.Do(req)
	if err!=nil {
		log.Fatal(err)
		return nil,&Error{ Reason: err.Error() }
	}
	defer resp.Body.Close()

	result := &Response{
		ApnsID: resp.Header.Get("apns-id"),
		StatusCode: resp.StatusCode,
	}
	if resp.StatusCode != http.StatusOK {
		var errRes Error
		if err := json.NewDecoder(resp.Body).Decode(&errRes); err!=nil {
			return nil, &Error{ Reason: err.Error() }
		}
		return result, &errRes
	}

	return result, nil
}