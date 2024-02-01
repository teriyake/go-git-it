package gitauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"teriyake/go-git-it/config"
	"time"
)

const (
	CLIENT_ID    = "Iv1.c83e19acec653315"
	CLIEN_SECRET = "secret"
)

func parseResponse(response *http.Response) (map[string]interface{}, error) {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == http.StatusOK || response.StatusCode == http.StatusCreated {
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, err
		}
		return data, nil
	} else if response.StatusCode == http.StatusUnauthorized {
		fmt.Println("You are not authorized. Run the `login` command.")
		os.Exit(1)
	} else {
		fmt.Println(response)
		fmt.Println(string(body))
		os.Exit(1)
	}
	return nil, nil
}

func requestDeviceCode() (map[string]interface{}, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://github.com/login/device/code", bytes.NewBufferString("client_id="+CLIENT_ID))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return parseResponse(resp)
}

func requestToken(deviceCode string) (map[string]interface{}, error) {
	client := &http.Client{}
	data := "client_id=" + CLIENT_ID + "&device_code=" + deviceCode + "&grant_type=urn:ietf:params:oauth:grant-type:device_code"
	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", bytes.NewBufferString(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return parseResponse(resp)
}

func pollForToken(deviceCode string, interval int) {
	for {
		response, err := requestToken(deviceCode)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		errorType, ok := response["error"].(string)
		if ok {
			switch errorType {
			case "authorization_pending":
				time.Sleep(time.Duration(interval) * time.Second)
				continue
			case "slow_down":
				time.Sleep(time.Duration(interval+5) * time.Second)
				continue
			case "expired_token":
				fmt.Println("The device code has expired. Please run `login` again.")
				os.Exit(1)
			case "access_denied":
				fmt.Println("Login cancelled by user.")
				os.Exit(1)
			default:
				fmt.Println(response)
				os.Exit(1)
			}
		}

		accessToken, ok := response["access_token"].(string)
		if ok {
			tokenPath := filepath.Join(os.Getenv("HOME"), ".go-git-it", ".token")
			ioutil.WriteFile(tokenPath, []byte(accessToken), 0600)
			break
		}
	}
}

func Login() {
	deviceCodeResponse, err := requestDeviceCode()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	verificationURI, ok1 := deviceCodeResponse["verification_uri"].(string)
	userCode, ok2 := deviceCodeResponse["user_code"].(string)
	deviceCode, ok3 := deviceCodeResponse["device_code"].(string)
	interval, ok4 := deviceCodeResponse["interval"].(float64)
	if !ok1 || !ok2 || !ok3 || !ok4 {
		fmt.Println("Error parsing response.")
		os.Exit(1)
	}

	fmt.Printf("Please visit: %s\nand enter code: %s\n", verificationURI, userCode)

	pollForToken(deviceCode, int(interval))

	fmt.Printf("Successfully authenticated! Configurating username...\n")

	profile, err := config.LoadUserProfile()
	if err != nil {
		fmt.Printf("failed to load user profile with %v\n", err)
		os.Exit(1)
	}
	username := Whoami()
	profile.SetUsername(username)
	if err := profile.Save(); err != nil {
		fmt.Printf("failed to save user profile with %v\n", err)
		os.Exit(1)
	}
}

func Whoami() string {
	tokenPath := filepath.Join(os.Getenv("HOME"), ".go-git-it", ".token")
	token, err := ioutil.ReadFile(tokenPath)
	if err != nil {
		fmt.Println("You are not authorized. Run the `login` command.")
		os.Exit(1)
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	req.Header.Set("Authorization", "Bearer "+string(token))
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	response, err := parseResponse(resp)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	login, ok := response["login"].(string)
	if !ok {
		fmt.Println("Error parsing response.")
		os.Exit(1)
	}
	return login
}

func GetJWT() string {

	pemFilePath := ".env"
	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iat": jwt.NewNumericDate(now.Add(-time.Minute)),
		"exp": jwt.NewNumericDate(now.Add(5 * time.Minute)),
		"iss": CLIENT_ID,
	})

	pemKey, _ := ioutil.ReadFile(pemFilePath)

	privateKey, _ := jwt.ParseRSAPrivateKeyFromPEM(pemKey)

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		panic(err)
	}

	return tokenString
}
