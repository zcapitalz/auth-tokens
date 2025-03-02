package slogutils

import "log/slog"

func Error(msg string, err error, args ...any) {
	args = append(args, slog.Attr{
		Key:   "error",
		Value: slog.StringValue(err.Error()),
	})
	slog.Error(
		msg,
		args...,
	)
}
