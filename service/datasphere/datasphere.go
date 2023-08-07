package datasphere

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/GermanVor/datasphere-cleaning-script/common"
	"github.com/GermanVor/datasphere-cleaning-script/service/operation"
)

// Dock API of Datasphere - https://cloud.yandex.com/en/docs/datasphere/api-ref/overview

const BASE_URL = "https://datasphere.api.cloud-preprod.yandex.net/datasphere/v2"

type Community struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type Project struct {
	Id          string `json:"id"`
	CommunityId string `json:"communityId"`
	Name        string `json:"name"`
}

type Client struct {
	httpClient      *http.Client
	operationClient *operation.Client

	AUTHORIZATION_TOKEN string
	ctx                 context.Context
}

func InitClient(ctx context.Context, authorizationToken string) *Client {
	httpClient := &http.Client{}
	operationClient := operation.InitClient(ctx, authorizationToken)

	return &Client{
		httpClient:          httpClient,
		operationClient:     operationClient,
		AUTHORIZATION_TOKEN: authorizationToken, ctx: ctx,
	}
}

func (c *Client) DoRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(
		method,
		BASE_URL+path,
		body,
	)
	common.Fatalln(err)

	req.Header.Add(
		"Authorization",
		c.AUTHORIZATION_TOKEN,
	)

	req = req.WithContext(c.ctx)

	return c.httpClient.Do(req)
}

func (c *Client) GetCommunities(organizationId, nameOrDescriptionPattern string) []Community {
	resp, err := c.DoRequest(
		"GET", "/communities",
		bytes.NewBuffer([]byte(fmt.Sprintf(
			`{"organizationId":"%s", "nameOrDescriptionPattern": "%s"}`,
			organizationId,
			nameOrDescriptionPattern,
		))),
	)

	common.Fatalln(err)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatalln(resp)
	}

	body, err := io.ReadAll(resp.Body)
	common.Fatalln(err)

	respBody := &struct {
		Communities []Community `json:"communities"`
	}{}
	err = json.Unmarshal(body, respBody)
	common.Fatalln(err)

	return respBody.Communities
}

func (c *Client) GetProjects(communityId string) ([]Project, error) {
	resp, err := c.DoRequest(
		"GET", "/projects",
		bytes.NewBuffer([]byte(fmt.Sprintf(`{"communityId":"%s"}`, communityId))),
	)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalln(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	respBody := &struct {
		Projects []Project `json:"projects"`
	}{}
	err = json.Unmarshal(body, respBody)

	return respBody.Projects, err
}

func (c *Client) DeleteCommunity(communityId string) error {
	resp, err := c.DoRequest("DELETE", fmt.Sprintf("/communities/%s", communityId), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	common.Fatalln(err)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s, %s", resp.Status, string(body))
	}

	respBody := &struct {
		Id string `json:"Id"`
	}{}
	err = json.Unmarshal(body, respBody)
	common.Fatalln(err)

	respCh, errCh := c.operationClient.PollOperation(respBody.Id, operation.POLL_COUNT_LIMIT)

	select {
	case <-respCh:
	case err = <-errCh:
	}

	return err
}

func (c *Client) DeleteProject(projectId string) error {
	resp, err := c.DoRequest("DELETE", fmt.Sprintf("/projects/%s", projectId), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	common.Fatalln(err)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s, %s", resp.Status, string(body))
	}

	respBody := &struct {
		Id string `json:"id"`
	}{}
	err = json.Unmarshal(body, respBody)
	common.Fatalln(err)

	respCh, errCh := c.operationClient.PollOperation(respBody.Id, operation.POLL_COUNT_LIMIT)

	select {
	case <-respCh:
	case err = <-errCh:
	}

	return err
}
