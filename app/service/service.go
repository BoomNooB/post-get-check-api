package service

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"p3/app/config"
	"p3/app/constant"
	"p3/app/model"
	"time"

	"go.uber.org/zap"
)

type ServiceInterface interface {
	BroadcastAndCheck(ctx context.Context, reqHeader string, reqBody model.RequestBodyBroadcastTxn) (error, string)
	CheckStatus(ctx context.Context, txStatus string) (error, string)
}

type service struct {
	conf   *config.AppConfig
	logger *zap.Logger
}

func NewService(conf *config.AppConfig, logger *zap.Logger) ServiceInterface {
	return &service{conf, logger}
}

func (s *service) BroadcastAndCheck(ctx context.Context, reqHeader string, reqBody model.RequestBodyBroadcastTxn) (error, string) {
	httpClient := http.Client{
		Timeout: s.conf.HTTPClientTimeOut,
	}
	// broadcast for tx_hash
	err, txHash := s.broadcastForTxn(ctx, httpClient, reqBody)
	if err != nil {
		return err, ""
	}
	httpClient.CloseIdleConnections()

	/* --------------- */
	// periodically check status
	err, txnStatusString := s.checkTxStatus(ctx, httpClient, txHash)
	if err != nil {
		return err, ""
	}
	if txnStatusString == "PENDING" {
		retryCount := 1
		for range s.conf.RetryForCheck.RetryTimes {
			s.logger.Info("tx_status is pending, retry to check again",
				zap.String(constant.TXNHash, txHash),
				zap.Int(constant.RetryCount, retryCount),
				zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
			time.Sleep(s.conf.RetryForCheck.RetryDelay)
			err, txnStatusString := s.checkTxStatus(ctx, httpClient, txHash)
			if err != nil {
				return err, ""
			}
			if txnStatusString != constant.Pending {
				return nil, txnStatusString
			}
			retryCount++
		}
	}
	return nil, txnStatusString
}

func (s *service) CheckStatus(ctx context.Context, txHash string) (error, string) {
	// new http client
	httpClient := http.Client{
		Timeout: s.conf.HTTPClientTimeOut,
	}
	err, txnStatusString := s.checkTxStatus(ctx, httpClient, txHash)
	if err != nil {
		return err, ""
	}

	return nil, txnStatusString
}

func (s *service) checkTxStatus(ctx context.Context, httpClient http.Client, txHash string) (error, string) {
	httpClient = http.Client{
		Timeout: s.conf.HTTPClientTimeOut,
	}

	txnStatus := new(model.RequestTxnStatusCheck)
	getPathWithTxnHash := s.conf.ApiPath.CheckTxnPath + txHash
	respTxnStatus, err := httpClient.Get(getPathWithTxnHash)
	if err != nil {
		s.logger.Error("Error sending HTTP POST request to get txn hash", zap.Error(err), zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
		return err, ""
	}
	s.logger.Info("HTTP GET to check tx_status successfully", zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))

	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(respTxnStatus.Body).Decode(&txnStatus); err != nil {
		s.logger.Error("Error decoding response to json", zap.Error(err), zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
		return err, ""
	}
	s.logger.Info("The txn status decode successfully",
		zap.Any(constant.XReqID, ctx.Value(constant.XReqID)),
		zap.String(constant.TXNStatus, txnStatus.TXStatus))
	return nil, txnStatus.TXStatus
}

func (s *service) broadcastForTxn(ctx context.Context, httpClient http.Client, reqBody model.RequestBodyBroadcastTxn) (error, string) {
	txnHash := new(model.ResponseFromBroadcastTxn)
	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		s.logger.Error("Error marshalling request body to JSON", zap.Error(err), zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
		return err, ""
	}

	respTxnHash, err := httpClient.Post(s.conf.ApiPath.BroadCastTxnPath, constant.ContentTypeJSON, bytes.NewBuffer(reqBodyJson))
	if err != nil {
		s.logger.Error("Error sending HTTP POST request to get txn hash", zap.Error(err), zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
		return err, ""
	}
	s.logger.Info("HTTP POST to get tx_hash successfully", zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(respTxnHash.Body).Decode(&txnHash); err != nil {
		s.logger.Error("Error decoding response to json", zap.Error(err), zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
		return err, ""
	}
	s.logger.Info("The txn hash decode successfully", zap.String(constant.TXNHash, txnHash.TXHash), zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
	return nil, txnHash.TXHash
}
