package services

import (
	"encoding/base64"
	"net/http"
	"os"
	"testing"

	"github.com/didil/inhooks/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestMessageVerifier_Verify_OK(t *testing.T) {
	v := NewMessageVerifier()

	algorithm := models.HMACAlgorithmSHA256
	signatureHeader := "X-WEBHOOK-HMAC-256"
	currentSecretEnvVar := "FLOW_VERIF_CURRENT_SECRET"
	os.Setenv(currentSecretEnvVar, "ABC123456")

	flow := &models.Flow{
		Source: &models.Source{
			Verification: &models.Verification{
				VerificationType:    models.VerificationTypeHMAC,
				HMACAlgorithm:       &algorithm,
				SignatureHeader:     signatureHeader,
				CurrentSecretEnvVar: currentSecretEnvVar,
			},
		},
	}

	expectedSignature := "b5dbd4567522ac835856391a2f1aaf41a2ea64a5167cf1886cdc974f799f4976"

	headers := http.Header{}
	headers.Add(signatureHeader, string(expectedSignature))

	m := &models.Message{
		HttpHeaders: headers,
		Payload:     []byte("test-payload"),
	}

	err := v.Verify(flow, m)
	assert.NoError(t, err)
}

func TestMessageVerifier_Verify_PrevSecret_OK(t *testing.T) {
	v := NewMessageVerifier()

	algorithm := models.HMACAlgorithmSHA256
	signatureHeader := "X-WEBHOOK-HMAC-256"
	currentSecretEnvVar := "FLOW_VERIF_CURRENT_SECRET"
	previousSecretEnvVar := "FLOW_VERIF_PREVIOUS_SECRET"
	os.Setenv(currentSecretEnvVar, "ABC123")
	os.Setenv(previousSecretEnvVar, "XYZ789")

	flow := &models.Flow{
		Source: &models.Source{
			Verification: &models.Verification{
				VerificationType:     models.VerificationTypeHMAC,
				HMACAlgorithm:        &algorithm,
				SignatureHeader:      signatureHeader,
				CurrentSecretEnvVar:  currentSecretEnvVar,
				PreviousSecretEnvVar: previousSecretEnvVar,
			},
		},
	}

	expectedSignature := "90bd3ccfe176c2b38bfa32a9364c355e8d85773cabdc4e6cf007c7f4b885f34d"

	headers := http.Header{}
	headers.Add(signatureHeader, string(expectedSignature))

	m := &models.Message{
		HttpHeaders: headers,
		Payload:     []byte("test-payload"),
	}

	err := v.Verify(flow, m)
	assert.NoError(t, err)
}

func TestMessageVerifier_Verify_WithSignaturePrefix_OK(t *testing.T) {
	v := NewMessageVerifier()

	algorithm := models.HMACAlgorithmSHA256
	signaturePrefix := "sha256="
	signatureHeader := "X-WEBHOOK-HMAC-256"
	currentSecretEnvVar := "FLOW_VERIF_CURRENT_SECRET"
	os.Setenv(currentSecretEnvVar, "ABC123456")

	flow := &models.Flow{
		Source: &models.Source{
			Verification: &models.Verification{
				VerificationType:    models.VerificationTypeHMAC,
				HMACAlgorithm:       &algorithm,
				SignatureHeader:     signatureHeader,
				SignaturePrefix:     signaturePrefix,
				CurrentSecretEnvVar: currentSecretEnvVar,
			},
		},
	}

	expectedSignature := "b5dbd4567522ac835856391a2f1aaf41a2ea64a5167cf1886cdc974f799f4976"

	headers := http.Header{}
	headers.Add(signatureHeader, signaturePrefix+string(expectedSignature))

	m := &models.Message{
		HttpHeaders: headers,
		Payload:     []byte("test-payload"),
	}

	err := v.Verify(flow, m)
	assert.NoError(t, err)
}

func TestMessageVerifier_Verify_Failed(t *testing.T) {
	v := NewMessageVerifier()

	algorithm := models.HMACAlgorithmSHA256
	signatureHeader := "X-WEBHOOK-HMAC-256"
	currentSecretEnvVar := "FLOW_VERIF_CURRENT_SECRET"
	os.Setenv(currentSecretEnvVar, "ABC123456")

	flow := &models.Flow{
		Source: &models.Source{
			Verification: &models.Verification{
				VerificationType:    models.VerificationTypeHMAC,
				HMACAlgorithm:       &algorithm,
				SignatureHeader:     signatureHeader,
				CurrentSecretEnvVar: currentSecretEnvVar,
			},
		},
	}

	expectedSignature, err := base64.StdEncoding.DecodeString("4BPsV0ldPwWryx3oI/FzP2Uur6hfvhtzysQAqenkFj8=")
	assert.NoError(t, err)

	headers := http.Header{}
	headers.Add(signatureHeader, string(expectedSignature))

	m := &models.Message{
		HttpHeaders: headers,
		Payload:     []byte("test-payload"),
	}

	err = v.Verify(flow, m)
	assert.EqualError(t, err, "failed to verify message: invalid signature")
}
