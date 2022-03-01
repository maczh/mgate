package model

//单个API接口的文档
type ApiDocument map[string]ApiInfo

type ApiParameter struct {
	Type        string `json:"type" bson:"type"`
	Description string `json:"description" bson:"description"`
	Name        string `json:"name" bson:"name"`
	In          string `json:"in" bson:"in"`
	Required    bool   `json:"required" bson:"required"`
}

type ApiInfo struct {
	Description string         `json:"description" bson:"description"`
	Consumes    []string       `json:"consumes" bson:"consumes"`
	Produces    []string       `json:"produces" bson:"produces"`
	Tags        []string       `json:"tags" bson:"tags"`
	Summary     string         `json:"summary" bson:"summary"`
	Parameters  []ApiParameter `json:"parameters" bson:"parameters"`
	Responses   struct {
		Field1 struct {
			Description string `json:"description" bson:"description"`
			Schema      struct {
				Type string `json:"type" bson:"type"`
			} `json:"schema" bson:"schema"`
		} `json:"200" bson:"200"`
	} `json:"responses" bson:"responses"`
}

//完整的swagger文档
type SwaggerDocument struct {
	Schemes     []interface{} `json:"schemes" bson:"schemes"`
	Version     string        `json:"version" bson:"version"`
	Title       string        `json:"title" bson:"title"`
	Description string        `json:"description" bson:"description"`
	Swagger     string        `json:"swagger" bson:"swagger"`
	Info        struct {
		Description string `json:"description" bson:"description"`
		Title       string `json:"title" bson:"title"`
		Contact     struct {
		} `json:"contact" bson:"contact"`
		Version string `json:"version" bson:"version"`
	} `json:"info" bson:"info"`
	Host     string                 `json:"host" bson:"host"`
	BasePath string                 `json:"basePath" bson:"basePath"`
	Paths    map[string]ApiDocument `json:"paths" bson:"paths"`
}
