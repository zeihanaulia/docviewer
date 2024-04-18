API Doc Viewer
=============

Description
-----------

This project is a web application that provides an interface for rendering and viewing API documentation. It supports both OpenAPI and AsyncAPI specifications. The application checks the version of AsyncAPI documents and displays a message if the version is unsupported (versions 3.x and above). The web interface uses the AsyncAPI React Component to render the API documentation dynamically from provided specifications.

Features
--------

-   Support for OpenAPI and AsyncAPI specifications.
-   Dynamic rendering of API documentation.
-   Version checking for AsyncAPI to ensure compatibility.
-   User-friendly error handling for unsupported API specification versions.

Getting Started
---------------

### Prerequisites

Ensure you have the following installed:

-   [Go](https://golang.org/dl/) (Version 1.20 or later recommended)
-   Fiber web framework for Go

### Installation

1.  Clone the repository:

    ```
    git clone https://github.com/zeihanaulia/docviewer
    cd docviewer
    ```

2.  Install the required Go packages:

    ```
    go mod tidy
    ```

### Running the Application

To run the application, execute:

```
go run main.go
```

This starts the server on `http://localhost:8080`. You can access the API documentation interfaces by navigating to specific URLs (detailed in the Usage section below).

Usage
-----

The application provides the following endpoints:

### `/docs/:type/*`

-   **Description**: Serve API documentation based on the specification type.
-   **Method**: `GET`
-   **URL Path**: `/docs/:type/{spec-url}`
    -   `:type` - The type of API specification (`openapi` or `asyncapi`)
    -   `{spec-url}` - The URL encoded path to the API specification file
-   **Example**:
    -   OpenAPI: `http://localhost:8080/docs/openapi/https://example.com/path/to/openapi.yaml`
    -   AsyncAPI (Supported Versions): `http://localhost:8080/docs/asyncapi/https://example.com/path/to/asyncapi.yaml`

### Error Handling

-   **Unsupported AsyncAPI Versions**: Accessing an AsyncAPI document with a version starting with "3." will render a page from `unsupported_version.html`, informing the user that the version is not supported.

Dependencies
------------

-   **Fiber**: Web framework used for routing and handling HTTP requests.
-   **Go-YAML**: Used for YAML parsing.
-   **Template**: Standard library used for rendering HTML templates.

Contributing
------------

Contributions are what make the open-source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1.  Fork the Project
2.  Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3.  Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4.  Push to the Branch (`git push origin feature/AmazingFeature`)
5.  Open a Pull Request

License
-------

Distributed under the MIT License. See `LICENSE` for more information.

Contact
-------

Zeihan Aulia - zeihan.aulia@outlook.com

Project Link: <https://github.com/zeihanaulia/docviewer>
