package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	db "github.com/projects/go/01_simple_bank/db/sqlc"
	"github.com/projects/go/01_simple_bank/util"
	"github.com/rs/zerolog/log"
)

const TaskSendVerifyEmail = "task:send_verify_email"

type PayloadSendVerifyEmail struct {
	Username string `json:"username"`
}

func (distrbutor *RedisTaskDistributor) DistributeTaskSendVerifyEmail(
	ctx context.Context,
	payload *PayloadSendVerifyEmail,
	opts ...asynq.Option,
) error {
	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %v", err)
	}

	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)

	info, err := distrbutor.client.EnqueueContext(ctx, task)

	if err != nil {
		return fmt.Errorf("failed to enque task: %v", err)
	}

	log.
		Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("queue", info.Queue).
		Int("max_retry", info.MaxRetry).
		Msg("enqueued task")

	return nil
}

func (processor *RedisProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	var payload PayloadSendVerifyEmail
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshall payload: %v", asynq.SkipRetry)
	}

	user, err := processor.store.GetUser(ctx, payload.Username)

	if err != nil {
		// if err == sql.ErrNoRows {
		// 	return fmt.Errorf("user doesnt exist: %v", asynq.SkipRetry)
		// }
		return fmt.Errorf("failed to get user: %v", err)
	}

	// TODO: send email to user
	verifyEmail, err := processor.store.CreateVerifyEmail(ctx, db.CreateVerifyEmailParams{
		Username:   user.Username,
		Email:      user.Email,
		SecretCode: util.RandomString(32),
	})

	if err != nil {
		return fmt.Errorf("failed to create verify email: %v", err)
	}

	subject := "Welcome to Simple Bank"
	verifyUrl := fmt.Sprintf("http://localhost:8080/v1/verify_email?email_id=%d&secret_code=%s", verifyEmail.ID, verifyEmail.SecretCode)
	content := fmt.Sprintf(`
	Hello %s, <br />
	Thank you for registering with us! <br />
	Please <a href="%s">click here</a> to verify your email address
	`, user.FullName, verifyUrl)
	to := []string{user.Email}

	err = processor.mailer.SendEmail(subject, content, to, nil, nil, nil)

	if err != nil {
		return fmt.Errorf("failed to send verify email: %v", err)
	}

	log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("email", user.Email).
		Msg("processed task")

	return nil
}
