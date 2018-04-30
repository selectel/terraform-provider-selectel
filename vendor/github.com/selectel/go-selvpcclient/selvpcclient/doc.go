/*
Package selvpcclient provides a library to work with the Selectel VPC API.

Authentication

To work with the Selectel VPC API you first need to:

	- create a Selectel account: https://my.selectel.ru/registration
	- obtain an API token: http://my.selectel.ru/profile/apikeys

You can then provide the API token to the selvpc service client.

Service clients

Service client is a special struct that implements a client for different part
of the Selectel VPC API.
You need to initialize the needed service client prior to do any requests:

	token := "token_string"
	resellClient := resell.NewV2ResellClient(token)

All methods of service clients uses the Go context to provide end-user of the
library with a native way to work with the cancellation signals
*/
package selvpcclient
