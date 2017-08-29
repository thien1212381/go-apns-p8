package apns

import (
	"time"
	"net/http"
	"fmt"
	"bytes"
	"encoding/json"
	"strconv"
)

type Notification struct{
	ID  		string
	Token 		string
	Expiration 	time.Time
	Topic 		string
	Payload		map[string]interface{}
}

func (this *Notification) newRequest(host string, jwt string) (*http.Request, error) {
	url := fmt.Sprintf("%s/3/device/%s", host, this.Token)

	payload, err := json.Marshal(this.Payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))

	if this.ID != "" {
		req.Header.Set("apns-id", this.ID)
	}

	if !this.Expiration.IsZero() {
		var exp string = "0"
		if !this.Expiration.Before(time.Now()) {
			exp = strconv.FormatInt(this.Expiration.Unix(), 10)
		}
		req.Header.Set("apns-expiration", exp)
	}

	if this.Topic != "" {
		req.Header.Set("apns-topic", this.Topic)
	}

	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("authorization", fmt.Sprintf("bearer %s", jwt))

	return req,nil
}

