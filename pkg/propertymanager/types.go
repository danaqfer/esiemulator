package propertymanager

import (
	"encoding/xml"
	"net/http"
	"time"
)

// Property represents an Akamai property configuration
type Property struct {
	XMLName   xml.Name  `xml:"property"`
	Name      string    `xml:"name,attr"`
	Version   int       `xml:"version,attr"`
	Rules     Rules     `xml:"rules"`
	Behaviors Behaviors `xml:"behaviors"`
	Variables Variables `xml:"variables"`
	Comments  string    `xml:"comments,omitempty"`
}

// Rules represents a collection of rules
type Rules struct {
	XMLName xml.Name `xml:"rules"`
	Rule    []Rule   `xml:"rule"`
}

// Rule represents a single rule with conditions and behaviors
type Rule struct {
	XMLName   xml.Name    `xml:"rule"`
	Name      string      `xml:"name,attr"`
	Comment   string      `xml:"comment,attr,omitempty"`
	Start     string      `xml:"start,attr,omitempty"`
	End       string      `xml:"end,attr,omitempty"`
	Criteria  []Criterion `xml:"criteria"`
	Behaviors []Behavior  `xml:"behaviors>behavior"`
	Children  []Rule      `xml:"children>rule,omitempty"`
}

// Criterion represents a condition that must be met for a rule to execute
type Criterion struct {
	XMLName xml.Name `xml:"criteria"`
	Name    string   `xml:"name,attr"`
	Option  string   `xml:"option,attr,omitempty"`
	Value   string   `xml:"value,attr,omitempty"`
	Case    bool     `xml:"case,attr,omitempty"`
	Extract string   `xml:"extract,attr,omitempty"`
}

// Behaviors represents a collection of behaviors
type Behaviors struct {
	XMLName  xml.Name   `xml:"behaviors"`
	Behavior []Behavior `xml:"behavior"`
}

// Behavior represents a single behavior with options
type Behavior struct {
	XMLName xml.Name         `xml:"behavior"`
	Name    string           `xml:"name,attr"`
	Option  []BehaviorOption `xml:"option"`
	// For JSON API compatibility
	Options map[string]interface{} `json:"options,omitempty" xml:"-"`
}

// BehaviorOption represents an option for a behavior
type BehaviorOption struct {
	XMLName xml.Name `xml:"option"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
}

// Variables represents a collection of variables
type Variables struct {
	XMLName  xml.Name   `xml:"variables"`
	Variable []Variable `xml:"variable"`
}

// Variable represents a single variable definition
type Variable struct {
	XMLName xml.Name `xml:"variable"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
	Type    string   `xml:"type,attr,omitempty"`
}

// HTTPContext represents the HTTP request/response context for rule processing
type HTTPContext struct {
	Request   *http.Request
	Response  *http.Response
	Headers   map[string]string
	Cookies   map[string]string
	Variables map[string]string
	Path      string
	Method    string
	Host      string
	Query     string
	ClientIP  string
	UserAgent string
	Timestamp time.Time
}

// RuleResult represents the result of rule processing
type RuleResult struct {
	MatchedRules              []string
	ExecutedBehaviors         []string
	ModifiedHeaders           map[string]string
	RemovedHeaders            []string
	ResponseContent           string
	Variables                 map[string]string
	Errors                    []string
	CacheSettings             map[string]interface{}
	CompressionSettings       map[string]interface{}
	ImageOptimizationSettings map[string]interface{}
	RedirectLocation          string
	RedirectStatus            int
	RewrittenURL              string
}

// PropertyManager represents the main property manager emulator
type PropertyManager struct {
	Property  *Property
	Debug     bool
	Rules     map[string]*Rule
	Behaviors map[string]*Behavior
	Variables map[string]string
}

// NewPropertyManager creates a new PropertyManager instance
func NewPropertyManager(debug bool) *PropertyManager {
	return &PropertyManager{
		Debug:     debug,
		Rules:     make(map[string]*Rule),
		Behaviors: make(map[string]*Behavior),
		Variables: make(map[string]string),
	}
}

// LoadProperty loads a property configuration from XML
func (pm *PropertyManager) LoadProperty(xmlData []byte) error {
	var property Property
	if err := xml.Unmarshal(xmlData, &property); err != nil {
		return err
	}

	pm.Property = &property

	// Build rule and behavior maps for quick lookup
	pm.buildRuleMap(&property.Rules)
	pm.buildBehaviorMap(&property.Behaviors)

	// Initialize variables
	for _, v := range property.Variables.Variable {
		pm.Variables[v.Name] = v.Value
	}

	return nil
}

// ProcessRequest processes an HTTP request through the property rules
func (pm *PropertyManager) ProcessRequest(req *http.Request) (*RuleResult, error) {
	context := pm.createHTTPContext(req)
	result := &RuleResult{
		MatchedRules:              []string{},
		ExecutedBehaviors:         []string{},
		ModifiedHeaders:           make(map[string]string),
		RemovedHeaders:            []string{},
		Variables:                 make(map[string]string),
		Errors:                    []string{},
		CacheSettings:             make(map[string]interface{}),
		CompressionSettings:       make(map[string]interface{}),
		ImageOptimizationSettings: make(map[string]interface{}),
	}

	// Process rules
	if err := pm.processRules(pm.Property.Rules.Rule, context, result); err != nil {
		result.Errors = append(result.Errors, err.Error())
	}

	return result, nil
}

// buildRuleMap builds a map of rules for quick lookup
func (pm *PropertyManager) buildRuleMap(rules *Rules) {
	for i := range rules.Rule {
		rule := &rules.Rule[i]
		pm.Rules[rule.Name] = rule
		// Recursively process child rules
		if len(rule.Children) > 0 {
			childRules := Rules{Rule: rule.Children}
			pm.buildRuleMap(&childRules)
		}
	}
}

// buildBehaviorMap builds a map of behaviors for quick lookup
func (pm *PropertyManager) buildBehaviorMap(behaviors *Behaviors) {
	for i := range behaviors.Behavior {
		behavior := &behaviors.Behavior[i]
		pm.Behaviors[behavior.Name] = behavior
	}
}

// createHTTPContext creates an HTTP context from a request
func (pm *PropertyManager) createHTTPContext(req *http.Request) *HTTPContext {
	headers := make(map[string]string)
	for key, values := range req.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	cookies := make(map[string]string)
	for _, cookie := range req.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}

	variables := make(map[string]string)
	for key, value := range pm.Variables {
		variables[key] = value
	}

	return &HTTPContext{
		Request:   req,
		Headers:   headers,
		Cookies:   cookies,
		Variables: variables,
		Path:      req.URL.Path,
		Method:    req.Method,
		Host:      req.Host,
		Query:     req.URL.RawQuery,
		ClientIP:  req.RemoteAddr,
		UserAgent: req.UserAgent(),
		Timestamp: time.Now(),
	}
}

// SetRules sets the rules for the property manager
func (pm *PropertyManager) SetRules(rules []Rule) {
	// Clear existing rules
	pm.Rules = make(map[string]*Rule)

	// Build rule map from the provided rules
	ruleCollection := Rules{Rule: rules}
	pm.buildRuleMap(&ruleCollection)
}

// ProcessHTTPContext processes an HTTP context directly
func (pm *PropertyManager) ProcessHTTPContext(context *HTTPContext) (*RuleResult, error) {
	result := &RuleResult{
		MatchedRules:              []string{},
		ExecutedBehaviors:         []string{},
		ModifiedHeaders:           make(map[string]string),
		RemovedHeaders:            []string{},
		Variables:                 make(map[string]string),
		Errors:                    []string{},
		CacheSettings:             make(map[string]interface{}),
		CompressionSettings:       make(map[string]interface{}),
		ImageOptimizationSettings: make(map[string]interface{}),
	}

	// If we have a property with rules, process them
	if pm.Property != nil && len(pm.Property.Rules.Rule) > 0 {
		if err := pm.processRules(pm.Property.Rules.Rule, context, result); err != nil {
			result.Errors = append(result.Errors, err.Error())
		}
	}

	return result, nil
}
