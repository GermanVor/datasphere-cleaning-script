package operation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/GermanVor/datasphere-cleaning-script/common"
)

// Dock API of Operation Service - https://cloud.yandex.com/en/docs/api-design-guide/concepts/operation#monitoring

const BASE_URL = "https://operation.api.cloud-preprod.yandex.net/operations"
const POLL_COUNT_LIMIT = 15

type Error struct {
	Code    int
	Message string
}

type OperationResp struct {
	Done     bool        `json:"done"`
	Response interface{} `json:"response"`
	Error    *Error      `json:"error"`
}

type Client struct {
	http.Client

	AUTHORIZATION_TOKEN string
	ctx                 context.Context
}

func InitClient(ctx context.Context, authorizationToken string) *Client {
	client := http.Client{}

	return &Client{Client: client, AUTHORIZATION_TOKEN: authorizationToken, ctx: ctx}
}

func (c *Client) DoRequest(method, operationId string) (*http.Response, error) {
	req, err := http.NewRequest(
		method,
		fmt.Sprintf("%s/%s", BASE_URL, operationId),
		nil,
	)
	common.Fatalln(err)

	req.Header.Add(
		"Authorization",
		c.AUTHORIZATION_TOKEN,
	)

	req = req.WithContext(c.ctx)

	return c.Client.Do(req)
}

func (c *Client) callOperation(operationId string) (*OperationResp, error) {
	resp, err := c.DoRequest("GET", operationId)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	respBody := &OperationResp{}
	err = json.Unmarshal(body, respBody)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

var (
	ErrPollLimitCount = errors.New("polling limit exceeded")
	ErrPollResponse   = errors.New("polling limit exceeded")
)

type PollRespError struct {
	error
	Data *Error
}

func (c *Client) PollOperation(operationId string, POLL_COUNT_LIMIT int) (
	<-chan interface{},
	<-chan *PollRespError,
) {
	respCh := make(chan interface{})
	errCh := make(chan *PollRespError)

	pollTimer := time.NewTimer(3 * time.Second)

	go func() {
		defer pollTimer.Stop()
		defer close(respCh)
		defer close(errCh)

		pollCount := 0

		for {
			select {
			case <-pollTimer.C:
				pollCount++

				if pollCount == POLL_COUNT_LIMIT {
					err := fmt.Errorf("operationId %s, [%w]", operationId, ErrPollLimitCount)
					errCh <- &PollRespError{error: err}
					return
				}

				resp, err := c.callOperation(operationId)

				if err != nil {
					errCh <- &PollRespError{error: err}
					return
				}

				if resp.Done {
					if resp.Error != nil {
						errCh <- &PollRespError{error: ErrPollResponse, Data: resp.Error}
					}

					respCh <- resp.Response
					return
				}

				pollTimer.Reset(time.Second)
			case <-c.ctx.Done():
				return
			}
		}
	}()

	return respCh, errCh
}
