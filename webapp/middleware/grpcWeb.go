package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
)

type GrpcWebMiddleware struct {
	*grpcweb.WrappedGrpcServer
}

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
func (m *GrpcWebMiddleware) Handler2() gin.HandlerFunc {

	return func(c *gin.Context) {
		if m.IsAcceptableGrpcCorsRequest(c.Request) || m.IsGrpcWebRequest(c.Request) {
			m.ServeHTTP(c.Writer, c.Request)
			return
		}

		c.Next()
	}
}

func (m *GrpcWebMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.IsAcceptableGrpcCorsRequest(r) || m.IsGrpcWebRequest(r) {
			m.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func NewGrpcWebMiddleware(grpcWeb *grpcweb.WrappedGrpcServer) *GrpcWebMiddleware {
	return &GrpcWebMiddleware{grpcWeb}
}
