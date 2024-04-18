package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strings"

	"io"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gopkg.in/yaml.v2"
)

//go:embed templates/*
var resources embed.FS

// Define a struct to hold our template
// Define a struct to hold our template
type Template struct {
	templates *template.Template
}

// Parse and load templates on initialization
func NewTemplate() *Template {
	// Parse templates from the embedded filesystem
	templates, err := template.ParseFS(resources, "templates/*")
	if err != nil {
		log.Fatalf("Error parsing templates: %v", err)
	}
	return &Template{
		templates: template.Must(templates, err),
	}
}

// Render template utility function
func (t *Template) Render(c *fiber.Ctx, name string, data interface{}) error {
	c.Type("html")
	return t.templates.ExecuteTemplate(c.Response().BodyWriter(), name, data)
}

func convertInterfaceMapToJSONMap(in interface{}) (interface{}, error) {
	switch x := in.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v := range x {
			ks, ok := k.(string)
			if !ok {
				return nil, errors.New("key is not a string")
			}
			var err error
			m[ks], err = convertInterfaceMapToJSONMap(v)
			if err != nil {
				return nil, err
			}
		}
		return m, nil
	case []interface{}:
		for i, v := range x {
			var err error
			x[i], err = convertInterfaceMapToJSONMap(v)
			if err != nil {
				return nil, err
			}
		}
	}
	return in, nil
}

func convertYAMLToJSON(data []byte) ([]byte, error) {
	var yamlObj interface{}
	if err := yaml.Unmarshal(data, &yamlObj); err != nil {
		return nil, err
	}

	jsonReadyObj, err := convertInterfaceMapToJSONMap(yamlObj)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.Marshal(jsonReadyObj)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func handleDocumentation(c *fiber.Ctx, t *Template, specType, templateName string) error {
	rawURL := c.Params("*")
	// Prepare the full URL including query parameters if they exist
	query := c.Request().URI().QueryArgs().String()
	if query != "" {
		rawURL += "?" + query
	}

	resp, err := http.Get(rawURL)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Failed to retrieve the specification")
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Failed to read the specification")
	}

	var jsonData []byte
	if strings.HasSuffix(rawURL, ".yaml") || strings.HasSuffix(rawURL, ".yml") {
		jsonData, err = convertYAMLToJSON(data)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Failed to convert YAML to JSON: " + err.Error())
		}
	} else {
		jsonData = data // Assume the data is already JSON
	}

	if specType == "openapi" {
		// Validate OpenAPI schema
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromData(jsonData)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Failed to load OpenAPI data: " + err.Error())
		}
		if err := doc.Validate(loader.Context); err != nil {
			return t.Render(c, "error_invalid_schema.html", nil)
		}
	}

	encodedData := base64.StdEncoding.EncodeToString(jsonData)
	return t.Render(c, templateName, map[string]interface{}{
		"EncodedSpecContent": encodedData,
	})
}

func main() {
	app := fiber.New()
	app.Use(logger.New())
	tmpl := NewTemplate()

	// Combined route for API documentation
	app.Get("/docs/:type/*", func(c *fiber.Ctx) error {
		specType := c.Params("type")
		switch specType {
		case "openapi", "asyncapi":
			templateName := specType + ".html" // assumes templates are named openapi.html and asyncapi.html
			return handleDocumentation(c, tmpl, specType, templateName)
		default:
			return c.Status(fiber.StatusBadRequest).SendString("Unsupported specification type")
		}
	})

	log.Fatal(app.Listen(":8080"))
}
