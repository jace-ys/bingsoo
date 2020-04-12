package slack

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/slack-go/slack"
)

type Config struct {
	AccessToken   string
	SigningSecret string
}

type Handler struct {
	*slack.Client
	signingSecret string
}

func NewHandler(accessToken, signingSecret string) *Handler {
	return &Handler{
		Client:        slack.New(accessToken),
		signingSecret: signingSecret,
	}
}

func (h *Handler) VerifySignature(r *http.Request) error {
	err := h.verifySignature(r)
	if err != nil {
		return fmt.Errorf("failed to verify request signature: %w", err)
	}
	return nil
}

func (h *Handler) verifySignature(r *http.Request) error {
	verifier, err := slack.NewSecretsVerifier(r.Header, h.signingSecret)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	_, err = verifier.Write(body)
	if err != nil {
		return err
	}

	return verifier.Ensure()
}
