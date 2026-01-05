package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RouteGenerator generates route definitions
type RouteGenerator struct{}

// NewRouteGenerator creates a new route generator
func NewRouteGenerator() *RouteGenerator {
	return &RouteGenerator{}
}

// Route represents a route definition
type Route struct {
	Method      string
	Path        string
	Handler     string
	Middlewares []string
	Description string
}

// RouteGroup represents a group of routes with a common prefix
type RouteGroup struct {
	Prefix      string
	Routes      []Route
	Middlewares []string
}

// RouteConfig configures route generation
type RouteConfig struct {
	Package       string
	Resource      string
	Handler       string
	RESTful       bool
	Routes        []Route
	Groups        []RouteGroup
	EnableCORS    bool
	Documentation bool
}

// GenerateRoutes generates route file
func (g *RouteGenerator) GenerateRoutes(targetPath string, config RouteConfig) error {
	if config.Package == "" {
		config.Package = "routes"
	}

	var routes []Route

	// Generate RESTful routes if configured
	if config.RESTful && config.Resource != "" {
		routes = g.generateRESTfulRoutes(config.Resource, config.Handler)
	} else {
		routes = config.Routes
	}

	content := g.generateRouteContent(config, routes)

	routeDir := filepath.Join(targetPath, "internal", config.Package)
	if err := os.MkdirAll(routeDir, 0755); err != nil {
		return err
	}

	routePath := filepath.Join(routeDir, "routes.go")
	return os.WriteFile(routePath, []byte(content), 0644)
}

func (g *RouteGenerator) generateRESTfulRoutes(resource, handler string) []Route {
	return []Route{
		{
			Method:      "GET",
			Path:        "/" + resource,
			Handler:     handler + ".List" + strings.Title(resource),
			Description: "List all " + resource,
		},
		{
			Method:      "POST",
			Path:        "/" + resource,
			Handler:     handler + ".Create" + strings.Title(singular(resource)),
			Description: "Create a new " + singular(resource),
		},
		{
			Method:      "GET",
			Path:        "/" + resource + "/{id}",
			Handler:     handler + ".Get" + strings.Title(singular(resource)),
			Description: "Get a " + singular(resource) + " by ID",
		},
		{
			Method:      "PUT",
			Path:        "/" + resource + "/{id}",
			Handler:     handler + ".Update" + strings.Title(singular(resource)),
			Description: "Update a " + singular(resource),
		},
		{
			Method:      "DELETE",
			Path:        "/" + resource + "/{id}",
			Handler:     handler + ".Delete" + strings.Title(singular(resource)),
			Description: "Delete a " + singular(resource),
		},
	}
}

func (g *RouteGenerator) generateRouteContent(config RouteConfig, routes []Route) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`package %s

import (
	"net/http"
	
	"github.com/gorilla/mux"
`, config.Package))

	if config.EnableCORS {
		sb.WriteString("\t\"github.com/rs/cors\"\n")
	}

	sb.WriteString(")\n\n")

	// Generate SetupRoutes function
	sb.WriteString("// SetupRoutes configures all application routes\n")
	sb.WriteString("func SetupRoutes() *mux.Router {\n")
	sb.WriteString("\trouter := mux.NewRouter()\n\n")

	// Add CORS if enabled
	if config.EnableCORS {
		sb.WriteString("\t// Configure CORS\n")
		sb.WriteString("\tc := cors.New(cors.Options{\n")
		sb.WriteString("\t\tAllowedOrigins: []string{\"*\"},\n")
		sb.WriteString("\t\tAllowedMethods: []string{\"GET\", \"POST\", \"PUT\", \"DELETE\", \"OPTIONS\"},\n")
		sb.WriteString("\t\tAllowedHeaders: []string{\"*\"},\n")
		sb.WriteString("\t})\n")
		sb.WriteString("\trouter.Use(c.Handler)\n\n")
	}

	// Generate routes
	if len(config.Groups) > 0 {
		for _, group := range config.Groups {
			sb.WriteString(g.generateGroupRoutes(group, config.Documentation))
		}
	} else {
		for _, route := range routes {
			sb.WriteString(g.generateRoute(route, config.Documentation, "\t"))
		}
	}

	sb.WriteString("\n\treturn router\n")
	sb.WriteString("}\n")

	return sb.String()
}

func (g *RouteGenerator) generateGroupRoutes(group RouteGroup, withDoc bool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("\t// %s routes\n", group.Prefix))
	sb.WriteString(fmt.Sprintf("\t%sRouter := router.PathPrefix(\"%s\").Subrouter()\n",
		sanitizePrefix(group.Prefix), group.Prefix))

	// Add group middlewares
	for _, mw := range group.Middlewares {
		sb.WriteString(fmt.Sprintf("\t%sRouter.Use(%s)\n", sanitizePrefix(group.Prefix), mw))
	}

	sb.WriteString("\n")

	// Add routes to group
	for _, route := range group.Routes {
		routeLine := g.generateRoute(route, withDoc, "\t")
		routeLine = strings.ReplaceAll(routeLine, "router.", sanitizePrefix(group.Prefix)+"Router.")
		sb.WriteString(routeLine)
	}

	sb.WriteString("\n")
	return sb.String()
}

func (g *RouteGenerator) generateRoute(route Route, withDoc bool, indent string) string {
	var sb strings.Builder

	if withDoc && route.Description != "" {
		sb.WriteString(fmt.Sprintf("%s// %s\n", indent, route.Description))
	}

	handler := route.Handler

	// Wrap with middlewares if any
	if len(route.Middlewares) > 0 {
		for i := len(route.Middlewares) - 1; i >= 0; i-- {
			handler = route.Middlewares[i] + "(" + handler + ")"
		}
	}

	sb.WriteString(fmt.Sprintf("%srouter.HandleFunc(\"%s\", %s).Methods(\"%s\")\n",
		indent, route.Path, handler, route.Method))

	return sb.String()
}

// Helper functions

func sanitizePrefix(prefix string) string {
	// Remove leading/trailing slashes and replace remaining slashes with underscores
	sanitized := strings.Trim(prefix, "/")
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	return sanitized
}

func singular(word string) string {
	// Simple singularization - remove trailing 's'
	if strings.HasSuffix(word, "s") {
		return word[:len(word)-1]
	}
	return word
}
