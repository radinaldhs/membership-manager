package notification

import (
	"fmt"
	"log/slog"
)

func logAttrsFailToSendMsg(msgModel NotificationMessageModel, token *FCMToken, err error) (string, []slog.Attr) {
	msg := fmt.Sprintf("fail to send notification: %s", err.Error())
	attrs := []slog.Attr{slog.Int("msg_id", msgModel.ID)}
	if token != nil {
		attrs = append(attrs, slog.Group("member",
			slog.Int("member_reg_branch_id", token.MemberRegistrationBranchID),
			slog.Int("member_id", token.MemberID),
		))

		attrs = append(attrs, slog.Group("device",
			slog.String("platform", token.Platform),
			slog.String("device_id", token.DeviceID),
		))
	}

	if msgModel.Topic != "" {
		attrs = append(attrs, slog.String("topic", msgModel.Topic))
	}

	return msg, attrs
}
