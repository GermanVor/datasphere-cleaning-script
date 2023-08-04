package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/GermanVor/datasphere-cleaning-script/common"
	"github.com/GermanVor/datasphere-cleaning-script/service/datasphere"
	"github.com/joho/godotenv"
)

// Take it from https://cloud.yandex.ru/docs/iam/operations/iam-token/create
var AUTH_TOKEN = ""
var ORGANIZATION_ID = ""
var AUTHORIZATION_TOKEN string
var COMMUNITY_SUBSTR = ""

func init() {
	godotenv.Load(".env")

	if authToken, ok := os.LookupEnv("AUTH_TOKEN"); ok {
		AUTH_TOKEN = authToken
	} else {
		log.Fatalln("there is no AUTH_TOKEN in env file.")
	}

	if organizationId, ok := os.LookupEnv("ORGANIZATION_ID"); ok {
		ORGANIZATION_ID = organizationId
	} else {
		log.Fatalln("there is no ORGANIZATION_ID in env file.")
	}

	if communitySubstr, ok := os.LookupEnv("COMMUNITY_SUBSTR"); ok {
		COMMUNITY_SUBSTR = communitySubstr
	} else {
		log.Fatalln("there is no ORGANIZATION_ID in env file.")
	}

	req, err := http.NewRequest(
		"POST",
		"https://iam.api.cloud-preprod.yandex.net/iam/v1/tokens",
		bytes.NewBuffer([]byte(fmt.Sprintf(`{"yandexPassportOauthToken":"%s"}`, AUTH_TOKEN))),
	)
	common.Fatalln(err)

	resp, err := http.DefaultClient.Do(req)
	common.Fatalln(err)

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalln(resp)
	}

	body, err := io.ReadAll(resp.Body)
	common.Fatalln(err)

	respBody := map[string]string{}
	err = json.Unmarshal(body, &respBody)
	common.Fatalln(err)

	if iamToken, ok := respBody["iamToken"]; ok {
		AUTHORIZATION_TOKEN = fmt.Sprintf("Bearer %s", iamToken)
	} else {
		log.Fatalln(respBody)
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	datasphereClient := datasphere.InitClient(ctx, AUTHORIZATION_TOKEN)
	communityArr := datasphereClient.GetCommunities(ORGANIZATION_ID, COMMUNITY_SUBSTR)

	for _, community := range communityArr {
		fmt.Println(community.Name)

		projectArr := datasphereClient.GetProjects(community.Id)
		for _, project := range projectArr {
			fmt.Println("\t", project.Name)
		}
	}
}
