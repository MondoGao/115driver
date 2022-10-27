package driver

import (
	"net/http"
	neturl "net/url"
	"strings"

	"github.com/pkg/errors"
)

func (c *Pan115Client) LoginCheck() error {
	result := LoginResp{}
	req := c.GetRequest().
		SetQueryParam("_", NowMilli().String()).
		SetResult(&result)
	resp, err := req.Get(ApiLoginCheck)
	return CheckErr(err, &result, resp)
}

func (c *Pan115Client) ImportCredential(cr *Credential) *Pan115Client {
	cookies := map[string]string{
		CookieNameUid:  cr.UID,
		CookieNameCid:  cr.CID,
		CookieNameSeid: cr.SEID,
	}
	c.ImportCookies(cookies, CookieDomain115, CookieDomainAnxia)
	return c
}

func (c *Pan115Client) ImportCookies(cookies map[string]string, domains ...string) {
	for _, domain := range domains {
		c.importCookies(cookies, domain, "/")
	}
}

func (c *Pan115Client) importCookies(cookies map[string]string, domain string, path string) {
	// Make a dummy URL for saving cookie
	url := &neturl.URL{
		Scheme: "https",
		Path:   "/",
	}
	if domain[0] == '.' {
		url.Host = "www" + domain
	} else {
		url.Host = domain
	}
	// Prepare cookies
	cks := make([]*http.Cookie, 0, len(cookies))
	for name, value := range cookies {
		cookie := &http.Cookie{
			Name:     name,
			Value:    value,
			Domain:   domain,
			Path:     path,
			HttpOnly: true,
		}
		cks = append(cks, cookie)
	}
	// Save cookies
	c.SetCookies(cks...)
}

type Credential struct {
	UID  string
	CID  string
	SEID string
}

func (cr *Credential) FromCookie(cookie string) error {
	items := strings.Split(cookie, ";")
	if len(items) < 3 {
		return errors.Wrap(ErrBadCookie, "number of cookie paris < 3")
	}

	cookieMap := map[string]string{}
	for _, item := range items {
		pairs := strings.Split(strings.TrimSpace(item), "=")
		if len(pairs) != 2 {
			return ErrBadCookie
		}
		key := pairs[0]
		value := pairs[1]
		cookieMap[strings.ToUpper(key)] = value
	}
	cr.UID = cookieMap["UID"]
	cr.CID = cookieMap["CID"]
	cr.SEID = cookieMap["SEID"]
	if cr.CID == "" || cr.UID == "" || cr.SEID == "" {
		return errors.Wrap(ErrBadCookie, "bad cookie, miss UID, CID or SEID")
	}
	return nil
}
