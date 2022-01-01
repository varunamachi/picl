package cmn

import (
	"fmt"
	"io"

	"github.com/labstack/echo/v4"
)

type noopWriter struct{}

func (nw *noopWriter) Write(b []byte) (int, error) {
	return 0, nil
}

type Endpoint struct {
	Method    string
	Path      string
	Category  string
	Desc      string
	Version   string
	NeedsAuth bool
	Route     *echo.Route
	Handler   echo.HandlerFunc
}

type Server struct {
	categories map[string][]*Endpoint
	endpoints  []*Endpoint
	// routes map[string]
	root    *echo.Echo
	printer io.Writer
}

func NewServer(printer io.Writer) *Server {
	if printer == nil {
		printer = &noopWriter{}
	}
	return &Server{
		categories: make(map[string][]*Endpoint),
		endpoints:  make([]*Endpoint, 0, 100),
		root:       echo.New(),
		printer:    printer,
	}
}

func (s *Server) AddEndpoints(ep ...*Endpoint) *Server {
	s.endpoints = append(s.endpoints, ep...)
	return s
}

func (s *Server) Start(port uint32) error {
	s.configure()
	s.Print()
	return s.root.Start(fmt.Sprintf(":%d", port))
}

func (s *Server) configure() {

	sessionMw := func(hf echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// TODO: Check the session information, like JWT token etc
			return hf(c)
		}
	}

	type groupPair struct {
		versionGrp *echo.Group
		inGrp      *echo.Group
	}
	groups := map[string]*groupPair{}

	for _, ep := range s.endpoints {
		ep := ep
		epMiddleware := func(hf echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				c.Set("endpoint", ep)
				fmt.Fprintf(s.printer,
					"ACCESS: %-3s %-5s %-40s %s\n",
					ep.Version, ep.Route.Method, ep.Route.Path, ep.Category)

				err := hf(c)
				if err != nil {
					fmt.Fprintln(s.printer, err.Error())
				}
				return err
			}
		}

		grp := groups[ep.Version]
		if grp == nil {
			grp = &groupPair{}
			grp.versionGrp = s.root.Group("api/" + ep.Version)
			grp.inGrp = grp.versionGrp.Group("in")
			grp.inGrp.Use(sessionMw)
		}

		if ep.NeedsAuth {
			ep.Route = grp.inGrp.Add(
				ep.Method, ep.Path, ep.Handler, epMiddleware)

		} else {
			ep.Route = grp.versionGrp.Add(
				ep.Method, ep.Path, ep.Handler, epMiddleware)
		}

		if _, found := s.categories[ep.Category]; !found {
			s.categories[ep.Category] = make([]*Endpoint, 0, 20)
		}
		s.categories[ep.Category] = append(s.categories[ep.Category], ep)
	}
}

func (s *Server) Print() {
	for _, ep := range s.endpoints {
		cat := ep.Category
		if len(cat) > 14 {
			cat = ep.Category[:14]
		}
		fmt.Fprintf(s.printer,
			"%-3s %-5s %-40s %-15s %s\n",
			ep.Version, ep.Route.Method, ep.Route.Path, cat, ep.Desc)
	}
	fmt.Print("\n\n")
}
