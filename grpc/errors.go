package grpc

import (
	"errors"
	"log"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleGrpcError(err error, connection, call string) error {
	if err == nil {
		return nil
	}
	// if this is a not implemented error we silently swallow it
	status, ok := status.FromError(err)
	if !ok {
		return err
	}

	// ignore unimplemented error
	if status.Code() == codes.Unimplemented {
		log.Printf("[TRACE] connection '%s' returned 'Unimplemented' error for call '%s' - plugin version does not support this call", connection, call)
		return nil
	}

	return errors.New(status.Message())
}

func IsGRPCConnectivityError(err error) bool {
	return err != nil && (strings.Contains(err.Error(), "error reading from server: EOF") || strings.Contains(err.Error(), "transport: error while dialing:"))
}