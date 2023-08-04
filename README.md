`.env` :

-   `AUTH_TOKEN` - a.k.`OAuth` from [here](https://cloud.yandex.ru/docs/iam/operations/iam-token/create)
-   `ORGANIZATION_ID` - as variant it is possible open `https://datasphere.yandex.com/` and take it from WebTool Network tab response of `listOrganizations` (field `id`) or take it via WebTool Console `YC.userSettings.orgId` (it is current user Organization id)
-   `COMMUNITY_SUBSTR` - better set this field like this `E2E-test` or something like that
