//go:build !dms

package service

import "github.com/labstack/echo/v4"

func (s *APIServer) initRouterDMS(v1 *echo.Group) error {
	return nil
}
