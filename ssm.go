package scv

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"log"
	"os"
)

// PrivValKey is a simplified version of the tendermint private consensus key file.
type PrivValKey struct {
	Address string `json:"address"`
	PubKey  struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"pub_key"`
	PrivKey struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"priv_key"`
}

// MustGetKey fetches the private key from SSM, failure is fatal.
func MustGetKey() *PrivValKey {
	log.Println("attempting to fetch consensus key from AWS parameter store")
	if os.Getenv("AWS_REGION") == "" || os.Getenv("AWS_PARAMETER") == "" {
		log.Fatal("Expected the AWS_REGION and AWS_PARAMETER environment variables to be set")
	}
	awsSession := session.Must(
		session.NewSession(
			&aws.Config{
				Region: aws.String(os.Getenv("AWS_REGION")),
			},
		),
	)

	ps := ssm.New(awsSession)
	keyOut, err := ps.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String(os.Getenv("AWS_PARAMETER")),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		log.Fatal(err)
	}
	if keyOut == nil || len(*keyOut.Parameter.Value) == 0 {
		log.Fatal("got an empty key from SSM")
	}
	pk := &PrivValKey{}
	err = json.Unmarshal([]byte(*keyOut.Parameter.Value), pk)
	if pk.PrivKey.Value == "" {
		log.Fatal("invalid private key")
	}

	log.Printf("retrieved key with public key: %s from parameter store", pk.PubKey.Value)
	return pk
}
