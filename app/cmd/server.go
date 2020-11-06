package cmd

import (
	"github.com/Semior001/decompract/app/rest/api"
)

// Server runs REST API web server
type Server struct {
	ServiceURL string `long:"service_url" env:"SERVICE_URL" description:"http service url" required:"true"`
	Port       int    `long:"service_port" env:"SERVICE_PORT" description:"http server port" default:"8080"`

	WebRoot string `long:"web-root" env:"WEB_ROOT" default:"./web" description:"web root directory"`

	CommonOpts
}

// Execute runs http web server
func (s *Server) Execute(_ []string) error {
	srv := api.Rest{Version: s.Version, WebRoot: s.WebRoot}
	srv.Run(s.Port)
	return nil
}
