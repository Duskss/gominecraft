package gominecraft

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gocolly/twocaptcha"
	"io/ioutil"
	"net/http"
	"strings"
)

type MojangError struct {
	Error        string `json:"error"`
	ErrorMessage string `json:"errorMessage"`
}

type UserResponse struct {
	ID               string `json:"id"`
	Email            string `json:"email"`
	Username         string `json:"username"`
	RegisterIP       string `json:"registerIp"`
	RegisteredAt     int64  `json:"registeredAt"`
	DateOfBirth      int64  `json:"dateOfBirth"`
	Secured          bool   `json:"secured"`
	EmailVerified    bool   `json:"emailVerified"`
	LegacyUser       bool   `json:"legacyUser"`
	VerifiedByParent bool   `json:"verifiedByParent"`
	Hashed           bool   `json:"hashed"`
}

type MinecraftResponse []struct {
	Agent         string `json:"agent"`
	ID            string `json:"id"`
	Name          string `json:"name"`
	UserID        string `json:"userId"`
	CreatedAt     int64  `json:"createdAt"`
	LegacyProfile bool   `json:"legacyProfile"`
	Suspended     bool   `json:"suspended"`
	Paid          bool   `json:"paid"`
	Migrated      bool   `json:"migrated"`
}

type Client struct {
	Captcha *twocaptcha.TwoCaptchaClient
	Client  *http.Client
}

type Session struct {
	Email, Password, Bearer string
	Client                  *http.Client
}

func (c *Client) Login(email, password string) (*Session, error) {
	c.Captcha = twocaptcha.New("6bb6c392e593b15ba6dcbd4a27d9bfe3")
	str, err := c.Captcha.SolveRecaptchaV2("https://minecraft.net/", "6LfbsiMUAAAAAOu1nGK8InBaFrIk17dcbI0sqvzj")
	if err != nil {
		return nil, err
	}

	return c.LoginC(email, password, str)
}

/*
	LoginC - Provides a method of logging in using a custom captcha solution.
	Captcha - {"sitekey": "6LfbsiMUAAAAAOu1nGK8InBaFrIk17dcbI0sqvzj"}
*/
func (c *Client) LoginC(email, password, str string) (*Session, error) {
	type Payload struct {
		Captcha          string `json:"Captcha"`
		CaptchaSupported string `json:"captchaSupported"`
		Password         string `json:"password"`
		RequestUser      bool   `json:"requestUser"`
		Username         string `json:"username"`
	}

	payloadBytes, _ := json.Marshal(Payload{
		Captcha:          str,
		CaptchaSupported: "InvisibleReCAPTCHA",
		Password:         password,
		RequestUser:      true,
		Username:         email,
	})

	req, _ := http.NewRequest("POST", "https://authserver.mojang.com/authenticate", bytes.NewReader(payloadBytes))

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36")
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var authentication struct {
		User struct {
			Username string `json:"username"`
			ID       string `json:"id"`
		} `json:"user"`
		AccessToken       string        `json:"accessToken"`
		ClientToken       string        `json:"clientToken"`
		AvailableProfiles []interface{} `json:"availableProfiles"`
	}

	body, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &authentication)

	if authentication.AccessToken == "" {
		var mojangError MojangError
		_ = json.Unmarshal(body, &mojangError)

		return nil, errors.New(mojangError.ErrorMessage)
	} else {
		return &Session{email, password, authentication.AccessToken, c.Client}, nil
	}
}

// Retrieves basic user data, incl. email, username & dob
func (s *Session) User() (*UserResponse, error) {
	req, _ := http.NewRequest("GET", "https://authserver.mojang.com/authenticate", nil)

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+s.Bearer)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userResponse UserResponse
	body, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &userResponse)

	if userResponse.ID == "" {
		var mojangError MojangError
		_ = json.Unmarshal(body, &mojangError)

		return nil, errors.New(mojangError.ErrorMessage)
	} else {
		return &userResponse, nil
	}
}

// Retrieves basic user data, incl. email, username & dob
func (s *Session) Allocate(username string) error {
	req, _ := http.NewRequest("PUT", "https://api.mojang.com/user/profile/agent/minecraft/name/"+username, nil)

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36")
	req.Header.Add("Authorization", "Bearer "+s.Bearer)

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		var mojangError MojangError
		body, _ := ioutil.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &mojangError)

		return errors.New(mojangError.ErrorMessage)
	} else {
		return nil
	}
}

// Redeem a Minecon cape using a 32 long hashed code
func (s *Session) CapeR(uuid, code string) error {
	req, _ := http.NewRequest("POST", "https://api.mojang.com/user/profile/"+uuid+"/capetokens/code/"+code, nil)

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36")
	req.Header.Add("Authorization", "Bearer "+s.Bearer)

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		var mojangError MojangError
		body, _ := ioutil.ReadAll(resp.Body)
		_ = json.Unmarshal(body, &mojangError)

		return errors.New(mojangError.ErrorMessage)
	} else {
		return nil
	}
}

func (s *Session) Minecraft() (*MinecraftResponse, error) {
	req, _ := http.NewRequest("GET", "https://api.mojang.com/user/profiles/agent/minecraft", nil)

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36")
	req.Header.Add("Authorization", "Bearer "+s.Bearer)

	resp, err := s.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var minecraftResponse MinecraftResponse
	body, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &minecraftResponse)

	if len(minecraftResponse) == 0 {
		var mojangError MojangError
		_ = json.Unmarshal(body, &mojangError)

		return nil, errors.New(mojangError.ErrorMessage)
	} else {
		return &minecraftResponse, nil
	}
}

func (s *Session) Validate() error {
	req, _ := http.NewRequest("POST", "https://authserver.mojang.com/validate", strings.NewReader(`{"accessToken":"`+s.Bearer+`"}`))

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+s.Bearer)

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var minecraftResponse MinecraftResponse
	body, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(body, &minecraftResponse)

	if resp.StatusCode != 204 {
		var mojangError MojangError
		_ = json.Unmarshal(body, &mojangError)

		return errors.New(mojangError.ErrorMessage)
	} else {
		return nil
	}
}
