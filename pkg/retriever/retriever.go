package retriever

import (
	"encoding/base64"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	vault "github.com/hashicorp/vault/api"
	"github.com/sirupsen/logrus"
)

// GetParameterFromSSM retrieves the parameter from SSM
func GetParameterFromSSM(c ssmiface.SSMAPI, name string, encrypted bool, encoded bool, log *logrus.Logger) (string, error) {

	log.Infof("Retrieving parameter '%s'", name)

	input := &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(encrypted),
	}
	param, err := c.GetParameter(input)

	if err != nil {
		log.Errorf("Error retrieving parameter: %s", err.Error())
		return "", err
	}

	log.Info("Successfully retrieved parameter")

	if encoded {
		return decodeParameterValue(*param.Parameter.Value, log), nil
	}

	return *param.Parameter.Value, nil

}

func GetSecretFromVault(path string, encoded bool, log *logrus.Logger, c *vault.Logical) map[string]string {

	secret, err := c.Read(path)

	if secret == nil {
		log.Fatalf("Secret returned from vault was nil")
	}

	if err != nil {
		log.Fatalf("Error reading secret from Vault path '%s': %s", path, err.Error())
	} else if len(secret.Warnings) > 0 {
		log.Fatalf("Errors returned from Vault: %v", secret.Warnings)
	}

	m := make(map[string]string)

	rv := secret.Data["data"]

	b, err := json.Marshal(rv)

	if err != nil {
		panic(err)
	}

	json.Unmarshal(b, &m)

	if encoded {
		newMap := make(map[string]string)

		for k, v := range m {
			newMap[k] = decodeParameterValue(v, log)
		}

		return newMap
	}

	// We return the entire map
	return m
}

// decodeParameterValue returns a base64-decoded string
func decodeParameterValue(value string, log *logrus.Logger) string {
	data, err := base64.StdEncoding.DecodeString(value)

	if err != nil {
		log.Fatalf("Error decoding parameter store value: %s", err.Error())
	}

	log.Info("Successfully decoded secret")

	return string(data)
}
