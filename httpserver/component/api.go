package component

import "github.com/labstack/echo/v4"

func (com *httpserverComponent) Engine() *echo.Echo {
	return com.engine
}
