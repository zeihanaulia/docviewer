package main

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

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
	specURL := c.Params("*")
	resp, err := http.Get(specURL)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Failed to retrieve the specification")
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return c.Status(http.StatusInternalServerError).SendString("Failed to read the specification")
	}

	var jsonData []byte
	if strings.HasSuffix(specURL, ".yaml") || strings.HasSuffix(specURL, ".yml") {
		jsonData, err = convertYAMLToJSON(data)
		if err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Failed to convert YAML to JSON: " + err.Error())
		}
	} else {
		jsonData = data // Assume the data is already JSON
	}

	// Check AsyncAPI version if specType is "asyncapi"
	if specType == "asyncapi" {
		var spec map[string]interface{}
		if err := json.Unmarshal(jsonData, &spec); err != nil {
			return c.Status(http.StatusInternalServerError).SendString("Failed to parse JSON: " + err.Error())
		}
		version, ok := spec["asyncapi"].(string)
		if ok && strings.HasPrefix(version, "3.") {
			return t.Render(c, "unsupported_version.html", nil)
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
