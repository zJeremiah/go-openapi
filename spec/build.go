package spec

import (
	"fmt"
	"reflect"

	jsoniter "github.com/json-iterator/go"
)

const Default = "default"

func New(title, version, description string) *OpenAPI {
	return &OpenAPI{
		Version: "3.0.3",
		Info: Info{
			Title:   title,
			Version: version,
			Desc:    description,
		},
		Tags:         make([]Tag, 0),
		Paths:        map[string]OperationMap{}, // a map of methods mapped to operations i.e., get, put, post, delete
		ExternalDocs: &ExternalDocs{},
	}
}

// key is the reference name for the open api spec
type Requests map[string]RequestBody
type Params map[string]Param

type RouteParam struct {
	Name     string // unique name reference
	Desc     string // A brief description of the parameter. This could contain examples of use. CommonMark syntax MAY be used for rich text representation.
	Required bool   // is this paramater required
	Location string // REQUIRED. The location of the parameter. Possible values are "query", "header", "path" or "cookie".
	Example  map[string]any
}

type Method string

const (
	GET     Method = "get"
	PUT     Method = "put"
	POST    Method = "post"
	DELETE  Method = "delete"
	OPTIONS Method = "options"
	HEAD    Method = "head"
	PATCH   Method = "patch"
	TRACE   Method = "trace"
)

type Type int
type Format int

const (
	Integer Type = iota + 1
	Number
	String
	Boolean
	Object
	Array
)

const (
	Int32 Format = iota + 1
	Int64
	Float
	Double
	Byte     // base64 encoded characters
	Binary   // any sequence of octets
	Date     // full-date - https://www.rfc-editor.org/rfc/rfc3339#section-5.6
	DateTime // date-time - https://www.rfc-editor.org/rfc/rfc3339#section-5.6
	Password
)

func (t Type) String() string {
	switch t {
	case Integer:
		return "integer"
	case Number:
		return "number"
	case String:
		return "string"
	case Boolean:
		return "boolean"
	case Object:
		return "object"
	case Array:
		return "array"
	}
	return ""
}

func (f Format) String() string {
	switch f {
	case Int32:
		return "int32"
	case Int64:
		return "int64"
	case Float:
		return "float"
	case Double:
		return "double"
	case Byte:
		return "byte"
	case Binary:
		return "binary"
	case Date:
		return "date"
	case DateTime:
		return "dateTime"
	case Password:
		return "password"
	}
	return ""
}

type MIMEType string
type Reference string

// common media types
const (
	Json    MIMEType = "application/json"
	Xml     MIMEType = "application/xml"
	Text    MIMEType = "text/plain"
	General MIMEType = "application/octet-stream"
	Html    MIMEType = "text/html"
	XForm   MIMEType = "application/x-www-form-urlencoded"
	Jscript MIMEType = "application/javascript"
	Form    MIMEType = "multipart/form-data"
)

type Route struct {
	Tag       string
	Desc      string
	Content   MIMEType
	ReqType   Type                  // the request type for the path i.e., array, object, string, integer
	RespType  Type                  // the response type for the path i.e., array, object, string, integer
	Responses map[string]RouteResp  // key references for responses
	Params    map[string]RouteParam // key reference for params
	Requests  map[string]RouteReq   // key reference for requests
}

type RouteResp struct {
	Code    string // response code (as a string) "200","400","302"
	Content MIMEType
	Ref     Reference // the reference name for the response object
	Array   bool      // is the response object an array
}

type RouteReq struct {
	Content MIMEType
	Ref     Reference
	Array   bool
}

type UniqueRoute struct {
	Path   string
	Method Method
}

type Tags []Tag

func (o *OpenAPI) AddTags(t Tags) {
	o.Tags = append(o.Tags, t...)
}

func (o *OpenAPI) AddTag(tag, description string) {
	o.Tags = append(o.Tags, Tag{
		Name: tag,
		Desc: description,
	})
}

// AddRoute will add a new route to the paths object for the openapi spec
func (o *OpenAPI) AddRoute(path, method, tag, desc, summary string) (ur UniqueRoute, err error) {
	if tag == "" {
		tag = Default
	}
	if path == "" || method == "" {
		return ur, fmt.Errorf("path and method cannot be an empty string")
	}

	ur = UniqueRoute{
		Path:   path,
		Method: Method(method),
	}

	// initialize the paths if nil
	if o.Paths == nil {
		o.Paths = make(Paths)
	}

	p, found := o.Paths[ur.Path]
	if !found {
		o.Paths[ur.Path] = make(OperationMap)
		p = o.Paths[ur.Path]
	}

	m := p[ur.Method]
	m.Desc = desc
	m.Tags = append(m.Tags, tag)
	m.OperationID = string(ur.Method) + "_" + ur.Path

	// save the given route in the spec object for later reference
	o.Routes[ur] = Route{
		Tag:  tag,
		Desc: desc,
	}

	p[ur.Method] = m
	o.Paths[ur.Path] = p

	return ur, nil
}

type BodyObject struct {
	MIMEType   MIMEType // the mimetype for the object
	HttpStatus string   // Any HTTP status code, '200', '201', '400' the value of 'default' can be used to cover all responses not defined
	Array      bool     // is the reference to an array
	Body       any      // the response object example used to determine the type and name of each field returned
	Desc       string   // description of the body
	Title      string   // object title
}

// NewBody is the data for mapping a request / response object to a specific route
func NewBody(mtype MIMEType, status, desc string, array bool, body any) BodyObject {
	return BodyObject{
		MIMEType:   mtype,
		HttpStatus: status,
		Array:      array,
		Body:       body,
		Desc:       desc,
	}
}

// AddParam adds a param object to the given unique route
func (o *OpenAPI) AddParam(ur UniqueRoute, rp RouteParam) error {
	if rp.Name == "" || rp.Location == "" {
		return fmt.Errorf("param name and location are required to add param")
	}
	p, found := o.Paths[ur.Path]
	if !found {
		return fmt.Errorf("could not find path to add param %v", ur)
	}
	m, found := p[ur.Method]
	if !found {
		return fmt.Errorf("could not find method to add param %v", ur)
	}

	m.Params = append(m.Params, Param{
		Name: rp.Name,
		Desc: rp.Desc,
		In:   rp.Location,
	})

	p[ur.Method] = m
	o.Paths[ur.Path] = p

	return nil
}

// AddResp adds response information to the api responses map
// this is used for a request body, response body
func (o *OpenAPI) AddResp(ur UniqueRoute, bo BodyObject) error {

	p, found := o.Paths[ur.Path]
	if !found {
		return fmt.Errorf("could not find path to add param %v", ur)
	}
	m, found := p[ur.Method]
	if !found {
		return fmt.Errorf("could not find method to add param %v", ur)
	}

	schema := Schema{
		Title: bo.Title,
		Desc:  bo.Desc,
	}

	t := reflect.TypeOf(bo.Body)
	k := t.Kind()
	switch k {
	case reflect.String:
		schema.Type = String.String()

	case reflect.Array, reflect.Slice:
		schema.Type = Array.String()
		schema.Items = &Schema{}
	}

	m.Responses = Responses{
		bo.HttpStatus: Response{
			Desc: bo.Desc,
			Content: map[string]Media{
				string(bo.MIMEType): Media{
					Schema: Schema{
						Title: bo.Title,
						Desc:  bo.Desc,
					},
				},
			},
		},
	}

	return nil
}

func (pr Properties) Construct(item any) {
	t := reflect.TypeOf(item)
	v := reflect.ValueOf(item)
	k := v.Kind()

	switch k {
	case reflect.Slice:
		t = reflect.SliceOf(t)
	case reflect.Array:

	}

}

// AddReq adds request information to the api requestBody object
func (o *OpenAPI) AddReqBody(ur UniqueRoute, bo BodyObject) error {
	return nil
}

// JSON returns the json string value for the OpenAPI object
func (o *OpenAPI) JSON() string {
	json := jsoniter.ConfigFastest
	b, _ := json.Marshal(o)
	return string(b)
}
