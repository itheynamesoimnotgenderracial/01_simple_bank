package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
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
		if err == sql.ErrNoRows {
			return fmt.Errorf("user doesnt exist: %v", asynq.SkipRetry)
		}
		return fmt.Errorf("failed to get user: %v", err)
	}

	// TODO: send email to user

	log.Info().
		Str("type", task.Type()).
		Bytes("payload", task.Payload()).
		Str("email", user.Email).
		Msg("processed task")

	return nil
}
