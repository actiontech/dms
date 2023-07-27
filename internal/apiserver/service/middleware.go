package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	pkgConst "github.com/actiontech/dms/internal/apiserver/pkg/constant"
	pkgApiError "github.com/actiontech/dms/internal/apiserver/pkg/error"
	pkgQueue "github.com/actiontech/dms/pkg/process_queue"

	"github.com/actiontech/dms/pkg/dms-common/api/jwt"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/labstack/echo/v4"
)

var controllerProcessRecordsMaker sync.Once
var controllerProcessRecordsInstance *pkgQueue.ProcessRecordQueue

type ResponseBodyWriter struct {
	io.Writer
	http.ResponseWriter
	http.Flusher
}

func (w *ResponseBodyWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func (w *ResponseBodyWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *ResponseBodyWriter) Flush() {
	if responseFlush, ok := w.ResponseWriter.(http.Flusher); ok {
		responseFlush.Flush()
	}
}

// echo middleware for process record
func ProcessRecordMiddleware(logger log.Logger) echo.MiddlewareFunc {
	log := log.NewHelper(log.With(logger, "middleware", "process_record"))
	controllerProcessRecordsMaker.Do(func() {
		controllerProcessRecordsInstance = pkgQueue.NewProcessRecordQueue(logger, pkgConst.ProcessRecordFile)
	})
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			req := c.Request()
			if req.Method == "GET" {
				return next(c)
			}

			userUid, err := jwt.GetUserUidStrFromContext(c)
			if err != nil {
				log.Debugf("get user from context error: %v, skip user uid in record", err)
			} else {
				log.Debugf("get user from context: %v", userUid)
			}

			p := controllerProcessRecordsInstance.Push(userUid, req.RequestURI)

			resBody := new(bytes.Buffer)
			mw := io.MultiWriter(c.Response().Writer, resBody)
			writer := &ResponseBodyWriter{
				Writer:         mw,
				ResponseWriter: c.Response().Writer,
			}

			c.Response().Writer = writer

			if err = next(c); err != nil {
				c.Error(err)
			}

			responseBody := resBody.Bytes()

			var response map[string]interface{}
			errorCode := fmt.Sprintf("%v", pkgApiError.Unknown)
			errorMessage := ""
			httpCode := c.Response().Status

			if err := json.Unmarshal(responseBody, &response); err == nil {
				if val, ok := response["code"]; ok {
					errorCode = fmt.Sprintf("%v", val)
				}
				errorMessage = fmt.Sprintf("%v", response["msg"])
			} else {
				errorMessage = fmt.Sprintf("failed to unmarshal response body: %v", err)
			}

			// update and persistent operation log
			{
				p.UpdateRecord(httpCode, errorCode, errorMessage)

				controllerProcessRecordsInstance.Save()
			}

			return
		}
	}
}
