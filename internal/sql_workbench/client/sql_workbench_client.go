package client

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"
	"time"

	config "github.com/actiontech/dms/internal/sql_workbench/config"
	utilLog "github.com/actiontech/dms/pkg/dms-common/pkg/log"
)

// SqlWorkbenchClient SQL工作台HTTP客户端
type SqlWorkbenchClient struct {
	cfg        *config.SqlWorkbenchOpts
	log        *utilLog.Helper
	httpClient *http.Client
	baseURL    string
}

// NewSqlWorkbenchClient 创建SQL工作台客户端
func NewSqlWorkbenchClient(cfg *config.SqlWorkbenchOpts, logger utilLog.Logger) *SqlWorkbenchClient {
	baseURL := fmt.Sprintf("http://%s:%s", cfg.Host, cfg.Port)

	return &SqlWorkbenchClient{
		cfg: cfg,
		log: utilLog.NewHelper(logger, utilLog.WithMessageKey("sql_workbench.client")),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: baseURL,
	}
}

// GetPublicKey 获取RSA公钥
func (c *SqlWorkbenchClient) GetPublicKey() (string, error) {
	c.log.Infof("Attempting to get public key")

	// 构建获取公钥URL
	publicKeyURL := fmt.Sprintf("%s/api/v2/encryption/publicKey", c.baseURL)

	// 创建GET请求
	req, err := http.NewRequest("GET", publicKeyURL, nil)
	if err != nil {
		c.log.Errorf("Failed to create get public key request: %v", err)
		return "", fmt.Errorf("failed to create get public key request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Errorf("Failed to send get public key request: %v", err)
		return "", fmt.Errorf("failed to send get public key request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Errorf("Failed to read get public key response: %v", err)
		return "", fmt.Errorf("failed to read get public key response: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		c.log.Errorf("Get public key failed with status code: %d, response: %s", resp.StatusCode, string(body))
		return "", fmt.Errorf("get public key failed with status code: %d", resp.StatusCode)
	}

	// 解析响应
	var publicKeyResp GetPublicKeyResponse
	if err := json.Unmarshal(body, &publicKeyResp); err != nil {
		c.log.Errorf("Failed to parse get public key response: %v", err)
		return "", fmt.Errorf("failed to parse get public key response: %v", err)
	}

	// 检查业务状态
	if !publicKeyResp.Successful {
		errorMsg := "get public key failed"
		if publicKeyResp.Message != nil {
			errorMsg = *publicKeyResp.Message
		}
		c.log.Errorf("Get public key failed: %s", errorMsg)
		return "", fmt.Errorf("get public key failed: %s", errorMsg)
	}

	c.log.Infof("Successfully retrieved public key")
	return publicKeyResp.Data, nil
}

// Login 登录到SQL工作台
func (c *SqlWorkbenchClient) Login(username, password, publicKey string) (*LoginResponse, error) {
	c.log.Infof("Attempting to login to SQL workbench for user: %s", username)

	// 加密密码
	encryptedPassword, err := c.EncryptPasswordWithRSA(password, publicKey)
	if err != nil {
		c.log.Errorf("Failed to encrypt password: %v", err)
		return nil, fmt.Errorf("failed to encrypt password: %v", err)
	}

	// 构建登录URL
	loginURL := fmt.Sprintf("%s/api/v2/iam/login?currentOrganizationId=&ignoreError=true", c.baseURL)

	// 准备multipart表单数据
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加表单字段
	writer.WriteField("username", username)
	writer.WriteField("password", encryptedPassword)
	writer.Close()

	// 创建请求
	req, err := http.NewRequest("POST", loginURL, &buf)
	if err != nil {
		c.log.Errorf("Failed to create login request: %v", err)
		return nil, fmt.Errorf("failed to create login request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Errorf("Failed to send login request: %v", err)
		return nil, fmt.Errorf("failed to send login request: %v", err)
	}
	defer resp.Body.Close()
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Errorf("Failed to read login response: %v", err)
		return nil, fmt.Errorf("failed to read login response: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		c.log.Errorf("Login failed with status code: %d, response: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("login failed with status code: %d", resp.StatusCode)
	}

	// 解析响应
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		c.log.Errorf("Failed to parse login response: %v", err)
		return nil, fmt.Errorf("failed to parse login response: %v", err)
	}

	// 检查业务状态
	if !loginResp.Successful {
		errorMsg := "login failed"
		if loginResp.Message != nil {
			errorMsg = *loginResp.Message
		}
		c.log.Errorf("Login failed: %s", errorMsg)
		return nil, fmt.Errorf("login failed: %s", errorMsg)
	}

	loginResp.Cookie = resp.Header.Get("Set-Cookie")

	c.log.Infof("Successfully logged in to SQL workbench for user: %s", username)
	return &loginResp, nil
}

// GetOrganizations 获取组织信息
func (c *SqlWorkbenchClient) GetOrganizations(cookie string) (*GetOrganizationsResponse, error) {
	c.log.Infof("Attempting to get organizations")

	// 构建获取组织URL
	organizationsURL := fmt.Sprintf("%s/api/v2/iam/users/me/organizations?currentOrganizationId=", c.baseURL)

	// 创建GET请求
	req, err := http.NewRequest("GET", organizationsURL, nil)
	if err != nil {
		c.log.Errorf("Failed to create get organizations request: %v", err)
		return nil, fmt.Errorf("failed to create get organizations request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Errorf("Failed to send get organizations request: %v", err)
		return nil, fmt.Errorf("failed to send get organizations request: %v", err)
	}
	defer resp.Body.Close()
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Errorf("Failed to read get organizations response: %v", err)
		return nil, fmt.Errorf("failed to read get organizations response: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		c.log.Errorf("Get organizations failed with status code: %d, response: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("get organizations failed with status code: %d", resp.StatusCode)
	}

	// 解析响应
	var organizationsResp GetOrganizationsResponse
	if err := json.Unmarshal(body, &organizationsResp); err != nil {
		c.log.Errorf("Failed to parse get organizations response: %v", err)
		return nil, fmt.Errorf("failed to parse get organizations response: %v", err)
	}
	organizationsResp.XsrfToken = resp.Header.Get("Set-Cookie")
	// 检查业务状态
	if !organizationsResp.Successful {
		errorMsg := "get organizations failed"
		if organizationsResp.Message != nil {
			errorMsg = *organizationsResp.Message
		}
		c.log.Errorf("Get organizations failed: %s", errorMsg)
		return nil, fmt.Errorf("get organizations failed: %s", errorMsg)
	}

	c.log.Infof("Successfully retrieved %d organizations", len(organizationsResp.Data.Contents))
	return &organizationsResp, nil
}

// Environment 环境信息
type Environment struct {
	BuiltIn        bool    `json:"builtIn"`
	CreateTime     int64   `json:"createTime"`
	Creator        *User   `json:"creator"`
	Description    string  `json:"description"`
	Enabled        bool    `json:"enabled"`
	ID             int64   `json:"id"`
	LastModifier   *User   `json:"lastModifier"`
	Name           string  `json:"name"`
	OrganizationID int64   `json:"organizationId"`
	OriginalName   string  `json:"originalName"`
	RulesetID      int64   `json:"rulesetId"`
	RulesetName    *string `json:"rulesetName"`
	Style          string  `json:"style"`
	UpdateTime     int64   `json:"updateTime"`
}

// GetEnvironmentsResponse 获取环境列表响应
type GetEnvironmentsResponse struct {
	Data struct {
		Contents []Environment `json:"contents"`
	} `json:"data"`
	DurationMillis int64   `json:"durationMillis"`
	HTTPStatus     string  `json:"httpStatus"`
	RequestID      string  `json:"requestId"`
	Server         string  `json:"server"`
	Successful     bool    `json:"successful"`
	Timestamp      float64 `json:"timestamp"`
	TraceID        string  `json:"traceId"`
}

// GetEnvironments 获取环境列表
func (c *SqlWorkbenchClient) GetEnvironments(organizationID int64, cookie string) (*GetEnvironmentsResponse, error) {
	c.log.Infof("Attempting to get environments for organization %d", organizationID)

	// 构建获取环境列表URL
	environmentsURL := fmt.Sprintf("%s/api/v2/collaboration/environments?currentOrganizationId=%d&enabled=true", c.baseURL, organizationID)

	// 创建请求
	req, err := http.NewRequest("GET", environmentsURL, nil)
	if err != nil {
		c.log.Errorf("Failed to create get environments request: %v", err)
		return nil, fmt.Errorf("failed to create get environments request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("X-Xsrf-Token", c.ExtractCookieValue(cookie, "XSRF-TOKEN"))

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Errorf("Failed to send get environments request: %v", err)
		return nil, fmt.Errorf("failed to send get environments request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Errorf("Failed to read get environments response: %v", err)
		return nil, fmt.Errorf("failed to read get environments response: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		c.log.Errorf("Get environments failed with status code: %d, response: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("get environments failed with status code: %d", resp.StatusCode)
	}

	// 解析响应
	var environmentsResp GetEnvironmentsResponse
	if err := json.Unmarshal(body, &environmentsResp); err != nil {
		c.log.Errorf("Failed to parse get environments response: %v", err)
		return nil, fmt.Errorf("failed to parse get environments response: %v", err)
	}

	// 检查业务状态
	if !environmentsResp.Successful {
		c.log.Errorf("Get environments failed: %s", environmentsResp.HTTPStatus)
		return nil, fmt.Errorf("get environments failed: %s", environmentsResp.HTTPStatus)
	}

	c.log.Infof("Successfully got %d environments", len(environmentsResp.Data.Contents))
	return &environmentsResp, nil
}

// CreateDatasources 创建数据源
func (c *SqlWorkbenchClient) CreateDatasources(datasource CreateDatasourceRequest, publicKey string, cookie string, organizationID int64) (*CreateDatasourceResponse, error) {
	c.log.Infof("Attempting to create datasource: %s", datasource.Name)

	// 加密密码
	encryptedPassword, err := c.EncryptPasswordWithRSA(datasource.Password, publicKey)
	if err != nil {
		c.log.Errorf("Failed to encrypt password for datasource %s: %v", datasource.Name, err)
		return nil, fmt.Errorf("failed to encrypt password for datasource %s: %v", datasource.Name, err)
	}
	datasource.Password = encryptedPassword

	// 构建创建数据源URL
	createDatasourceURL := fmt.Sprintf("%s/api/v2/datasource/datasources?currentOrganizationId=%d&wantCatchError=false&holdErrorTip=true", c.baseURL, organizationID)

	// 准备JSON请求体
	jsonData, err := json.Marshal(datasource)
	if err != nil {
		c.log.Errorf("Failed to marshal datasource data: %v", err)
		return nil, fmt.Errorf("failed to marshal datasource data: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", createDatasourceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.log.Errorf("Failed to create create datasource request: %v", err)
		return nil, fmt.Errorf("failed to create create datasource request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("X-Xsrf-Token", c.ExtractCookieValue(cookie, "XSRF-TOKEN"))

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Errorf("Failed to send create datasource request: %v", err)
		return nil, fmt.Errorf("failed to send create datasource request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Errorf("Failed to read create datasource response: %v", err)
		return nil, fmt.Errorf("failed to read create datasource response: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		c.log.Errorf("Create datasource failed with status code: %d, response: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("create datasource failed with status code: %d", resp.StatusCode)
	}

	// 解析响应
	var createDatasourceResp CreateDatasourceResponse
	if err := json.Unmarshal(body, &createDatasourceResp); err != nil {
		c.log.Errorf("Failed to parse create datasource response: %v", err)
		return nil, fmt.Errorf("failed to parse create datasource response: %v", err)
	}

	// 检查业务状态
	if !createDatasourceResp.Successful {
		errorMsg := "create datasource failed"
		if createDatasourceResp.Message != nil {
			errorMsg = *createDatasourceResp.Message
		}
		c.log.Errorf("Create datasource failed: %s", errorMsg)
		return nil, fmt.Errorf("create datasource failed: %s", errorMsg)
	}

	c.log.Infof("Successfully created datasource: %s (ID: %d)", createDatasourceResp.Data.Name, createDatasourceResp.Data.ID)
	return &createDatasourceResp, nil
}

// UpdateDatasource 修改数据源
func (c *SqlWorkbenchClient) UpdateDatasource(datasourceId int64, updateDatasource UpdateDatasourceRequest, publicKey string, cookie string, organizationID int64) (*CreateDatasourceResponse, error) {
	c.log.Infof("Attempting to update datasource with ID: %d", datasourceId)

	if updateDatasource.Password != nil {
		// 加密密码
		encryptedPassword, err := c.EncryptPasswordWithRSA(*updateDatasource.Password, publicKey)
		if err != nil {
			c.log.Errorf("Failed to encrypt password for datasource %s: %v", *updateDatasource.Name, err)
			return nil, fmt.Errorf("failed to encrypt password for datasource %s: %v", *updateDatasource.Name, err)
		}
		updateDatasource.Password = &encryptedPassword
	}

	// 构建修改数据源URL
	updateDatasourceURL := fmt.Sprintf("%s/api/v2/datasource/datasources/%d?currentOrganizationId=%d", c.baseURL, datasourceId, organizationID)

	// 准备JSON请求体
	jsonData, err := json.Marshal(updateDatasource)
	if err != nil {
		c.log.Errorf("Failed to marshal update datasource data: %v", err)
		return nil, fmt.Errorf("failed to marshal update datasource data: %v", err)
	}

	// 创建PUT请求
	req, err := http.NewRequest("PUT", updateDatasourceURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.log.Errorf("Failed to create update datasource request: %v", err)
		return nil, fmt.Errorf("failed to create update datasource request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("X-Xsrf-Token", c.ExtractCookieValue(cookie, "XSRF-TOKEN"))

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Errorf("Failed to send update datasource request: %v", err)
		return nil, fmt.Errorf("failed to send update datasource request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Errorf("Failed to read update datasource response: %v", err)
		return nil, fmt.Errorf("failed to read update datasource response: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		c.log.Errorf("Update datasource failed with status code: %d, response: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("update datasource failed with status code: %d", resp.StatusCode)
	}

	// 解析响应
	var updateDatasourceResp CreateDatasourceResponse
	if err := json.Unmarshal(body, &updateDatasourceResp); err != nil {
		c.log.Errorf("Failed to parse update datasource response: %v", err)
		return nil, fmt.Errorf("failed to parse update datasource response: %v", err)
	}

	// 检查业务状态
	if !updateDatasourceResp.Successful {
		errorMsg := "update datasource failed"
		if updateDatasourceResp.Message != nil {
			errorMsg = *updateDatasourceResp.Message
		}
		c.log.Errorf("Update datasource failed: %s", errorMsg)
		return nil, fmt.Errorf("update datasource failed: %s", errorMsg)
	}

	c.log.Infof("Successfully updated datasource: %s (ID: %d)", updateDatasourceResp.Data.Name, updateDatasourceResp.Data.ID)
	return &updateDatasourceResp, nil
}

// DeleteDatasource 删除数据源
func (c *SqlWorkbenchClient) DeleteDatasource(id int64, cookie string, organizationID int64) (*DeleteDatasourceResponse, error) {
	c.log.Infof("Attempting to delete datasource with ID: %d", id)

	// 构建删除数据源URL
	deleteDatasourceURL := fmt.Sprintf("%s/api/v2/datasource/datasources/%d?currentOrganizationId=%d", c.baseURL, id, organizationID)

	// 创建DELETE请求
	req, err := http.NewRequest("DELETE", deleteDatasourceURL, nil)
	if err != nil {
		c.log.Errorf("Failed to create delete datasource request: %v", err)
		return nil, fmt.Errorf("failed to create delete datasource request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("X-Xsrf-Token", c.ExtractCookieValue(cookie, "XSRF-TOKEN"))

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Errorf("Failed to send delete datasource request: %v", err)
		return nil, fmt.Errorf("failed to send delete datasource request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Errorf("Failed to read delete datasource response: %v", err)
		return nil, fmt.Errorf("failed to read delete datasource response: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		c.log.Errorf("Delete datasource failed with status code: %d, response: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("delete datasource failed with status code: %d", resp.StatusCode)
	}

	// 解析响应
	var deleteDatasourceResp DeleteDatasourceResponse
	if err := json.Unmarshal(body, &deleteDatasourceResp); err != nil {
		c.log.Errorf("Failed to parse delete datasource response: %v", err)
		return nil, fmt.Errorf("failed to parse delete datasource response: %v", err)
	}

	// 检查业务状态
	if !deleteDatasourceResp.Successful {
		errorMsg := "delete datasource failed"
		if deleteDatasourceResp.Message != nil {
			errorMsg = *deleteDatasourceResp.Message
		}
		c.log.Errorf("Delete datasource failed: %s", errorMsg)
		return nil, fmt.Errorf("delete datasource failed: %s", errorMsg)
	}

	c.log.Infof("Successfully deleted datasource: %s (ID: %d)", deleteDatasourceResp.Data.Name, deleteDatasourceResp.Data.ID)
	return &deleteDatasourceResp, nil
}

// CreateUsers 创建用户
func (c *SqlWorkbenchClient) CreateUsers(users []CreateUserRequest, publicKey string, cookie string) (*CreateUsersResponse, error) {
	c.log.Infof("Attempting to create %d users", len(users))

	// 加密所有用户的密码
	for i := range users {
		encryptedPassword, err := c.EncryptPasswordWithRSA(users[i].Password, publicKey)
		if err != nil {
			c.log.Errorf("Failed to encrypt password for user %s: %v", users[i].AccountName, err)
			return nil, fmt.Errorf("failed to encrypt password for user %s: %v", users[i].AccountName, err)
		}
		users[i].Password = encryptedPassword
	}

	// 构建创建用户URL
	createUsersURL := fmt.Sprintf("%s/api/v2/iam/users/batchCreate?currentOrganizationId=1", c.baseURL)

	// 准备JSON请求体
	jsonData, err := json.Marshal(users)
	if err != nil {
		c.log.Errorf("Failed to marshal users data: %v", err)
		return nil, fmt.Errorf("failed to marshal users data: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", createUsersURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.log.Errorf("Failed to create create users request: %v", err)
		return nil, fmt.Errorf("failed to create create users request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("X-Xsrf-Token", c.ExtractCookieValue(cookie, "XSRF-TOKEN"))

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Errorf("Failed to send create users request: %v", err)
		return nil, fmt.Errorf("failed to send create users request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Errorf("Failed to read create users response: %v", err)
		return nil, fmt.Errorf("failed to read create users response: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		c.log.Errorf("Create users failed with status code: %d, response: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("create users failed with status code: %d", resp.StatusCode)
	}

	// 解析响应
	var createUsersResp CreateUsersResponse
	if err := json.Unmarshal(body, &createUsersResp); err != nil {
		c.log.Errorf("Failed to parse create users response: %v", err)
		return nil, fmt.Errorf("failed to parse create users response: %v", err)
	}

	// 检查业务状态
	if !createUsersResp.Successful {
		errorMsg := "create users failed"
		if createUsersResp.Message != nil {
			errorMsg = *createUsersResp.Message
		}
		c.log.Errorf("Create users failed: %s", errorMsg)
		return nil, fmt.Errorf("create users failed: %s", errorMsg)
	}

	c.log.Infof("Successfully created %d users", len(createUsersResp.Data.Contents))
	return &createUsersResp, nil
}

// ActivateUser 激活用户
func (c *SqlWorkbenchClient) ActivateUser(username, currentPassword, newPassword, publicKey, cookie string) (*ActivateUserResponse, error) {
	c.log.Infof("Attempting to activate user: %s", username)

	// 加密当前密码和新密码
	encryptedCurrentPassword, err := c.EncryptPasswordWithRSA(currentPassword, publicKey)
	if err != nil {
		c.log.Errorf("Failed to encrypt current password for user %s: %v", username, err)
		return nil, fmt.Errorf("failed to encrypt current password for user %s: %v", username, err)
	}

	encryptedNewPassword, err := c.EncryptPasswordWithRSA(newPassword, publicKey)
	if err != nil {
		c.log.Errorf("Failed to encrypt new password for user %s: %v", username, err)
		return nil, fmt.Errorf("failed to encrypt new password for user %s: %v", username, err)
	}

	// 构建激活用户URL
	activateUserURL := fmt.Sprintf("%s/api/v2/iam/users/%s/activate?currentOrganizationId=", c.baseURL, username)

	// 准备激活用户请求
	activateUserReq := ActivateUserRequest{
		Username:        username,
		CurrentPassword: encryptedCurrentPassword,
		NewPassword:     encryptedNewPassword,
	}

	// 准备JSON请求体
	jsonData, err := json.Marshal(activateUserReq)
	if err != nil {
		c.log.Errorf("Failed to marshal activate user data: %v", err)
		return nil, fmt.Errorf("failed to marshal activate user data: %v", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", activateUserURL, bytes.NewBuffer(jsonData))
	if err != nil {
		c.log.Errorf("Failed to create activate user request: %v", err)
		return nil, fmt.Errorf("failed to create activate user request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("X-Xsrf-Token", c.ExtractCookieValue(cookie, "XSRF-TOKEN"))

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Errorf("Failed to send activate user request: %v", err)
		return nil, fmt.Errorf("failed to send activate user request: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Errorf("Failed to read activate user response: %v", err)
		return nil, fmt.Errorf("failed to read activate user response: %v", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		c.log.Errorf("Activate user failed with status code: %d, response: %s", resp.StatusCode, string(body))
		return nil, fmt.Errorf("activate user failed with status code: %d", resp.StatusCode)
	}

	// 解析响应
	var activateUserResp ActivateUserResponse
	if err := json.Unmarshal(body, &activateUserResp); err != nil {
		c.log.Errorf("Failed to parse activate user response: %v", err)
		return nil, fmt.Errorf("failed to parse activate user response: %v", err)
	}

	// 检查业务状态
	if !activateUserResp.Successful {
		errorMsg := "activate user failed"
		if activateUserResp.Message != nil {
			errorMsg = *activateUserResp.Message
		}
		c.log.Errorf("Activate user failed: %s", errorMsg)
		return nil, fmt.Errorf("activate user failed: %s", errorMsg)
	}

	c.log.Infof("Successfully activated user: %s (ID: %d)", activateUserResp.Data.Name, activateUserResp.Data.ID)
	return &activateUserResp, nil
}

// EncryptPasswordWithRSA 使用RSA公钥加密密码（兼容JSEncrypt）
func (c *SqlWorkbenchClient) EncryptPasswordWithRSA(password, publicKey string) (string, error) {
	// 解码Base64格式的公钥
	keyBytes, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 public key: %v", err)
	}

	// 解析公钥
	pub, err := x509.ParsePKIXPublicKey(keyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %v", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("not an RSA public key")
	}

	// 使用PKCS#1 v1.5填充加密（兼容JSEncrypt）
	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, []byte(password))
	if err != nil {
		return "", fmt.Errorf("failed to encrypt password: %v", err)
	}

	// 返回Base64编码的加密结果
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// ExtractCookieValue 从Cookie字符串中提取指定名称的值
func (c *SqlWorkbenchClient) ExtractCookieValue(cookieStr, cookieName string) string {
	if cookieStr == "" {
		return ""
	}

	// 使用正则表达式提取Cookie值
	pattern := fmt.Sprintf(`%s=([^;]+)`, regexp.QuoteMeta(cookieName))
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(cookieStr)

	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// MergeCookies 合并XSRF-TOKEN和JSESSIONID
func (c *SqlWorkbenchClient) MergeCookies(xsrfTokenStr, jsessionIdStr string) string {
	xsrfToken := c.ExtractCookieValue(xsrfTokenStr, "XSRF-TOKEN")
	jsessionId := c.ExtractCookieValue(jsessionIdStr, "JSESSIONID")

	if xsrfToken == "" && jsessionId == "" {
		return ""
	}

	var cookies []string
	if jsessionId != "" {
		cookies = append(cookies, fmt.Sprintf("JSESSIONID=%s", jsessionId))
	}
	if xsrfToken != "" {
		cookies = append(cookies, fmt.Sprintf("XSRF-TOKEN=%s", xsrfToken))
	}

	return strings.Join(cookies, "; ")
}

// LoginRequest 登录请求结构
type LoginRequest struct {
	Username string `form:"username"`
	Password string `form:"password"`
}

// LoginResponse 登录响应结构
type LoginResponse struct {
	Data           string  `json:"data"`
	DurationMillis *int64  `json:"durationMillis"`
	HTTPStatus     string  `json:"httpStatus"`
	RequestID      *string `json:"requestId"`
	Server         *string `json:"server"`
	Successful     bool    `json:"successful"`
	Timestamp      *int64  `json:"timestamp"`
	TraceID        *string `json:"traceId"`
	// 错误相关字段
	Code    *string     `json:"code,omitempty"`
	Message *string     `json:"message,omitempty"`
	Cookie  string      `json:"cookie"`
	Error   interface{} `json:"error"`
}

// CreateUserRequest 创建用户请求结构
type CreateUserRequest struct {
	AccountName string  `json:"accountName"`
	Name        string  `json:"name"`
	Password    string  `json:"password"`
	Enabled     bool    `json:"enabled"`
	RoleIDs     []int64 `json:"roleIds"`
}

// Address 地址结构
type Address struct {
	Country       *string `json:"country"`
	Formatted     *string `json:"formatted"`
	Locality      *string `json:"locality"`
	PostalCode    *string `json:"postalCode"`
	Region        *string `json:"region"`
	StreetAddress *string `json:"streetAddress"`
}

// User 用户结构
type User struct {
	AccessTokenHash               *string                `json:"accessTokenHash"`
	AccountName                   string                 `json:"accountName"`
	AccountNonExpired             bool                   `json:"accountNonExpired"`
	AccountNonLocked              bool                   `json:"accountNonLocked"`
	Active                        bool                   `json:"active"`
	Address                       Address                `json:"address"`
	Attributes                    interface{}            `json:"attributes"`
	Audience                      *string                `json:"audience"`
	AuthenticatedAt               *int64                 `json:"authenticatedAt"`
	AuthenticationContextClass    *string                `json:"authenticationContextClass"`
	AuthenticationMethods         interface{}            `json:"authenticationMethods"`
	Authorities                   interface{}            `json:"authorities"`
	AuthorizationCodeHash         *string                `json:"authorizationCodeHash"`
	AuthorizedActions             interface{}            `json:"authorizedActions"`
	AuthorizedParty               *string                `json:"authorizedParty"`
	Birthdate                     *string                `json:"birthdate"`
	BuiltIn                       bool                   `json:"builtIn"`
	Claims                        map[string]interface{} `json:"claims"`
	CreateTime                    int64                  `json:"createTime"`
	CreatorID                     int64                  `json:"creatorId"`
	CreatorName                   *string                `json:"creatorName"`
	CredentialsNonExpired         bool                   `json:"credentialsNonExpired"`
	Description                   *string                `json:"description"`
	Email                         *string                `json:"email"`
	EmailVerified                 *bool                  `json:"emailVerified"`
	Enabled                       bool                   `json:"enabled"`
	ExpiresAt                     *int64                 `json:"expiresAt"`
	ExtraProperties               interface{}            `json:"extraProperties"`
	FamilyName                    *string                `json:"familyName"`
	FullName                      *string                `json:"fullName"`
	Gender                        *string                `json:"gender"`
	GivenName                     *string                `json:"givenName"`
	ID                            int64                  `json:"id"`
	IDToken                       *string                `json:"idToken"`
	IssuedAt                      *int64                 `json:"issuedAt"`
	Issuer                        *string                `json:"issuer"`
	LastLoginTime                 *int64                 `json:"lastLoginTime"`
	Locale                        *string                `json:"locale"`
	LoginTime                     *int64                 `json:"loginTime"`
	MiddleName                    *string                `json:"middleName"`
	Name                          string                 `json:"name"`
	NickName                      *string                `json:"nickName"`
	Nonce                         *string                `json:"nonce"`
	OrganizationID                int64                  `json:"organizationId"`
	OrganizationIDs               interface{}            `json:"organizationIds"`
	OrganizationType              *string                `json:"organizationType"`
	PhoneNumber                   *string                `json:"phoneNumber"`
	PhoneNumberVerified           *bool                  `json:"phoneNumberVerified"`
	Picture                       *string                `json:"picture"`
	PreferredUsername             *string                `json:"preferredUsername"`
	Profile                       *string                `json:"profile"`
	ResourceManagementPermissions interface{}            `json:"resourceManagementPermissions"`
	RoleIDs                       interface{}            `json:"roleIds"`
	Roles                         interface{}            `json:"roles"`
	Subject                       *string                `json:"subject"`
	SystemOperationPermissions    interface{}            `json:"systemOperationPermissions"`
	Type                          string                 `json:"type"`
	UpdateTime                    int64                  `json:"updateTime"`
	UpdatedAt                     *int64                 `json:"updatedAt"`
	UserInfo                      interface{}            `json:"userInfo"`
	Username                      string                 `json:"username"`
	Website                       *string                `json:"website"`
	ZoneInfo                      *string                `json:"zoneInfo"`
}

// CreateUsersResponse 创建用户响应结构
type CreateUsersResponse struct {
	Data struct {
		Contents []User `json:"contents"`
	} `json:"data"`
	DurationMillis int64   `json:"durationMillis"`
	HTTPStatus     string  `json:"httpStatus"`
	RequestID      string  `json:"requestId"`
	Server         string  `json:"server"`
	Successful     bool    `json:"successful"`
	Timestamp      float64 `json:"timestamp"`
	TraceID        string  `json:"traceId"`
	// 错误相关字段
	Code    *string     `json:"code,omitempty"`
	Message *string     `json:"message,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// Organization 组织结构
type Organization struct {
	Builtin          bool   `json:"builtin"`
	CreateTime       int64  `json:"createTime"`
	Description      string `json:"description"`
	DisplayName      string `json:"displayName"`
	ID               int64  `json:"id"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	UniqueIdentifier string `json:"uniqueIdentifier"`
	UpdateTime       int64  `json:"updateTime"`
}

// GetOrganizationsResponse 获取组织响应结构
type GetOrganizationsResponse struct {
	Data struct {
		Contents []Organization `json:"contents"`
	} `json:"data"`
	DurationMillis int64   `json:"durationMillis"`
	HTTPStatus     string  `json:"httpStatus"`
	RequestID      *string `json:"requestId"`
	Server         string  `json:"server"`
	Successful     bool    `json:"successful"`
	Timestamp      float64 `json:"timestamp"`
	TraceID        string  `json:"traceId"`
	XsrfToken      string  `json:"xsrfToken"`
	// 错误相关字段
	Code    *string     `json:"code,omitempty"`
	Message *string     `json:"message,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// SSLConfig SSL配置结构
type SSLConfig struct {
	Enabled            bool    `json:"enabled"`
	CACertObjectID     *string `json:"CACertObjectId,omitempty"`
	CacertObjectID     *string `json:"cacertObjectId,omitempty"`
	ClientCertObjectID *string `json:"clientCertObjectId,omitempty"`
	ClientKeyObjectID  *string `json:"clientKeyObjectId,omitempty"`
}

// CreateDatasourceRequest 创建数据源请求结构
type CreateDatasourceRequest struct {
	CreatorID         int64                  `json:"creatorId"`
	Type              string                 `json:"type"`
	Name              string                 `json:"name"`
	Username          string                 `json:"username"`
	Password          string                 `json:"password"`
	SysTenantUsername *string                `json:"sysTenantUsername"`
	ServiceName       *string                `json:"serviceName"`
	SSLConfig         SSLConfig              `json:"sslConfig"`
	SysTenantPassword *string                `json:"sysTenantPassword"`
	Properties        interface{}            `json:"properties"`
	EnvironmentID     int64                  `json:"environmentId"`
	JdbcURLParameters map[string]interface{} `json:"jdbcUrlParameters"`
	Host              string                 `json:"host"`
	Port              string                 `json:"port"`
}

// UpdateDatasourceRequest 创建数据源请求结构
type UpdateDatasourceRequest struct {
	Id                int64                   `json:"id"`
	CreatorID         *int64                  `json:"creatorId"`
	Type              string                  `json:"type"`
	Name              *string                 `json:"name"`
	Username          string                  `json:"username"`
	Password          *string                 `json:"password"`
	SysTenantUsername *string                 `json:"sysTenantUsername"`
	ServiceName       *string                 `json:"serviceName"`
	SSLConfig         SSLConfig               `json:"sslConfig"`
	SysTenantPassword *string                 `json:"sysTenantPassword"`
	Properties        *interface{}            `json:"properties"`
	EnvironmentID     int64                   `json:"environmentId"`
	JdbcURLParameters *map[string]interface{} `json:"jdbcUrlParameters"`
	Host              string                  `json:"host"`
	Port              string                  `json:"port"`
}

// DataSourceStatus 数据源状态结构
type DataSourceStatus struct {
	ErrorCode    *string `json:"errorCode"`
	ErrorMessage *string `json:"errorMessage"`
	Status       string  `json:"status"`
	Type         *string `json:"type"`
}

// DataSource 数据源结构
type DataSource struct {
	CatalogName          *string                `json:"catalogName"`
	CloudProvider        *string                `json:"cloudProvider"`
	ClusterName          *string                `json:"clusterName"`
	ConnectType          string                 `json:"connectType"`
	CreateTime           int64                  `json:"createTime"`
	CreatorID            int64                  `json:"creatorId"`
	CreatorName          *string                `json:"creatorName"`
	DbObjectLastSyncTime *int64                 `json:"dbObjectLastSyncTime"`
	DefaultSchema        string                 `json:"defaultSchema"`
	DialectType          string                 `json:"dialectType"`
	DssDataSourceID      *string                `json:"dssDataSourceId"`
	Enabled              bool                   `json:"enabled"`
	EnvironmentID        int64                  `json:"environmentId"`
	EnvironmentName      string                 `json:"environmentName"`
	EnvironmentStyle     string                 `json:"environmentStyle"`
	Host                 string                 `json:"host"`
	ID                   int64                  `json:"id"`
	InstanceNickName     *string                `json:"instanceNickName"`
	JdbcURLParameters    map[string]interface{} `json:"jdbcUrlParameters"`
	LabelIDs             interface{}            `json:"labelIds"`
	LastAccessTime       *int64                 `json:"lastAccessTime"`
	Name                 string                 `json:"name"`
	ObtenantName         *string                `json:"obtenantName"`
	OrganizationID       int64                  `json:"organizationId"`
	OwnerID              *string                `json:"ownerId"`
	PasswordSaved        bool                   `json:"passwordSaved"`
	PermittedActions     interface{}            `json:"permittedActions"`
	Port                 int64                  `json:"port"`
	ProjectID            *string                `json:"projectId"`
	ProjectName          *string                `json:"projectName"`
	Properties           map[string]interface{} `json:"properties"`
	QueryTimeoutSeconds  int64                  `json:"queryTimeoutSeconds"`
	ReadonlyUsername     *string                `json:"readonlyUsername"`
	Region               *string                `json:"region"`
	ServiceName          *string                `json:"serviceName"`
	SessionInitScript    *string                `json:"sessionInitScript"`
	SetTop               bool                   `json:"setTop"`
	Sid                  *string                `json:"sid"`
	SSLConfig            SSLConfig              `json:"sslConfig"`
	Status               DataSourceStatus       `json:"status"`
	SupportedOperations  interface{}            `json:"supportedOperations"`
	SysTenantUsername    *string                `json:"sysTenantUsername"`
	Temp                 bool                   `json:"temp"`
	TenantName           *string                `json:"tenantName"`
	TenantNickName       *string                `json:"tenantNickName"`
	Type                 string                 `json:"type"`
	UpdateTime           int64                  `json:"updateTime"`
	UserRole             *string                `json:"userRole"`
	Username             string                 `json:"username"`
	VisibleScope         string                 `json:"visibleScope"`
}

// CreateDatasourceResponse 创建数据源响应结构
type CreateDatasourceResponse struct {
	Data           DataSource `json:"data"`
	DurationMillis int64      `json:"durationMillis"`
	HTTPStatus     string     `json:"httpStatus"`
	RequestID      string     `json:"requestId"`
	Server         string     `json:"server"`
	Successful     bool       `json:"successful"`
	Timestamp      float64    `json:"timestamp"`
	TraceID        string     `json:"traceId"`
	// 错误相关字段
	Code    *string     `json:"code,omitempty"`
	Message *string     `json:"message,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// DeleteDatasourceResponse 删除数据源响应结构
type DeleteDatasourceResponse struct {
	Data           DataSource `json:"data"`
	DurationMillis int64      `json:"durationMillis"`
	HTTPStatus     string     `json:"httpStatus"`
	RequestID      string     `json:"requestId"`
	Server         string     `json:"server"`
	Successful     bool       `json:"successful"`
	Timestamp      float64    `json:"timestamp"`
	TraceID        string     `json:"traceId"`
	// 错误相关字段
	Code    *string     `json:"code,omitempty"`
	Message *string     `json:"message,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// GetPublicKeyResponse 获取公钥响应结构
type GetPublicKeyResponse struct {
	Data           string  `json:"data"`
	DurationMillis int64   `json:"durationMillis"`
	HTTPStatus     string  `json:"httpStatus"`
	RequestID      *string `json:"requestId"`
	Server         string  `json:"server"`
	Successful     bool    `json:"successful"`
	Timestamp      float64 `json:"timestamp"`
	TraceID        string  `json:"traceId"`
	// 错误相关字段
	Code    *string     `json:"code,omitempty"`
	Message *string     `json:"message,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// ActivateUserRequest 激活用户请求结构
type ActivateUserRequest struct {
	Username        string `json:"username"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ActivateUserResponse 激活用户响应结构
type ActivateUserResponse struct {
	Data           User    `json:"data"`
	DurationMillis int64   `json:"durationMillis"`
	HTTPStatus     string  `json:"httpStatus"`
	RequestID      string  `json:"requestId"`
	Server         string  `json:"server"`
	Successful     bool    `json:"successful"`
	Timestamp      float64 `json:"timestamp"`
	TraceID        string  `json:"traceId"`
	// 错误相关字段
	Code    *string     `json:"code,omitempty"`
	Message *string     `json:"message,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}
