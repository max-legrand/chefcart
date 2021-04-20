/*
Package middleware ...
	Used for integrating the GRPC server with the gin library
*/
// written by: Kevin Lin
// tested by: Shreyas Heragu
// debugged by: Milos Seskar
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
)

// GrpcWebMiddleware ...
type GrpcWebMiddleware struct {
	*grpcweb.WrappedGrpcServer
}

// GinGrpcWebMiddleware ...
func GinGrpcWebMiddleware(m *grpcweb.WrappedGrpcServer) gin.HandlerFunc {

	return func(context *gin.Context) {
		if m.IsAcceptableGrpcCorsRequest(context.Request) || m.IsGrpcWebRequest(context.Request) {
			context.Status(http.StatusOK) // <------------prevent 404 in gin
			m.ServeHTTP(context.Writer, context.Request)
		} else {
			context.Next()
		}
	}
}
