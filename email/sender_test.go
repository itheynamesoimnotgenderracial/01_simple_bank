package email

import (
	"testing"

	"github.com/projects/go/01_simple_bank/util"
	"github.com/stretchr/testify/require"
)

func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	config, err := util.LoadConfig("..")
	require.NoError(t, err)

	sender := NewGmailSend(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)
	subject := "A test email"
	content := `
	<h1>Hello world</h1>
	<p>This is a test message from <a href="google.com">Google</a>
	`
	to := []string{"cyrilelagumbay@gmail.com"}
	attachFiles := []string{"../main.go"}

	err = sender.SendEmail(subject, content, to, nil, nil, attachFiles)
	require.NoError(t, err)
}
