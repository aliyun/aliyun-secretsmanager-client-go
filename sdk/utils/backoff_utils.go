package utils

import (
	sdkerr "github.com/aliyun/alibaba-cloud-sdk-go/sdk/errors"
	"strings"
)

const (
	// KMS限流返回错误码
	RejectedThrottling = "Rejected.Throttling"

	// KMS服务不可用返回错误码
	ServiceUnavailableTemporary = "ServiceUnavailableTemporary"

	// KMS服务内部错误返回错误码
	InternalFailure = "InternalFailure"
)

// 根据Client异常判断是否进行规避重试
func JudgeNeedBackoff(err error) bool {
	switch e := err.(type) {
	case sdkerr.Error:
		if RejectedThrottling == e.ErrorCode() || ServiceUnavailableTemporary == e.ErrorCode() || InternalFailure == e.ErrorCode() {
			return true
		}
	}
	return false
}

// 根据Client异常判断是否进行容灾重试
func JudgeNeedRecoveryException(err error) bool {
	switch e := err.(type) {
	case sdkerr.Error:
		if SdkReadTimeout == e.ErrorCode() || SdkServerUnreachable == e.ErrorCode() || SdkTimeoutError == e.ErrorCode() {
			return true
		}
	}
	return JudgeNeedBackoff(err)
}

func TransferErrorToClientError(err error) error {
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "i/o timeout") || strings.Contains(errStr, "Client.Timeout") || strings.Contains(errStr, "connection timed out") {
			return sdkerr.NewClientError(SdkTimeoutError, errStr, err)
		} else if strings.Contains(errStr, "unreachable host") || strings.Contains(errStr, "Bad Gateway") || strings.Contains(errStr, "no such host") ||
			strings.Contains(errStr, "connected host has failed") {
			return sdkerr.NewClientError(SdkServerUnreachable, errStr, err)
		}
	}
	return err
}
