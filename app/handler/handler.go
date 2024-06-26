package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"p3/app/config"
	"p3/app/constant"
	"p3/app/model"
	"p3/app/service"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type HandlerInterface interface {
	HealthCheck(echo.Context) error
	BroadcastExtTxn(echo.Context) error
	PendingExtCheck(echo.Context) error
}

type handler struct {
	conf      *config.AppConfig
	service   service.ServiceInterface
	validator *validator.Validate
	logger    *zap.Logger
}

func NewHandler(conf *config.AppConfig, service service.ServiceInterface, logger *zap.Logger) HandlerInterface {
	validator := validator.New()
	return &handler{conf, service, validator, logger}
}

func (h *handler) HealthCheck(ec echo.Context) error {
	return ec.String(http.StatusOK, "Service is healthy!")
}

func (h *handler) BroadcastExtTxn(ec echo.Context) error {
	reqBody := new(model.RequestBodyBroadcastTxn)
	reqHeader := ec.Request().Header.Get(constant.XReqID)
	binder := echo.DefaultBinder{}

	ctx := context.WithValue(context.Background(), constant.XReqID, reqHeader)
	// bind request body
	err := binder.BindBody(ec, reqBody)
	if err != nil {
		h.logger.Error("Cannot bind request", zap.Error(err), zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
		return ec.JSON(http.StatusBadRequest, model.ResponseExternal{
			Message: "Cannot bind request",
		})
	}

	// validate request header
	if reqHeader == "" {
		h.logger.Info("Request header is invalid", zap.Error(err), zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
		return ec.JSON(http.StatusBadRequest, model.ResponseExternal{
			Message: "Request header is invalid, X-Request-ID is REQUIRED",
		})
	}

	// validate request body
	err = h.validator.Struct(reqBody)
	if err != nil {
		h.logger.Info("Request body is invalid", zap.Error(err), zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
		return ec.JSON(http.StatusBadRequest, model.ResponseExternal{
			Message: "Request body is invalid",
		})
	}

	// validate time
	err = timeFormatCheckUnix(reqBody.TimeStamp)
	if err != nil {
		h.logger.Info("timestamp is invalid", zap.Error(err), zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
		return ec.JSON(http.StatusBadRequest, model.ResponseExternal{
			Message: "timestamp is invalid",
		})

	}

	// send request body to service for business logic
	err, txStatus, txHash := h.service.BroadcastAndCheck(ctx, reqHeader, *reqBody)
	if err != nil {
		h.logger.Error("Cannot broadcast and check for transactions", zap.Error(err), zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
		return ec.JSON(http.StatusInternalServerError, model.ResponseExternal{
			Message: "Cannot broadcast and check for transactions",
			TXHash:  txHash,
		})
	}
	if txStatus == constant.Pending {

		return ec.JSON(http.StatusOK, model.ResponseExternal{
			Message: fmt.Sprintf("After retry %d times, status are still %s, please check via %s",
				h.conf.RetryForCheck.RetryTimes,
				txStatus,
				h.conf.ApiPath.PendingCheck),
			TXHash: txHash,
		})
	}
	return ec.JSON(http.StatusOK, model.ResponseExternal{
		Message:  "Status checking success",
		TXStatus: txStatus,
		TXHash:   txHash,
	})
}

func (h *handler) PendingExtCheck(ec echo.Context) error {
	reqBody := new(model.RequestBodyTxnStatusCheckExt)
	reqHeader := ec.Request().Header.Get(constant.XReqID)
	binder := echo.DefaultBinder{}

	ctx := context.WithValue(context.Background(), constant.XReqID, reqHeader)
	// bind request body
	err := binder.BindBody(ec, reqBody)
	if err != nil {
		h.logger.Error("Cannot bind request", zap.Error(err), zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
		return ec.JSON(http.StatusBadRequest, model.ResponseExternal{
			Message: "Cannot bind request",
		})
	}
	// validate request header
	if reqHeader == "" {
		h.logger.Info("Request header is invalid", zap.Error(err), zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
		return ec.JSON(http.StatusBadRequest, model.ResponseExternal{
			Message: "Request header is invalid, X-Request-ID is REQUIRED",
		})
	}

	// validate request body
	err = h.validator.Struct(reqBody)
	if err != nil {
		h.logger.Info("Request body is invalid", zap.Error(err), zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
		return ec.JSON(http.StatusBadRequest, model.ResponseExternal{
			Message: "Request body is invalid",
		})
	}

	// send request body to service for business logic
	err, txStatus := h.service.CheckStatus(ctx, reqBody.TXHash)
	if err != nil {
		h.logger.Error("Cannot check for transactions", zap.Error(err), zap.Any(constant.XReqID, ctx.Value(constant.XReqID)))
		return ec.JSON(http.StatusInternalServerError, model.ResponseExternal{
			Message: "Cannot check for transactions",
		})
	}

	return ec.JSON(http.StatusOK, model.ResponseExternal{
		Message:  "Status checking success",
		TXStatus: txStatus,
	})
}

func timeFormatCheckUnix(input uint64) error {
	strInput := strconv.FormatUint(input, 10)
	if len(strInput) < 10 || len(strInput) > 13 {
		return errors.New("input must be between 10 and 13 digits")
	}
	return nil
}
