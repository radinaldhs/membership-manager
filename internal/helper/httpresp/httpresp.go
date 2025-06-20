package httpresp

import (
	"log/slog"
	"net/http"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/internal/helper/customerror"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/errors"
	"gitlab.com/toko-mas-jawa/toko-mas-jawa-backend/pkg/logging"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type Error struct {
	Code    string       `json:"code,omitempty"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields,omitempty"`
}

type Response struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

func HandleError(ectx echo.Context, logger *logging.Logger, err error) error {
	xErr, ok := errors.GetError(err)
	if !ok {
		logger.Error(err.Error())
		return ectx.JSON(http.StatusInternalServerError, Response{
			Error: &Error{
				Message: "Internal server error",
			},
		})
	}

	switch xErr.Group() {
	case customerror.ErrGroupClientErr:
		return ectx.JSON(http.StatusBadRequest, Response{
			Message: "failed",
			Error: &Error{
				Code:    xErr.Code(),
				Message: xErr.Error(),
			},
		})

	case customerror.ErrGroupForbiddenErr:
		return ectx.JSON(http.StatusForbidden, Response{
			Message: "failed",
			Error: &Error{
				Code:    xErr.Code(),
				Message: xErr.Error(),
			},
		})

	case customerror.ErrGroupDataNotFoundErr:
		return ectx.JSON(http.StatusNotFound, Response{
			Message: "failed",
			Error: &Error{
				Code:    xErr.Code(),
				Message: xErr.Error(),
			},
		})

	case customerror.ErrGroupUnauthorized:
		return ectx.JSON(http.StatusUnauthorized, Response{
			Message: "failed",
			Error: &Error{
				Code:    xErr.Code(),
				Message: xErr.Error(),
			},
		})

	case customerror.ErrGroupForbidden:
		return ectx.JSON(http.StatusForbidden, Response{
			Message: "failed",
			Error: &Error{
				Code:    xErr.Code(),
				Message: xErr.Error(),
			},
		})

	case customerror.ErrGroupServiceUnavailable:
		return ectx.JSON(http.StatusServiceUnavailable, Response{
			Message: "failed",
			Error: &Error{
				Code:    xErr.Code(),
				Message: xErr.Error(),
			},
		})

	case customerror.ErrGroupInternalErr:
		var logAttrs []slog.Attr
		logAttrs = append(logAttrs, slog.Group("endpoint",
			slog.String("method", ectx.Request().Method),
			slog.String("pattern", ectx.Request().Pattern),
			slog.String("path", ectx.Request().URL.String()),
		))

		logAttrs = append(logAttrs, xErrToSlogAttr(xErr))

		logger.Error("Internal server error", logAttrs...)
		return ectx.JSON(http.StatusInternalServerError, Response{
			Message: "failed",
			Error: &Error{
				Message: "Internal server error",
			},
		})
	}

	return nil
}

func HandleValidationError(ectx echo.Context, logger *logging.Logger, validate *validator.Validate, trans ut.Translator, err error) error {
	validateErr, ok := err.(validator.ValidationErrors)
	if ok {
		var fieldErrs []FieldError
		for _, verr := range validateErr {
			fieldErrs = append(fieldErrs, FieldError{
				Field:   verr.Field(),
				Message: verr.Translate(trans),
			})
		}

		resp := Response{
			Message: "Tidak dapat memproses permintaan",
			Error: &Error{
				Message: "Format data tidak valid",
				Fields:  fieldErrs,
			},
		}

		return ectx.JSON(http.StatusBadRequest, resp)
	}

	logger.Error(err.Error())
	return ectx.JSON(http.StatusInternalServerError, Response{
		Error: &Error{
			Message: "Internal server error",
		},
	})
}

func xErrToSlogAttr(xErr errors.Error) slog.Attr {
	attr := slog.Group("error",
		slog.String("msg", xErr.Error()),
	)

	if source := xErr.Source(); source != nil {
		groupAttr := attr.Value.Group()
		groupAttr = append(groupAttr, slog.Group("source",
			slog.String("file", source.File),
			slog.Int("line", source.Line),
			slog.String("func_name", source.FuncName),
		))

		attr.Value = slog.GroupValue(groupAttr...)
	}

	if uErr := errors.Unwrap(xErr); uErr != nil {
		groupAttr := attr.Value.Group()
		groupAttr = append(groupAttr, slog.Group("error_cause", errToSlogAttr(uErr)))
		attr.Value = slog.GroupValue(groupAttr...)
	}

	return attr
}

func errToSlogAttr(err error) slog.Attr {
	xErr, ok := errors.GetError(err)
	if !ok {
		attr := slog.Group("error", slog.String("msg", err.Error()))
		if uErr := errors.Unwrap(err); uErr != nil {
			groupAttr := attr.Value.Group()
			groupAttr = append(groupAttr, slog.Group("error_cause", errToSlogAttr(uErr)))
			attr.Value = slog.GroupValue(groupAttr...)
		}

		return attr
	}

	return xErrToSlogAttr(xErr)
}

func BadRequestResponse(ectx echo.Context) error {
	res := Response{
		Message: "Tidak dapat memproses permintaan",
		Error: &Error{
			Message: "Format data tidak valid",
		},
	}

	return ectx.JSON(http.StatusBadRequest, res)
}

func OKResponse(ectx echo.Context, msg string, data any) error {
	res := Response{
		Message: msg,
		Data:    data,
	}

	return ectx.JSON(http.StatusOK, res)
}
