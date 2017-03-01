package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/krinklesaurus/jwt_proxy"
	"github.com/krinklesaurus/jwt_proxy/log"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
)

func NewFacebook(rootURI string, clientID string, clientSecret string, scopes []string) app.Provider {
	return &FacebookProvider{conf: oauth2.Config{
		RedirectURL:  rootURI + "/callback/facebook",
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       scopes,
		Endpoint:     facebook.Endpoint,
	}}
}

type FacebookProvider struct {
	conf  oauth2.Config
	token *oauth2.Token
}

func (f *FacebookProvider) AuthCodeURL(state string) string {
	return f.conf.AuthCodeURL(state)
}

func (f *FacebookProvider) UniqueUserID() (string, error) {
	url := fmt.Sprintf("https://graph.facebook.com/v2.7/me?access_token=%s", f.token.AccessToken)

	response, err := http.Get(url)
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", nil
	}

	log.Debugf("contents from facebook: %s", contents)

	dec := json.NewDecoder(bytes.NewReader(contents))
	var asMap map[string]string
	dec.Decode(&asMap)
	return asMap["id"], nil
}

func (f *FacebookProvider) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	token, err := f.conf.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}
	f.token = token
	return f.token, err
}